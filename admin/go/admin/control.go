package admin

import (
	"context"
	"database/sql"
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

// Control parameters are the /onlineconf settings the admin reads for its
// own behavior. They are read by path — not through the requesting user's
// ACL, since a global fact must not depend on who asks — following
// symlinks up to controlSymlinkHops, and only a plain, unencrypted text at
// the end of the chain counts: anything else is "absent".

// controlSymlinkHops bounds the symlink chain of a control parameter. It
// is a constant of its own rather than the deletion check's symlink depth,
// which is itself destined to become a control parameter.
const controlSymlinkHops = 5

// controlParameter is the path of such a setting.
type controlParameter string

const (
	// valueEncryptionSwitch enables encrypted writes when truthy. Absent
	// means off; fresh installations are seeded with "1" by schema.sql and
	// the example dumps.
	valueEncryptionSwitch = controlParameter("/onlineconf/value-encryption")
	// revealPolicy holds "rw" (default: write access needed to reveal a
	// secret) or "ro" (read access suffices).
	revealPolicy = controlParameter("/onlineconf/value-encryption/reveal")
	// activeKeyID names the key new encryptions use; writing a new id is
	// how a key is rotated. Absent, the registry's active key stays.
	activeKeyID = controlParameter("/onlineconf/value-encryption/key-id")
	// deletedSymlinksCheckDisabled switches the deletion check for
	// referring symlinks and templates off when truthy.
	deletedSymlinksCheckDisabled = controlParameter("/onlineconf/disable-deleted-key-symlinks-check")
)

// text returns the parameter's plain text; ok is false when there is none.
func (c controlParameter) text(ctx context.Context) (value string, ok bool, err error) {
	path := string(c)
	for hop := 0; hop <= controlSymlinkHops; hop++ {
		row := DB.QueryRowContext(ctx, "SELECT ContentType, Encrypted, Value FROM my_config_tree WHERE Path = ? AND NOT Deleted", path)
		var contentType string
		var encrypted bool
		var stored sql.NullString
		err := row.Scan(&contentType, &encrypted, &stored)
		if err == sql.ErrNoRows {
			return "", false, nil
		} else if err != nil {
			return "", false, err
		}
		if contentType == ContentTypeSymlink && stored.Valid {
			path = stored.String
			continue
		}
		if contentType != ContentTypeText || encrypted || !stored.Valid {
			return "", false, nil
		}
		return stored.String, true, nil
	}
	log.Ctx(ctx).Warn().Str("path", string(c)).Int("hops", controlSymlinkHops).Msg("control parameter: symlink chain too long or cyclic, treated as absent")
	return "", false, nil
}

// enabled reports whether the parameter holds a truthy value; absent means
// off.
func (c controlParameter) enabled(ctx context.Context) (bool, error) {
	value, ok, err := c.text(ctx)
	if err != nil || !ok {
		return false, err
	}
	return truthy(value), nil
}

// truthy is the switch convention of /onlineconf: anything but "" and "0".
func truthy(value string) bool {
	return value != "" && value != "0"
}

// policyMemo holds the reveal policy for one request: a request that serves
// rows reads the policy once (revealRequiresWrite), so that a change of the
// policy applies to the next request on every instance.
// policyMemoMiddleware attaches one to every request of the admin API; a
// background context carries none and reads the database each time.
type policyMemo struct {
	mu           sync.Mutex
	read         bool
	requireWrite bool
}

type policyMemoKey struct{}

// withPolicyMemo attaches a fresh memo to a request's context.
func withPolicyMemo(ctx context.Context) context.Context {
	return context.WithValue(ctx, policyMemoKey{}, &policyMemo{})
}

// policyMemoMiddleware gives every request of the admin API its memo.
func policyMemoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(w, req.WithContext(withPolicyMemo(req.Context())))
	})
}
