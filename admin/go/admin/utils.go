package admin

import (
	"regexp"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

type AdminConfig struct {
	Auth                 AuthenticatorConfig
	NotificationDatabase DatabaseConfig `yaml:"notification_database"`
}

var adminConfig *AdminConfig

func Initialize(config AdminConfig) {
	adminConfig = &config
	initAuthenticator(config.Auth)
	notifyDB = OpenDatabase(config.NotificationDatabase)
}

var pathRe = regexp.MustCompile(`^(.*)/([^/]+)$`)

func splitPath(path string) (string, string) {
	m := pathRe.FindStringSubmatch(path)
	parentPath, name := m[1], m[2]
	if parentPath == "" {
		parentPath = "/"
	}
	return parentPath, name
}

var likeRe = regexp.MustCompile("([%_])")

func likeEscape(str string) string {
	return likeRe.ReplaceAllString(str, `\$1`)
}
