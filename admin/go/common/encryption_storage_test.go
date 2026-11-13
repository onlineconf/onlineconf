//go:build !integration

package common

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEncryptionEligible(t *testing.T) {
	for mime, want := range map[string]bool{
		"text/plain":             true,
		"application/json":       true,
		"application/x-yaml":     true,
		"application/x-template": true,
		"application/x-list":     true,
		"application/x-server":   true,
		"application/x-server2":  true,
		"application/x-null":     false,
		"application/x-symlink":  false,
		"application/x-case":     false,
	} {
		if got := EncryptionEligible(mime); got != want {
			t.Errorf("EncryptionEligible(%s) = %v, want %v", mime, got, want)
		}
	}
}

func TestScalarStorageBoundary(t *testing.T) {
	withKeys(t, testKeyring(t, "k1 "+testKeyA))
	plain := NullStringOf("hunter2")

	stored, err := EncryptStoredValue("text/plain", true, plain)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(stored.String, "oc1:k1:") {
		t.Errorf("stored = %q, want an envelope", stored.String)
	}
	back, err := DecryptStoredValue("text/plain", true, stored)
	if err != nil || back.String != "hunter2" {
		t.Errorf("decrypt = %q, %v", back.String, err)
	}
	if masked := MaskStoredValue("text/plain", true, stored); masked.Valid {
		t.Errorf("masked scalar is not null: %q", masked.String)
	}
	if !HasEncryptedValue("text/plain", true, stored) || HasEncryptedValue("text/plain", false, plain) {
		t.Errorf("HasEncryptedValue wrong for scalars")
	}
	if ids := StoredKeyIDs("text/plain", true, stored); len(ids) != 1 || ids[0] != "k1" {
		t.Errorf("StoredKeyIDs = %v", ids)
	}

	// unencrypted values pass through untouched, null included
	for _, v := range []NullString{plain, {}} {
		if got, err := EncryptStoredValue("text/plain", false, v); err != nil || got != v {
			t.Errorf("plain passthrough changed %v: %v, %v", v, got, err)
		}
		if got, err := DecryptStoredValue("text/plain", false, v); err != nil || got != v {
			t.Errorf("plain passthrough changed %v: %v, %v", v, got, err)
		}
	}
	// a marker without an envelope is a decryption error, never plaintext
	if _, err := DecryptStoredValue("text/plain", true, plain); !IsDecryptionError(err) {
		t.Errorf("non-envelope under the marker: err = %v", err)
	}
}

const nestedCase = `[{"group":"prod","mime":"application/json","encrypted":true,"value":"{\"token\":\"p\"}"},` +
	`{"mime":"application/x-case","value":"[{\"server\":\"a*\",\"mime\":\"text/plain\",\"encrypted\":true,\"value\":\"inner\"},{\"mime\":\"text/plain\",\"value\":\"dev\"}]"},` +
	`{"mime":"application/x-null","value":null}]`

func branchValues(t *testing.T, caseJSON string) []NullString {
	t.Helper()
	var branches []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(caseJSON), &branches); err != nil {
		t.Fatalf("not a case: %v in %s", err, caseJSON)
	}
	values := make([]NullString, len(branches))
	for i, b := range branches {
		if raw, ok := b["value"]; ok {
			_ = json.Unmarshal(raw, &values[i])
		}
	}
	return values
}

func TestCaseStorageBoundary(t *testing.T) {
	withKeys(t, testKeyring(t, "k1 "+testKeyA))
	plain := NullStringOf(nestedCase)

	stored, err := EncryptStoredValue(ContentTypeCase, false, plain)
	if err != nil {
		t.Fatal(err)
	}
	vals := branchValues(t, stored.String)
	if !strings.HasPrefix(vals[0].String, "oc1:k1:") {
		t.Errorf("encrypted branch not enveloped: %q", vals[0].String)
	}
	inner := branchValues(t, vals[1].String)
	if !strings.HasPrefix(inner[0].String, "oc1:k1:") || inner[1].String != "dev" {
		t.Errorf("nested case branches wrong: %+v", inner)
	}
	if strings.Contains(stored.String, "hunter") || strings.Contains(stored.String, `\"token\"`) {
		t.Errorf("plaintext leaked into stored case: %s", stored.String)
	}
	// conditions and markers stay readable
	if !strings.Contains(stored.String, `"group":"prod"`) || !strings.Contains(stored.String, `"encrypted":true`) {
		t.Errorf("conditions or markers lost: %s", stored.String)
	}

	back, err := DecryptStoredValue(ContentTypeCase, false, stored)
	if err != nil {
		t.Fatal(err)
	}
	bv := branchValues(t, back.String)
	if bv[0].String != `{"token":"p"}` {
		t.Errorf("decrypted branch = %q", bv[0].String)
	}
	if ib := branchValues(t, bv[1].String); ib[0].String != "inner" {
		t.Errorf("decrypted nested branch = %q", ib[0].String)
	}

	masked := MaskStoredValue(ContentTypeCase, false, stored)
	mv := branchValues(t, masked.String)
	if mv[0].Valid {
		t.Errorf("masked branch still has a value: %q", mv[0].String)
	}
	if im := branchValues(t, mv[1].String); im[0].Valid || im[1].String != "dev" {
		t.Errorf("masked nested branches wrong: %+v", im)
	}
	if !strings.Contains(masked.String, `"encrypted":true`) {
		t.Errorf("masking dropped the marker: %s", masked.String)
	}
	if !HasEncryptedValue(ContentTypeCase, false, masked) {
		t.Errorf("HasEncryptedValue false on a masked case")
	}

	// a plain case is stored byte for byte
	plainCase := NullStringOf(`[ {"mime": "text/plain", "value": "x"} ]`)
	if got, err := EncryptStoredValue(ContentTypeCase, false, plainCase); err != nil || got.String != plainCase.String {
		t.Errorf("plain case rewritten: %q, %v", got.String, err)
	}
	if HasEncryptedValue(ContentTypeCase, false, plainCase) {
		t.Errorf("HasEncryptedValue true on a plain case")
	}
	if got := MaskStoredValue(ContentTypeCase, false, plainCase); got.String != plainCase.String {
		t.Errorf("plain case masked: %q", got.String)
	}
	// unparsable case mentioning encryption is blanked entirely
	if got := MaskStoredValue(ContentTypeCase, false, NullStringOf(`not json "encrypted":true`)); got.Valid {
		t.Errorf("unparsable case not blanked: %q", got.String)
	}
}

func TestMarkAllEncrypted(t *testing.T) {
	if _, enc, err := MarkAllEncrypted("application/json", NullStringOf("{}")); err != nil || !enc {
		t.Errorf("scalar eligible: enc = %v, err = %v", enc, err)
	}
	if _, enc, _ := MarkAllEncrypted("application/x-symlink", NullStringOf("/x")); enc {
		t.Errorf("symlink marked encrypted")
	}
	if _, enc, _ := MarkAllEncrypted("text/plain", NullString{}); enc {
		t.Errorf("null value marked encrypted")
	}
	caseJSON := `[{"mime":"text/plain","value":"a"},{"mime":"application/x-symlink","value":"/x"},` +
		`{"mime":"application/x-case","value":"[{\"mime\":\"application/x-yaml\",\"value\":\"y: 1\"},{\"mime\":\"application/x-null\",\"value\":null}]"}]`
	out, _, err := MarkAllEncrypted(ContentTypeCase, NullStringOf(caseJSON))
	if err != nil {
		t.Fatal(err)
	}
	var branches []map[string]json.RawMessage
	_ = json.Unmarshal([]byte(out.String), &branches)
	if string(branches[0]["encrypted"]) != "true" {
		t.Errorf("text branch not marked: %s", out.String)
	}
	if _, ok := branches[1]["encrypted"]; ok {
		t.Errorf("symlink branch marked: %s", out.String)
	}
	var nested string
	_ = json.Unmarshal(branches[2]["value"], &nested)
	var nb []map[string]json.RawMessage
	_ = json.Unmarshal([]byte(nested), &nb)
	if string(nb[0]["encrypted"]) != "true" {
		t.Errorf("nested yaml branch not marked: %s", nested)
	}
	if _, ok := nb[1]["encrypted"]; ok {
		t.Errorf("nested null branch marked: %s", nested)
	}
}

func TestReencryptStoredValue(t *testing.T) {
	withKeys(t, testKeyring(t, "old "+testKeyB))
	stored, err := EncryptStoredValue(ContentTypeCase, false, NullStringOf(nestedCase))
	if err != nil {
		t.Fatal(err)
	}
	scalar, _ := EncryptStoredValue("text/plain", true, NullStringOf("s"))

	// rotate: new active key, old retired
	withKeys(t, testKeyring(t, "new "+testKeyA+"\nold "+testKeyB))
	out, changed, err := ReencryptStoredValue(ContentTypeCase, false, stored, "new")
	if err != nil || !changed {
		t.Fatalf("reencrypt: changed = %v, err = %v", changed, err)
	}
	if ids := StoredKeyIDs(ContentTypeCase, false, out); len(ids) != 1 || ids[0] != "new" {
		t.Errorf("key ids after reencrypt = %v", ids)
	}
	back, err := DecryptStoredValue(ContentTypeCase, false, out)
	if err != nil || branchValues(t, back.String)[0].String != `{"token":"p"}` {
		t.Errorf("reencrypted case does not decrypt: %v", err)
	}
	// idempotent
	if _, changed, err := ReencryptStoredValue(ContentTypeCase, false, out, "new"); err != nil || changed {
		t.Errorf("second reencrypt changed = %v, err = %v", changed, err)
	}
	if out, changed, err := ReencryptStoredValue("text/plain", true, scalar, "new"); err != nil || !changed || EnvelopeKeyID(out.String) != "new" {
		t.Errorf("scalar reencrypt: %q, %v, %v", out.String, changed, err)
	}
	// a plaintext value is untouched
	if _, changed, err := ReencryptStoredValue("text/plain", false, NullStringOf("p"), "new"); err != nil || changed {
		t.Errorf("plaintext reencrypt changed = %v, err = %v", changed, err)
	}
}
