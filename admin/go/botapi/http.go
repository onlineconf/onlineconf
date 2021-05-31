package botapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	. "github.com/onlineconf/onlineconf/admin/go/common"
	"github.com/onlineconf/onlineconf/admin/go/resolver"
)

var writeError = WriteErrorFunc(map[error]ErrorResponse{
	ErrLimitTooLarge: {HTTPCode: 400, ErrorCode: "InvalidParam"},
})
var writeResponseOrError = WriteResponseOrErrorFunc(writeError)

func RegisterRoutes(r *mux.Router) {
	r.Use(authMiddleware)

	n := r.Path("/notification/").Subrouter()
	n.Use(scopeMiddleware("notifications"))
	n.Methods("GET").HandlerFunc(serveNotifications)
}

type NotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
	LastID        int            `json:"lastID"`
}

func serveNotifications(w http.ResponseWriter, req *http.Request) {
	lastID, err := GetQueryInt(req, "lastID")
	if err != nil {
		writeError(req.Context(), w, err)
		return
	}
	limit, err := GetQueryInt(req, "limit")
	if err != nil {
		writeError(req.Context(), w, err)
		return
	}
	waitSeconds, wait, err := GetQueryIntOpt(req, "wait")
	if err != nil {
		writeError(req.Context(), w, err)
		return
	}

	notifications, lastID, err := SelectNotifications(req.Context(), lastID, limit)
	if !wait || err != nil || len(notifications) > 0 || limit == 0 {
		writeResponseOrError(req.Context(), w, NotificationsResponse{
			Notifications: notifications,
			LastID:        lastID,
		}, err)
		return
	}

	waitTimer := time.NewTimer(time.Duration(waitSeconds) * time.Second)
	for {
		timer := time.NewTimer(time.Second)
		select {
		case <-waitTimer.C:
			WriteResponse(req.Context(), w, NotificationsResponse{Notifications: notifications, LastID: lastID})
			return
		case <-timer.C:
			notifications, lastID, err = SelectNotifications(req.Context(), lastID, limit)
			if err != nil || len(notifications) > 0 {
				writeResponseOrError(req.Context(), w, NotificationsResponse{
					Notifications: notifications,
					LastID:        lastID,
				}, err)
				return
			}
		}
	}
}

type botCredentials struct {
	passwordHash []byte
	scopes       map[string]bool
}

type botsCredentials struct {
	m    sync.RWMutex
	bots map[string]botCredentials
}

func (bc *botsCredentials) Update(ctx context.Context, param *resolver.Param) {
	bc.m.Lock()
	defer bc.m.Unlock()
	bc.bots = map[string]botCredentials{}
	if param == nil {
		return
	}
	for name, childPtr := range param.Children {
		child := *childPtr
		if child == nil {
			continue
		}
		pwdHash, err := hex.DecodeString(child.Value.String)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("bot", name).Msg("invalid password hash string")
			continue
		}
		var scopes map[string]bool
		if scopesPtr := child.Children["scopes"]; scopesPtr != nil && *scopesPtr != nil {
			list, err := (*scopesPtr).Strings()
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Str("bot", name).Msg("invalid scopes")
			} else {
				scopes = map[string]bool{}
				for _, scope := range list {
					scopes[scope] = true
				}
			}
		}
		bc.bots[name] = botCredentials{
			passwordHash: pwdHash,
			scopes:       scopes,
		}
	}
}

func (bc *botsCredentials) get(bot string) botCredentials {
	bc.m.RLock()
	defer bc.m.RUnlock()
	return bc.bots[bot]
}

var bots botsCredentials

func init() {
	resolver.Synchronize("/onlineconf/botapi/bot", &bots)
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if username, password, ok := req.BasicAuth(); ok {
			if storedHash := bots.get(username).passwordHash; storedHash != nil {
				receivedHash := sha256.Sum256([]byte(password))
				if bytes.Equal(receivedHash[:], storedHash) {
					req = AddUsernameToRequest(req, username)
					next.ServeHTTP(w, req)
					return
				}
			}
		}
		w.Header().Add("WWW-Authenticate", "Basic realm=onlineconf-botapi")
		w.WriteHeader(401)
	})
}

func scopeMiddleware(scope string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if bots.get(Username(req.Context())).scopes[scope] {
				next.ServeHTTP(w, req)
				return
			}
			WriteResponse(req.Context(), w, ErrorResponse{
				HTTPCode:  403,
				ErrorCode: "ScopeRequired",
				Message:   "Grant to scope \"" + scope + "\" required",
			})
		})
	}
}
