package common

import (
	"context"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/url"
	"strconv"
)

type usernameKey struct{}

func Username(ctx context.Context) string {
	return ctx.Value(usernameKey{}).(string)
}

type AuthenticateHandler struct {
	next http.Handler
}

func Authenticate(h http.Handler) http.Handler {
	return &AuthenticateHandler{h}
}

func (h *AuthenticateHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if username, password, ok := req.BasicAuth(); ok {
		ok, err := checkUserPassword(req.Context(), username, password)
		if err != nil {
			WriteServerError(req.Context(), w, err)
			return
		}
		if ok {
			req = req.WithContext(context.WithValue(req.Context(), usernameKey{}, username))
			req.URL.User = url.User(username)
			h.next.ServeHTTP(w, req)
			return
		}
	}
	w.Header().Add("WWW-Authenticate", "Basic realm="+url.PathEscape(AdminConfig.Auth.Realm))
	w.WriteHeader(401)
}

func WriteServerError(ctx context.Context, w http.ResponseWriter, err error) {
	log.Ctx(ctx).Error().Err(err).Msg("500")
	header := w.Header()
	content, merr := json.Marshal(map[string]string{
		"status":  "Error",
		"message": err.Error(),
	})
	if merr == nil {
		header.Add("Content-Type", "application/json")
	} else {
		log.Ctx(ctx).Warn().Err(merr).Msg("failed to marshal json")
		content = []byte(err.Error())
	}
	header.Add("Content-Length", strconv.Itoa(len(content)))
	w.WriteHeader(500)
	w.Write(content)
}
