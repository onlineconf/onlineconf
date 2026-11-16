//go:build !integration

package admin

import (
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

func TestRevealQuery(t *testing.T) {
	for _, c := range []struct {
		query string
		want  bool
		bad   bool
	}{
		{"", false, false},
		{"?reveal=1", true, false},
		{"?reveal=yes", false, true},
		{"?reveal=0", false, true},
	} {
		got, err := revealQuery(httptest.NewRequest("GET", "/config/a"+c.query, nil))
		response, isResponse := err.(ErrorResponse)
		if c.bad != (err != nil) || got != c.want || (c.bad && (!isResponse || response.HTTPCode != 400 || response.ErrorCode != "InvalidParam")) {
			t.Errorf("revealQuery(%q) = %v, %v", c.query, got, err)
		}
	}
}

func TestFormInt(t *testing.T) {
	for _, c := range []struct {
		value string
		want  int
		bad   bool
	}{
		{"12", 12, false},
		{"abc", 0, true},
		{"", 0, true},
	} {
		got, err := formInt(url.Values{"version": {c.value}}, "version")
		response, isResponse := err.(ErrorResponse)
		if c.bad != (err != nil) || got != c.want || (c.bad && (!isResponse || response.HTTPCode != 400 || response.ErrorCode != "InvalidParam")) {
			t.Errorf("formInt(%q) = %v, %v", c.value, got, err)
		}
	}
	if _, err := formInt(url.Values{}, "version"); err == nil {
		t.Errorf("an absent field parsed")
	}
}
