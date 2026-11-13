//go:build !integration

package admin

import (
	"context"
	"testing"

	. "github.com/onlineconf/onlineconf/admin/go/common"
	"github.com/onlineconf/onlineconf/admin/go/resolver"
)

func TestWatchedControl(t *testing.T) {
	w := &watchedControl{param: revealPolicy}
	ctx := context.Background()
	node := func(contentType string, encrypted bool, value string) *resolver.Param {
		return &resolver.Param{ContentType: contentType, Encrypted: encrypted, Value: NullStringOf(value)}
	}
	cases := []struct {
		name  string
		param *resolver.Param
		value string
		ok    bool
	}{
		{"text", node(ContentTypeText, false, "ro"), "ro", true},
		{"absent", nil, "", false},
		{"encrypted text", node(ContentTypeText, true, "ro"), "", false},
		{"json", node(ContentTypeJSON, false, "\"ro\""), "", false},
		{"text again", node(ContentTypeText, false, "rw"), "rw", true},
	}
	for _, c := range cases {
		w.Update(ctx, c.param)
		value, ok, err := w.text(ctx)
		if err != nil || value != c.value || ok != c.ok {
			t.Errorf("%s: text = %q, %v, %v; want %q, %v", c.name, value, ok, err, c.value, c.ok)
		}
	}
}

func TestTruthy(t *testing.T) {
	for value, want := range map[string]bool{"": false, "0": false, "1": true, "yes": true, " 0": true, "00": true} {
		if got := truthy(value); got != want {
			t.Errorf("truthy(%q) = %v, want %v", value, got, want)
		}
	}
}
