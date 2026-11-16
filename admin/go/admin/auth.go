package admin

import (
	"context"
	"net/http"
	"net/url"

	. "github.com/onlineconf/onlineconf/admin/go/common"
	"github.com/rs/zerolog/log"
)

type AuthenticatorConfig struct {
	Method     string
	MySQLAuth  MySQLAuthenticatorConfig  `yaml:",inline"`
	HeaderAuth HeaderAuthenticatorConfig `yaml:",inline"`
	Realm      string
}

type Authenticator interface {
	Authenticate(req *http.Request) (string, error)
	SelectUsers(ctx context.Context, term string) ([]string, error)
}

var authenticator Authenticator

func initAuthenticator(config AuthenticatorConfig) {
	switch config.Method {
	case "mysql", "":
		authenticator = NewMySQLAuthenticator(config.MySQLAuth)
	case "header":
		authenticator = NewHeaderAuthenticator(config.HeaderAuth)
	default:
		log.Fatal().Str("method", config.Method).Msg("unknown authentication method")
	}
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		username, err := authenticator.Authenticate(req)
		if err != nil {
			WriteServerError(req.Context(), w, err)
			return
		}
		if username == "" {
			w.Header().Add("WWW-Authenticate", "Basic realm="+url.PathEscape(adminConfig.Auth.Realm))
			w.WriteHeader(401)
			return
		}
		req = AddUsernameToRequest(req, username)
		next.ServeHTTP(w, req)
	})
}
