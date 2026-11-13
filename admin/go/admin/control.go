package admin

import (
	"context"
	"database/sql"
	"sync"

	"github.com/rs/zerolog/log"

	. "github.com/onlineconf/onlineconf/admin/go/common"
	"github.com/onlineconf/onlineconf/admin/go/resolver"
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

// watchedControl keeps a control parameter's text in memory for the read
// path, which consults it for every row it serves and must not pay a
// query. The resolver's tree reloads refresh it on every instance within
// one reload, about five seconds, with the node as the common graph
// resolves it (symlinks followed); the instance that handles a write or a
// deletion of the parameter refreshes it at once (controlWritten). Until
// the first reload has delivered it, it reads the database.
type watchedControl struct {
	param controlParameter
	mu    sync.RWMutex
	seen  bool
	value string
	ok    bool
}

// revealPolicyWatch serves the reveal policy to finishScan.
var revealPolicyWatch = &watchedControl{param: revealPolicy}

// watchControls registers the watched control parameters with the
// resolver; the tree is loaded right after, before the server serves.
func watchControls() {
	resolver.Synchronize(string(revealPolicy), revealPolicyWatch)
}

// Update implements resolver.Synchronized; param is nil when absent.
func (w *watchedControl) Update(ctx context.Context, param *resolver.Param) {
	value, ok := "", false
	if param != nil && param.ContentType == ContentTypeText && !param.Encrypted && param.Value.Valid {
		value, ok = param.Value.String, true
	}
	w.set(value, ok)
}

func (w *watchedControl) set(value string, ok bool) {
	w.mu.Lock()
	w.seen, w.value, w.ok = true, value, ok
	w.mu.Unlock()
}

// refresh re-reads the parameter from the database.
func (w *watchedControl) refresh(ctx context.Context) error {
	value, ok, err := w.param.text(ctx)
	if err != nil {
		return err
	}
	w.set(value, ok)
	return nil
}

// text returns the parameter's text as last seen; ok is false when there
// is none.
func (w *watchedControl) text(ctx context.Context) (value string, ok bool, err error) {
	w.mu.RLock()
	seen, value, ok := w.seen, w.value, w.ok
	w.mu.RUnlock()
	if seen {
		return value, ok, nil
	}
	return w.param.text(ctx)
}

// controlWritten is called by the write handlers after a successful write
// or deletion at path: the handling instance applies a change of a control
// parameter at once, the others at their next tree reload.
func controlWritten(ctx context.Context, path string) {
	switch controlParameter(path) {
	case revealPolicy:
		if err := revealPolicyWatch.refresh(ctx); err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("path", path).Msg("control parameter: refresh after a write failed, the next tree reload applies it")
		}
	case activeKeyID:
		// a rotation requested through this instance starts at once
		triggerKeyReconcile("parameter written")
	}
}
