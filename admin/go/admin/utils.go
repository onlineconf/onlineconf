package admin

import (
	"regexp"
)

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
