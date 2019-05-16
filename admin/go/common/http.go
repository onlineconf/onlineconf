package common

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/rs/zerolog/log"
)

type usernameKey struct{}

func Username(ctx context.Context) string {
	return ctx.Value(usernameKey{}).(string)
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

func AddUsernameToRequest(req *http.Request, username string) *http.Request {
	req = req.WithContext(context.WithValue(req.Context(), usernameKey{}, username))
	req.URL.User = url.User(username)
	return req
}

func SetStatusUnauthorized(w http.ResponseWriter) {
	w.Header().Add("WWW-Authenticate", "Basic realm="+url.PathEscape(AdminConfig.Auth.Realm))
	w.WriteHeader(401)
}
