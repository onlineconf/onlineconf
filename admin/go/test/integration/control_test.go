//go:build integration

package integration

import (
	"net/http"
	"net/url"
	"testing"
)

// TestControlParameterSymlinks: a control parameter may be a symlink to a
// plain text parameter; a symlink to an encrypted one, and a symlink cycle,
// count as absent (the switch is off).
func TestControlParameterSymlinks(t *testing.T) {
	sw := "/onlineconf/value-encryption"
	t.Cleanup(func() { setParam(t, sw, url.Values{"mime": {"text/plain"}, "data": {"1"}}) })
	target := uniqPath("ctl-target")
	mkParam(t, target, "text/plain", "1", "mk")
	secret := uniqPath("ctl-secret")
	mkEncrypted(t, secret, "text/plain", "1", "mk")

	encryptedWrite := func(name string) (int, []byte) {
		return adminReq(t, "POST", "/config"+uniqPath(name), url.Values{"mime": {"text/plain"}, "data": {"x"}, "encrypted": {"1"}, "comment": {"mk"}})
	}
	if status, body := setParam(t, sw, url.Values{"mime": {"application/x-symlink"}, "data": {target}}); status != http.StatusOK {
		t.Fatalf("switch as a symlink: status %d, body %s", status, body)
	}
	if status, body := encryptedWrite("ctl-enc1"); status != http.StatusOK {
		t.Errorf("encrypted write with the switch symlinked to a text: status %d, body %s", status, body)
	}
	if status, body := setParam(t, sw, url.Values{"mime": {"application/x-symlink"}, "data": {secret}}); status != http.StatusOK {
		t.Fatalf("switch as a symlink to a secret: status %d, body %s", status, body)
	}
	status, body := encryptedWrite("ctl-enc2")
	expectError(t, status, body, http.StatusBadRequest, "EncryptionDisabled")
	if status, body := setParam(t, sw, url.Values{"mime": {"application/x-symlink"}, "data": {sw}}); status != http.StatusOK {
		t.Fatalf("switch as a symlink to itself: status %d, body %s", status, body)
	}
	status, body = encryptedWrite("ctl-enc3")
	expectError(t, status, body, http.StatusBadRequest, "EncryptionDisabled")
}
