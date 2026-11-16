//go:build !integration

package admin

import (
	"context"
	"errors"
	"net/url"
	"testing"

	. "github.com/onlineconf/onlineconf/admin/go/common"
	"github.com/onlineconf/onlineconf/admin/go/resolver"
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
}

func TestValidateParameterEncryptionEligibility(t *testing.T) {
	for _, mime := range []string{"application/x-null", "application/x-symlink", ContentTypeCase, ContentTypeTemplate} {
		if err := validateParameter(nil, mime, true, NullStringOf("x")); !errors.Is(err, ErrEncryptionNotAllowed) {
			t.Errorf("%s with the encrypted flag: err = %v, want ErrEncryptionNotAllowed", mime, err)
		}
	}
	for _, mime := range []string{"text/plain", "application/x-list", "application/x-server"} {
		if err := validateParameter(nil, mime, true, NullStringOf("x")); err != nil {
			t.Errorf("%s with the encrypted flag: err = %v", mime, err)
		}
	}
	if err := validateParameter(nil, ContentTypeTemplate, false, NullStringOf("${/x}")); err != nil {
		t.Errorf("plaintext template: %v", err)
	}
	if err := validateParameter(nil, "application/json", true, NullStringOf(`{"a": 1}`)); err != nil {
		t.Errorf("encrypted json: %v", err)
	}
	if err := validateParameter(nil, "application/json", true, NullStringOf(`{broken`)); err == nil {
		t.Errorf("broken encrypted json accepted")
	}
	// a case branch is never encrypted: a branch carrying the marker is
	// refused, the same case without it is accepted
	marked := `[{"server":"a*","mime":"text/plain","encrypted":true,"value":"s"},{"mime":"application/json","value":"1"}]`
	if err := validateParameter(nil, ContentTypeCase, false, NullStringOf(marked)); !errors.Is(err, ErrEncryptionNotAllowed) {
		t.Errorf("marked branch: err = %v, want ErrEncryptionNotAllowed", err)
	}
	plain := `[{"server":"a*","mime":"text/plain","value":"s"},{"mime":"application/json","value":"1"}]`
	if err := validateParameter(nil, ContentTypeCase, false, NullStringOf(plain)); err != nil {
		t.Errorf("plain case: %v", err)
	}
	// a branch without a value is refused by the non-null rule of its type
	if err := validateParameter(nil, ContentTypeCase, false, NullStringOf(`[{"mime":"text/plain","value":null}]`)); !errors.Is(err, ErrInvalidValue) {
		t.Errorf("null text branch: err = %v, want ErrInvalidValue", err)
	}
}

func TestEncryptedField(t *testing.T) {
	for _, c := range []struct {
		form  url.Values
		valid bool
		value bool
		err   error
	}{
		{url.Values{}, false, false, nil},
		{url.Values{"encrypted": {"0"}}, true, false, nil},
		{url.Values{"encrypted": {"1"}}, true, true, nil},
		{url.Values{"encrypted": {"true"}}, false, false, errInvalidEncryptedField},
		{url.Values{"encrypted": {"on"}}, false, false, errInvalidEncryptedField},
		{url.Values{"encrypted": {""}}, false, false, errInvalidEncryptedField},
	} {
		got, err := encryptedField(c.form)
		if !errors.Is(err, c.err) || (err == nil && (got.Valid != c.valid || got.Bool != c.value)) {
			t.Errorf("encryptedField(%v) = %+v, %v; want valid=%v value=%v err=%v", c.form, got.NullBool, err, c.valid, c.value, c.err)
		}
	}
}

// withKeys installs a keyring for the test's duration.
func withKeys(t *testing.T, kr *Keyring) {
	t.Helper()
	prev := Keys
	Keys = kr
	t.Cleanup(func() { Keys = prev })
}

// Without a keyring the feature is off whatever the switch parameter says:
// an encrypted write is refused before the database is consulted.
func TestEncryptForStorageGate(t *testing.T) {
	withKeys(t, nil)
	if _, _, err := encryptForStorage(context.Background(), nil, true, NullStringOf("s")); !errors.Is(err, ErrEncryptionDisabled) {
		t.Errorf("encrypted write without a keyring: err = %v, want ErrEncryptionDisabled", err)
	}
	// a plaintext value passes through untouched, null included
	for _, v := range []NullString{NullStringOf("p"), {}} {
		if got, _, err := encryptForStorage(context.Background(), nil, false, v); err != nil || got != v {
			t.Errorf("plaintext passthrough of %v: %v, %v", v, got, err)
		}
	}
}

func TestPlanActiveKey(t *testing.T) {
	for _, c := range []struct {
		name            string
		wanted, active  string
		hasActive, held bool
		want            keyAction
	}{
		{"fresh installation", "k1", "", false, true, keyActionRegister},
		{"fresh installation, key not held", "k1", "", false, false, keyActionFatal},
		{"the key in use", "k1", "k1", true, true, keyActionNone},
		{"a key not held", "k2", "k1", true, false, keyActionRefuse},
		{"a held key", "k2", "k1", true, true, keyActionPromote},
		{"a retired key again", "k0", "k1", true, true, keyActionPromote},
	} {
		if got, _ := planActiveKey(c.wanted, "parameter", c.active, c.hasActive, c.held); got != c.want {
			t.Errorf("%s: action = %v, want %v", c.name, got, c.want)
		}
	}
}

// The watcher compares what the reader of the key id parameter would see:
// a plain text, anything else absent. (Without a keyring nothing is
// triggered, which keeps the test off the database.)
func TestKeyIDWatcherValue(t *testing.T) {
	withKeys(t, nil)
	node := func(contentType string, encrypted bool, value string) *resolver.Param {
		return &resolver.Param{ContentType: contentType, Encrypted: encrypted, Value: NullStringOf(value)}
	}
	w := &keyIDWatcher{}
	for _, c := range []struct {
		name  string
		param *resolver.Param
		want  string
	}{
		{"text", node(ContentTypeText, false, "k1"), "k1"},
		{"absent", nil, ""},
		{"encrypted text", node(ContentTypeText, true, "k1"), ""},
		{"a case resolved to absent", node(ContentTypeNull, false, ""), ""},
		{"json", node(ContentTypeJSON, false, `"k1"`), ""},
		{"text again", node(ContentTypeText, false, "k2"), "k2"},
	} {
		w.Update(context.Background(), c.param)
		if !w.seen || w.last != c.want {
			t.Errorf("%s: last = %q, want %q", c.name, w.last, c.want)
		}
	}
	w.prime("k3")
	if !w.seen || w.last != "k3" {
		t.Errorf("primed: last = %q", w.last)
	}
}
