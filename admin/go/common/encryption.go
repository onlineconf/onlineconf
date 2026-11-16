package common

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// Value encryption at rest.
//
// Secret values are stored as AES-256-GCM envelopes
//
//	oc1:<key-id>:<base64(nonce ‖ ciphertext ‖ tag)>
//
// written and read by onlineconf-admin only: the resolver decrypts at tree
// load, the admin API decrypts on an explicit reveal or a restore. The keys
// come from a key file, one per line, in no particular order: the key new
// encryptions use is the one the admin's registry records as active (a
// fresh installation starts with the file's first key); every other key is
// kept for decryption.

// EncryptionConfig is the `encryption` stanza of the admin configuration.
type EncryptionConfig struct {
	KeyFile string `yaml:"key_file"`
}

// Keys is the keyring loaded at startup; nil when encryption is not
// configured on this instance.
var Keys *Keyring

const (
	envelopePrefix = "oc1"
	keySize        = 32
	nonceSize      = 12
	// verifierPlaintext is sealed under every key, with the key id as AAD, and
	// stored in the key registry so that an instance can prove its key
	// material matches the one registered without touching any stored value.
	verifierPlaintext = "onlineconf-encryption-key-verifier"
)

var keyIDRe = regexp.MustCompile(`^[A-Za-z0-9._-]{1,32}$`)

var (
	ErrNoEncryptionKey = errors.New("value encryption is not configured (no encryption.key_file)")
	ErrInvalidEnvelope = errors.New("invalid encrypted value envelope")
	ErrUnknownKeyID    = errors.New("unknown encryption key id")
	ErrDecryptFailed   = errors.New("decryption failed")
)

// DecryptionError marks a failure that means the loaded key material and the
// stored data disagree (unknown key id, authentication failure, malformed
// envelope under an encrypted marker) — as opposed to transient errors such
// as an unreachable database. The resolver fails closed on it: fatal at
// startup, keep the last good tree on a reload.
type DecryptionError struct {
	KeyID string
	Err   error
}

func (e *DecryptionError) Error() string {
	msg := "cannot decrypt value"
	if e.KeyID != "" {
		msg += " (key " + e.KeyID + ")"
	}
	return msg + ": " + e.Err.Error()
}

func (e *DecryptionError) Unwrap() error { return e.Err }

// IsDecryptionError reports whether err is (or wraps) a DecryptionError.
func IsDecryptionError(err error) bool {
	var de *DecryptionError
	return errors.As(err, &de)
}

// Key is one AES-256-GCM key of the keyring.
type Key struct {
	ID   string
	aead cipher.AEAD
}

func newKey(id string, material []byte) (*Key, error) {
	if !keyIDRe.MatchString(id) {
		return nil, fmt.Errorf("invalid encryption key id %q", id)
	}
	if len(material) != keySize {
		return nil, fmt.Errorf("encryption key %s: expected %d bytes of key material, got %d", id, keySize, len(material))
	}
	block, err := aes.NewCipher(material)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Key{ID: id, aead: aead}, nil
}

// seal returns nonce ‖ ciphertext ‖ tag for the plaintext under the key.
func (k *Key) seal(plaintext, aad []byte) ([]byte, error) {
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return k.aead.Seal(nonce, nonce, plaintext, aad), nil
}

func (k *Key) open(sealed, aad []byte) ([]byte, error) {
	if len(sealed) < nonceSize {
		return nil, ErrInvalidEnvelope
	}
	plaintext, err := k.aead.Open(nil, sealed[:nonceSize], sealed[nonceSize:], aad)
	if err != nil {
		return nil, ErrDecryptFailed
	}
	return plaintext, nil
}

// Verifier seals the registry constant under the key (AAD: the key id). It
// contains no key material.
func (k *Key) Verifier() ([]byte, error) {
	return k.seal([]byte(verifierPlaintext), []byte(k.ID))
}

// CheckVerifier reports whether a registry verifier was produced by this key.
func (k *Key) CheckVerifier(verifier []byte) bool {
	plaintext, err := k.open(verifier, []byte(k.ID))
	return err == nil && string(plaintext) == verifierPlaintext
}

// Keyring holds the keys of the key file, in file order and by id. It has no
// notion of an active key: the admin reads the key in use from its registry
// for every write (First is what a fresh installation starts with).
type Keyring struct {
	keys []*Key
	byID map[string]*Key
}

// ParseKeyring parses the key file format: one `<key-id> <base64 of 32
// random bytes>` per line, `#` comments and blank lines ignored, in file
// order.
func ParseKeyring(r io.Reader) (*Keyring, error) {
	kr := &Keyring{byID: map[string]*Key{}}
	scanner := bufio.NewScanner(r)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		fields := strings.Fields(text)
		if len(fields) != 2 {
			return nil, fmt.Errorf("key file line %d: expected `<key-id> <base64 key>`", line)
		}
		material, err := base64.StdEncoding.DecodeString(fields[1])
		if err != nil {
			return nil, fmt.Errorf("key file line %d: %w", line, err)
		}
		key, err := newKey(fields[0], material)
		if err != nil {
			return nil, fmt.Errorf("key file line %d: %w", line, err)
		}
		if _, dup := kr.byID[key.ID]; dup {
			return nil, fmt.Errorf("key file line %d: duplicate key id %s", line, key.ID)
		}
		kr.keys = append(kr.keys, key)
		kr.byID[key.ID] = key
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(kr.keys) == 0 {
		return nil, errors.New("key file contains no keys")
	}
	return kr, nil
}

// LoadKeyring reads the key file, warning when it is readable by others.
func LoadKeyring(path string) (*Keyring, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if info, err := f.Stat(); err == nil && info.Mode().Perm()&0o077 != 0 {
		log.Warn().Str("file", path).Str("mode", info.Mode().Perm().String()).
			Msg("encryption key file is readable by group or others")
	}
	kr, err := ParseKeyring(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return kr, nil
}

// First returns the key file's first key: what a fresh installation starts
// encrypting with.
func (kr *Keyring) First() *Key { return kr.keys[0] }

// Key returns the key with the given id, if present.
func (kr *Keyring) Key(id string) (*Key, bool) {
	key, ok := kr.byID[id]
	return key, ok
}

// All returns the keys in file order.
func (kr *Keyring) All() []*Key { return kr.keys }

// EncryptWith seals a plaintext under the key with the given id.
func EncryptWith(keyID, plaintext string) (string, error) {
	if Keys == nil {
		return "", ErrNoEncryptionKey
	}
	key, ok := Keys.Key(keyID)
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrUnknownKeyID, keyID)
	}
	return key.encrypt(plaintext)
}

func (k *Key) encrypt(plaintext string) (string, error) {
	header := envelopePrefix + ":" + k.ID
	sealed, err := k.seal([]byte(plaintext), []byte(header))
	if err != nil {
		return "", err
	}
	return header + ":" + base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt opens an envelope with the key it names.
func Decrypt(envelope string) (string, error) {
	if Keys == nil {
		return "", &DecryptionError{KeyID: EnvelopeKeyID(envelope), Err: ErrNoEncryptionKey}
	}
	keyID, sealed, err := ParseEnvelope(envelope)
	if err != nil {
		return "", &DecryptionError{Err: err}
	}
	key, ok := Keys.Key(keyID)
	if !ok {
		return "", &DecryptionError{KeyID: keyID, Err: ErrUnknownKeyID}
	}
	plaintext, err := key.open(sealed, []byte(envelopePrefix+":"+keyID))
	if err != nil {
		return "", &DecryptionError{KeyID: keyID, Err: err}
	}
	return string(plaintext), nil
}

// ParseEnvelope splits an envelope into its key id and sealed payload.
func ParseEnvelope(envelope string) (keyID string, sealed []byte, err error) {
	parts := strings.SplitN(envelope, ":", 3)
	if len(parts) != 3 || parts[0] != envelopePrefix || !keyIDRe.MatchString(parts[1]) {
		return "", nil, ErrInvalidEnvelope
	}
	sealed, err = base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return "", nil, ErrInvalidEnvelope
	}
	return parts[1], sealed, nil
}

// EnvelopeKeyID returns the key id an envelope names, or "" for a non-envelope.
func EnvelopeKeyID(envelope string) string {
	keyID, _, err := ParseEnvelope(envelope)
	if err != nil {
		return ""
	}
	return keyID
}

// EncryptionEligibleTypes are the content types whose values may be stored
// encrypted: the scalars. Null has no value, a symlink's target path drives
// resolution, a case's conditions and branches must stay readable (a branch
// reaches a secret through a symlink to an encrypted parameter), and a
// template's ${/path} references must stay searchable (a secret belongs in
// an encrypted parameter the template references).
var EncryptionEligibleTypes = []string{ContentTypeText, ContentTypeJSON, ContentTypeYAML, ContentTypeList, ContentTypeServer, ContentTypeServer2}

// EncryptionEligible reports whether a value of the content type may be
// stored encrypted.
func EncryptionEligible(contentType string) bool {
	for _, t := range EncryptionEligibleTypes {
		if t == contentType {
			return true
		}
	}
	return false
}

// Reencrypt rewrites an envelope under the target key and reports whether it
// did: an envelope already under the target is returned byte for byte.
func Reencrypt(envelope, target string) (string, bool, error) {
	if Keys == nil {
		return envelope, false, ErrNoEncryptionKey
	}
	if _, ok := Keys.Key(target); !ok {
		return envelope, false, fmt.Errorf("%w: %s", ErrUnknownKeyID, target)
	}
	if EnvelopeKeyID(envelope) == target {
		return envelope, false, nil
	}
	plaintext, err := Decrypt(envelope)
	if err != nil {
		return envelope, false, err
	}
	out, err := EncryptWith(target, plaintext)
	if err != nil {
		return envelope, false, err
	}
	return out, true, nil
}
