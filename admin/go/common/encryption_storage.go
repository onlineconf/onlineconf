package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// The storage boundary of value encryption: every reader and writer of
// my_config_tree.Value / my_config_tree_log.Value goes through these
// functions, which know how a value is laid out (a scalar, or a case whose
// branches — nested cases included — carry their own "encrypted" marker) and
// apply one transformation per encrypted leaf.

// EncryptionEligible reports whether a value of the content type may be
// stored encrypted: everything except null (no value), symlink (the target
// path drives resolution) and case (branches carry their own marker).
func EncryptionEligible(contentType string) bool {
	switch contentType {
	case ContentTypeNull, ContentTypeSymlink, ContentTypeCase:
		return false
	}
	return true
}

var errNotCase = errors.New("not a case value")

type caseBranch map[string]json.RawMessage

func (b caseBranch) str(key string) NullString {
	var s NullString
	if raw, ok := b[key]; ok {
		_ = json.Unmarshal(raw, &s) // a non-string stays invalid, as absent
	}
	return s
}

func (b caseBranch) bool(key string) bool {
	var v bool
	if raw, ok := b[key]; ok {
		_ = json.Unmarshal(raw, &v)
	}
	return v
}

func (b caseBranch) set(key string, v interface{}) {
	raw, _ := json.Marshal(v)
	b[key] = raw
}

// leafVisitor is applied to every scalar value: the node value itself, or a
// case branch value (nested cases are descended into). It returns the new
// value and whether it changed it.
type leafVisitor func(encrypted bool, value NullString) (NullString, bool, error)

// branchVisitor may additionally rewrite a branch's own markers; nil leaves
// them alone.
type branchVisitor func(b caseBranch, mime string, value NullString) bool

// walkValue applies the visitors to a stored value of the given content type
// and returns the (possibly) rewritten value. A case whose branches were not
// touched is returned byte for byte.
func walkValue(contentType string, encrypted bool, value NullString, leaf leafVisitor, branch branchVisitor) (NullString, bool, error) {
	if contentType != ContentTypeCase {
		return leaf(encrypted, value)
	}
	if !value.Valid {
		return value, false, nil
	}
	rewritten, changed, err := walkCase(value.String, leaf, branch)
	if err != nil || !changed {
		return value, false, err
	}
	return NullStringOf(rewritten), true, nil
}

func walkCase(caseJSON string, leaf leafVisitor, branch branchVisitor) (string, bool, error) {
	var branches []caseBranch
	if err := json.Unmarshal([]byte(caseJSON), &branches); err != nil {
		return "", false, errNotCase
	}
	changed := false
	for _, b := range branches {
		mime := b.str("mime").String
		value := b.str("value")
		if branch != nil && branch(b, mime, value) {
			changed = true
		}
		var newValue NullString
		var valueChanged bool
		var err error
		if mime == ContentTypeCase {
			if !value.Valid {
				continue
			}
			var nested string
			nested, valueChanged, err = walkCase(value.String, leaf, branch)
			newValue = NullStringOf(nested)
		} else {
			newValue, valueChanged, err = leaf(b.bool("encrypted"), value)
		}
		if err != nil {
			return "", false, err
		}
		if valueChanged {
			if newValue.Valid {
				b.set("value", newValue.String)
			} else {
				b.set("value", nil)
			}
			changed = true
		}
	}
	if !changed {
		return caseJSON, false, nil
	}
	out, err := json.Marshal(branches)
	if err != nil {
		return "", false, err
	}
	return string(out), true, nil
}

// EncryptStoredValue turns a plaintext value as posted by a client into its
// stored form: every encrypted leaf becomes an envelope under the active key.
func EncryptStoredValue(contentType string, encrypted bool, value NullString) (NullString, error) {
	out, _, err := walkValue(contentType, encrypted, value, func(encrypted bool, v NullString) (NullString, bool, error) {
		if !encrypted || !v.Valid {
			return v, false, nil
		}
		envelope, err := Encrypt(v.String)
		if err != nil {
			return v, false, err
		}
		return NullStringOf(envelope), true, nil
	}, nil)
	return out, err
}

// DecryptStoredValue turns a stored value into plaintext: every encrypted
// leaf is opened with the key its envelope names. A leaf that is marked
// encrypted but is not an envelope is a DecryptionError, never plaintext.
func DecryptStoredValue(contentType string, encrypted bool, value NullString) (NullString, error) {
	out, _, err := walkValue(contentType, encrypted, value, func(encrypted bool, v NullString) (NullString, bool, error) {
		if !encrypted || !v.Valid {
			return v, false, nil
		}
		plaintext, err := Decrypt(v.String)
		if err != nil {
			return v, false, err
		}
		return NullStringOf(plaintext), true, nil
	}, nil)
	return out, err
}

// MaskStoredValue blanks every encrypted leaf (null value, marker kept)
// without decrypting anything — the shape every listing serves. A case whose
// JSON cannot be parsed is blanked entirely when it mentions encryption: fail
// closed.
func MaskStoredValue(contentType string, encrypted bool, value NullString) NullString {
	out, _, err := walkValue(contentType, encrypted, value, func(encrypted bool, v NullString) (NullString, bool, error) {
		if !encrypted || !v.Valid {
			return v, false, nil
		}
		return NullString{}, true, nil
	}, nil)
	if err != nil {
		if value.Valid && strings.Contains(value.String, "encrypted") {
			return NullString{}
		}
		return value
	}
	return out
}

// HasEncryptedValue reports whether the value holds an encrypted leaf at any
// depth: the node-level marker, or any branch of a (nested) case.
func HasEncryptedValue(contentType string, encrypted bool, value NullString) bool {
	found := false
	_, _, _ = walkValue(contentType, encrypted, value, func(encrypted bool, v NullString) (NullString, bool, error) {
		if encrypted {
			found = true
		}
		return v, false, nil
	}, nil)
	return found
}

// MarkAllEncrypted returns the value's markers set on every leaf that can
// carry one: the node-level marker for a scalar (reported as the returned
// flag), "encrypted": true on every eligible non-null branch of a (nested)
// case. Used where encryption must never downgrade (restore) and when sealing
// history rows.
func MarkAllEncrypted(contentType string, value NullString) (NullString, bool, error) {
	if contentType != ContentTypeCase {
		return value, EncryptionEligible(contentType) && value.Valid, nil
	}
	out, _, err := walkValue(contentType, false, value, func(encrypted bool, v NullString) (NullString, bool, error) {
		return v, false, nil
	}, func(b caseBranch, mime string, value NullString) bool {
		if !EncryptionEligible(mime) || !value.Valid || b.bool("encrypted") {
			return false
		}
		b.set("encrypted", true)
		return true
	})
	return out, false, err
}

// ReencryptStoredValue rewrites every encrypted leaf that is not under the
// target key, and reports whether anything changed. Leaves already under
// the target key are left byte for byte.
func ReencryptStoredValue(contentType string, encrypted bool, value NullString, target string) (NullString, bool, error) {
	if Keys == nil {
		return value, false, ErrNoEncryptionKey
	}
	if _, ok := Keys.Key(target); !ok {
		return value, false, fmt.Errorf("%w: %s", ErrUnknownKeyID, target)
	}
	return walkValue(contentType, encrypted, value, func(encrypted bool, v NullString) (NullString, bool, error) {
		if !encrypted || !v.Valid || EnvelopeKeyID(v.String) == target {
			return v, false, nil
		}
		plaintext, err := Decrypt(v.String)
		if err != nil {
			return v, false, err
		}
		envelope, err := EncryptWith(target, plaintext)
		if err != nil {
			return v, false, err
		}
		return NullStringOf(envelope), true, nil
	}, nil)
}

// StoredKeyIDs lists the key ids the value's envelopes name (duplicates
// removed, order of appearance).
func StoredKeyIDs(contentType string, encrypted bool, value NullString) []string {
	var ids []string
	seen := map[string]bool{}
	_, _, _ = walkValue(contentType, encrypted, value, func(encrypted bool, v NullString) (NullString, bool, error) {
		if encrypted && v.Valid {
			if id := EnvelopeKeyID(v.String); id != "" && !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
		return v, false, nil
	}, nil)
	return ids
}
