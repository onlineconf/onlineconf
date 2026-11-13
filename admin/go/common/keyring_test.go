//go:build !integration

package common

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

func TestKeyringActive(t *testing.T) {
	material := func(b byte) string { return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{b}, 32)) }
	kr, err := ParseKeyring(strings.NewReader("first " + material(1) + "\nsecond " + material(2) + "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if kr.Active().ID != "first" || kr.First().ID != "first" {
		t.Errorf("a fresh keyring encrypts with %s, want the first key", kr.Active().ID)
	}
	if !kr.SetActive("second") || kr.Active().ID != "second" || kr.First().ID != "first" {
		t.Errorf("SetActive(second): active %s, first %s", kr.Active().ID, kr.First().ID)
	}
	if kr.SetActive("third") || kr.Active().ID != "second" {
		t.Errorf("SetActive of an unknown key changed the active key to %s", kr.Active().ID)
	}
}
