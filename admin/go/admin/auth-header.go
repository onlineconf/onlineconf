package admin

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"strings"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

type HeaderAuthenticatorConfig struct {
	Header string
	Secret string
}

type headerAuthenticator struct {
	config HeaderAuthenticatorConfig
}

func NewHeaderAuthenticator(config HeaderAuthenticatorConfig) *headerAuthenticator {
	if config.Header == "" {
		config.Header = "X-Username"
	}
	return &headerAuthenticator{
		config: config,
	}
}

func (auth *headerAuthenticator) Authenticate(req *http.Request) (string, error) {
	username := req.Header.Get(auth.config.Header)
	if auth.config.Secret == "" {
		return username, nil
	}

	signHeader := req.Header.Get(auth.config.Header + "-Sign")
	if signHeader == "" {
		return "", nil
	}
	sign, err := hex.DecodeString(signHeader)
	if err != nil {
		return "", nil
	}

	remoteIP := strings.Split(req.RemoteAddr, ":")[0]

	h := md5.New()
	h.Write([]byte(username))
	h.Write([]byte(remoteIP))
	h.Write([]byte(auth.config.Secret))
	if bytes.Equal(sign, h.Sum(nil)) {
		return username, nil
	}
	return "", nil
}

func (auth *headerAuthenticator) SelectUsers(ctx context.Context, term string) ([]string, error) {
	order := "1"
	bind := make([]interface{}, 0)
	query := "SELECT DISTINCT User FROM my_config_user_group"
	if term != "" {
		query += " WHERE User LIKE ?"
		bind = append(bind, "%"+termRe.ReplaceAllString(term, "\\$1")+"%", term)
		order = "INSTR(User, ?), 1"
	}
	query += " ORDER BY " + order
	rows, err := DB.QueryContext(ctx, query, bind...)
	if err != nil {
		return nil, err
	}
	return ReadStrings(rows)
}
