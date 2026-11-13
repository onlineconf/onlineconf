//go:build !integration

package admin

import (
	"errors"
	"testing"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

func nullBool(valid, value bool) NullBool {
	var b NullBool
	b.Valid, b.Bool = valid, value
	return b
}

func TestCanReveal(t *testing.T) {
	noAccess, readOnly, readWrite := nullBool(false, false), nullBool(true, false), nullBool(true, true)
	cases := []struct {
		requireWrite bool
		rw           NullBool
		want         bool
	}{
		{true, noAccess, false},
		{true, readOnly, false},
		{true, readWrite, true},
		{false, noAccess, false},
		{false, readOnly, true},
		{false, readWrite, true},
	}
	for _, c := range cases {
		if got := canReveal(c.requireWrite, c.rw); got != c.want {
			t.Errorf("canReveal(requireWrite=%v, rw=%+v) = %v, want %v", c.requireWrite, c.rw.NullBool, got, c.want)
		}
	}
	if revealPolicyName(true) != "rw" || revealPolicyName(false) != "ro" {
		t.Errorf("revealPolicyName wrong")
	}
}

func TestTemplateTexts(t *testing.T) {
	if got := templateTexts("application/x-template", NullStringOf("a ${/x} b")); len(got) != 1 || got[0] != "a ${/x} b" {
		t.Errorf("template text = %v", got)
	}
	if got := templateTexts("text/plain", NullStringOf("${/x}")); got != nil {
		t.Errorf("plain text yielded templates: %v", got)
	}
	nested := `[{"group":"g","mime":"application/x-template","value":"${/a}"},` +
		`{"mime":"application/x-case","value":"[{\"mime\":\"application/x-template\",\"value\":\"${/b}\"},{\"mime\":\"text/plain\",\"value\":\"${/c}\"}]"}]`
	got := templateTexts(ContentTypeCase, NullStringOf(nested))
	if len(got) != 2 || got[0] != "${/a}" || got[1] != "${/b}" {
		t.Errorf("nested case template texts = %v", got)
	}
	if got := templateTexts(ContentTypeCase, NullStringOf("not json")); got != nil {
		t.Errorf("unparsable case yielded templates: %v", got)
	}
	if got := encryptedTemplateVarRe.FindAllStringSubmatch("${/a/b} ${hostname} ${/c}", -1); len(got) != 2 || got[0][1] != "/a/b" || got[1][1] != "/c" {
		t.Errorf("path variables = %v", got)
	}
}

func TestValidateParameterEncryptionEligibility(t *testing.T) {
	enc := ValueFlags{Encrypted: true}
	for _, mime := range []string{"application/x-null", "application/x-symlink", ContentTypeCase} {
		if err := validateParameter(nil, mime, enc, NullStringOf("x")); !errors.Is(err, ErrEncryptionNotAllowed) {
			t.Errorf("%s with the encrypted flag: err = %v, want ErrEncryptionNotAllowed", mime, err)
		}
	}
	for _, mime := range []string{"text/plain", "application/x-list", "application/x-server", "application/x-template"} {
		if err := validateParameter(nil, mime, enc, NullStringOf("x")); err != nil {
			t.Errorf("%s with the encrypted flag: err = %v", mime, err)
		}
	}
	if err := validateParameter(nil, "application/json", enc, NullStringOf(`{"a": 1}`)); err != nil {
		t.Errorf("encrypted json: %v", err)
	}
	if err := validateParameter(nil, "application/json", enc, NullStringOf(`{broken`)); err == nil {
		t.Errorf("broken encrypted json accepted")
	}
	// a case with an encrypted branch whose value is masked (null) is refused
	// by the non-null rule of the branch's own type: the stale-client guard
	masked := `[{"mime":"text/plain","encrypted":true,"value":null}]`
	if err := validateParameter(nil, ContentTypeCase, ValueFlags{}, NullStringOf(masked)); !errors.Is(err, ErrInvalidValue) {
		t.Errorf("masked branch: err = %v, want ErrInvalidValue", err)
	}
	okCase := `[{"server":"a*","mime":"text/plain","encrypted":true,"value":"s"},{"mime":"application/json","value":"1"}]`
	if err := validateParameter(nil, ContentTypeCase, ValueFlags{}, NullStringOf(okCase)); err != nil {
		t.Errorf("valid case with an encrypted branch: %v", err)
	}
	badBranch := `[{"mime":"application/x-symlink","encrypted":true,"value":"/x"}]`
	if err := validateParameter(nil, ContentTypeCase, ValueFlags{}, NullStringOf(badBranch)); !errors.Is(err, ErrEncryptionNotAllowed) {
		t.Errorf("encrypted symlink branch: err = %v", err)
	}
}
