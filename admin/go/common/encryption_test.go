//go:build !integration

package common

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"testing"
)

const (
	testKeyA = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=" // 32 zero bytes
	testKeyB = "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE=" // 32 bytes of 0x01
)

func testKeyring(t *testing.T, text string) *Keyring {
	t.Helper()
	kr, err := ParseKeyring(strings.NewReader(text))
	if err != nil {
		t.Fatalf("ParseKeyring: %v", err)
	}
	return kr
}

func withKeys(t *testing.T, kr *Keyring) {
	t.Helper()
	prev := Keys
	Keys = kr
	t.Cleanup(func() { Keys = prev })
}

func TestParseKeyring(t *testing.T) {
	kr := testKeyring(t, "# comment\n\n2026-11 "+testKeyA+"\n  2025-04   "+testKeyB+"  \n")
	if kr.Active().ID != "2026-11" {
		t.Errorf("active key = %q, want the first line", kr.Active().ID)
	}
	if _, ok := kr.Key("2025-04"); !ok {
		t.Errorf("retired key not found")
	}
	if len(kr.All()) != 2 {
		t.Errorf("keys = %d, want 2", len(kr.All()))
	}

	for name, text := range map[string]string{
		"empty":       "# nothing\n",
		"short key":   "k " + base64.StdEncoding.EncodeToString([]byte("short")),
		"bad base64":  "k not-base64!",
		"bad id":      "bad/id " + testKeyA,
		"duplicate":   "k " + testKeyA + "\nk " + testKeyB,
		"extra field": "k " + testKeyA + " extra",
		"missing key": "k",
		"too long id": strings.Repeat("x", 33) + " " + testKeyA,
	} {
		if _, err := ParseKeyring(strings.NewReader(text)); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestEnvelopeRoundTrip(t *testing.T) {
	withKeys(t, testKeyring(t, "k1 "+testKeyA+"\nk0 "+testKeyB))

	for _, plaintext := range []string{"", "secret", "multi\nline\n", strings.Repeat("x", 10000), "ünïcödé"} {
		envelope, err := Encrypt(plaintext)
		if err != nil {
			t.Fatalf("Encrypt: %v", err)
		}
		if !strings.HasPrefix(envelope, "oc1:k1:") {
			t.Errorf("envelope %q does not name the active key", envelope)
		}
		if EnvelopeKeyID(envelope) != "k1" {
			t.Errorf("EnvelopeKeyID = %q", EnvelopeKeyID(envelope))
		}
		got, err := Decrypt(envelope)
		if err != nil || got != plaintext {
			t.Errorf("Decrypt(%q) = %q, %v; want %q", envelope, got, err, plaintext)
		}
	}

	// two encryptions of the same plaintext differ (fresh nonce)
	a, _ := Encrypt("same")
	b, _ := Encrypt("same")
	if a == b {
		t.Errorf("identical envelopes for the same plaintext")
	}

	// a retired key decrypts, the active one is used for new envelopes
	retired, _ := Keys.Key("k0")
	envelope, err := retired.encrypt("old")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := Decrypt(envelope); err != nil || got != "old" {
		t.Errorf("retired key decrypt = %q, %v", got, err)
	}
}

func TestDecryptFailures(t *testing.T) {
	withKeys(t, testKeyring(t, "k1 "+testKeyA))
	envelope, _ := Encrypt("secret")

	// tampered payload
	tampered := envelope[:len(envelope)-5] + "AAAA="
	if _, err := Decrypt(tampered); !IsDecryptionError(err) || !errors.Is(err, ErrDecryptFailed) && !errors.Is(err, ErrInvalidEnvelope) {
		t.Errorf("tampered envelope: err = %v", err)
	}
	// unknown key id
	unknown := strings.Replace(envelope, "oc1:k1:", "oc1:k9:", 1)
	if _, err := Decrypt(unknown); !errors.Is(err, ErrUnknownKeyID) || !IsDecryptionError(err) {
		t.Errorf("unknown key: err = %v", err)
	}
	// not an envelope at all
	if _, err := Decrypt("plaintext that claims to be encrypted"); !errors.Is(err, ErrInvalidEnvelope) {
		t.Errorf("non-envelope: err = %v", err)
	}
	// re-labelled header fails authentication (AAD)
	withKeys(t, testKeyring(t, "k2 "+testKeyA+"\nk1 "+testKeyA))
	relabelled := strings.Replace(envelope, "oc1:k1:", "oc1:k2:", 1)
	if _, err := Decrypt(relabelled); !errors.Is(err, ErrDecryptFailed) {
		t.Errorf("relabelled envelope: err = %v", err)
	}
	// no keyring at all
	withKeys(t, nil)
	if _, err := Decrypt(envelope); !errors.Is(err, ErrNoEncryptionKey) || !IsDecryptionError(err) {
		t.Errorf("no keyring: err = %v", err)
	}
	if _, err := Encrypt("x"); !errors.Is(err, ErrNoEncryptionKey) {
		t.Errorf("Encrypt without keyring: err = %v", err)
	}
}

func TestVerifier(t *testing.T) {
	kr := testKeyring(t, "k1 "+testKeyA+"\nk2 "+testKeyB)
	k1, _ := kr.Key("k1")
	k2, _ := kr.Key("k2")
	v, err := k1.Verifier()
	if err != nil {
		t.Fatal(err)
	}
	if !k1.CheckVerifier(v) {
		t.Errorf("a key does not verify its own verifier")
	}
	if k2.CheckVerifier(v) {
		t.Errorf("a different key verifies k1's verifier")
	}
	// the same material under another id does not verify (AAD binds the id)
	same, _ := newKey("other", make([]byte, 32))
	if same.CheckVerifier(v) {
		t.Errorf("same material under another id verifies")
	}
	if len(v) > 64 {
		t.Errorf("verifier is %d bytes, must fit varbinary(64)", len(v))
	}
}

func TestInitializeKeyringSources(t *testing.T) {
	file := t.TempDir() + "/keys"
	if err := os.WriteFile(file, []byte("env-key "+testKeyA+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	withKeys(t, nil)

	// nothing configured: encryption stays off, loudly but without failing
	t.Setenv(EncryptionKeyFileEnv, "")
	initializeKeyring(EncryptionConfig{})
	if Keys != nil {
		t.Fatalf("keyring loaded from nowhere")
	}
	// the environment names the file when the configuration does not
	t.Setenv(EncryptionKeyFileEnv, file)
	initializeKeyring(EncryptionConfig{})
	if Keys == nil || Keys.Active().ID != "env-key" {
		t.Fatalf("keyring not loaded from %s", EncryptionKeyFileEnv)
	}
	// the configuration wins over the environment
	other := t.TempDir() + "/keys"
	if err := os.WriteFile(other, []byte("cfg-key "+testKeyB+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	initializeKeyring(EncryptionConfig{KeyFile: other})
	if Keys == nil || Keys.Active().ID != "cfg-key" {
		t.Fatalf("keyring not loaded from the configuration")
	}
}
