package admin

import (
	"context"
	"fmt"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

// Reveals: the only way plaintext of an encrypted value leaves the server.
// A reveal is policy-checked against the caller's access, recorded in the
// audit log before any decryption happens (and thereby announced to the
// parameter's owners), and then decrypted.

// RevealParameter replaces the parameter's masked value with its plaintext
// for the caller: policy check, audit row, decryption. A parameter without
// encrypted values is left alone (no event).
func RevealParameter(ctx context.Context, p *Parameter) error {
	if !HasEncryptedValue(p.ContentType, p.Encrypted, p.Value) {
		return nil
	}
	requireWrite, err := revealRequiresWrite(ctx)
	if err != nil {
		return err
	}
	if !canReveal(requireWrite, p.RW) {
		return ErrRevealDenied
	}
	if err := recordAudit(ctx, p.ID, p.Path, AuditRevealView); err != nil {
		return err
	}
	plaintext, err := DecryptStoredValue(p.ContentType, p.Encrypted, p.storedValue)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}
	p.Value = plaintext
	return nil
}

// RevealLog decrypts the encrypted versions of a history page whose node the
// caller may reveal, recording one audit row per distinct node revealed.
// Versions the policy does not admit stay masked, without error.
func RevealLog(ctx context.Context, entries []LogEntry, path string) error {
	requireWrite, err := revealRequiresWrite(ctx)
	if err != nil {
		return err
	}
	recorded := map[int]bool{}
	for i := range entries {
		l := &entries[i]
		if !HasEncryptedValue(l.ContentType, l.Encrypted, l.Value) || !canReveal(requireWrite, l.RW) {
			continue
		}
		if !recorded[l.NodeID] {
			if err := recordAudit(ctx, l.NodeID, path, AuditRevealHistory); err != nil {
				return err
			}
			recorded[l.NodeID] = true
		}
		plaintext, err := DecryptStoredValue(l.ContentType, l.Encrypted, l.storedValue)
		if err != nil {
			return fmt.Errorf("%w: version %d: %v", ErrDecryptionFailed, l.Version, err)
		}
		l.Value = plaintext
	}
	return nil
}
