package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"

	. "github.com/onlineconf/onlineconf/admin/go/common"
	"github.com/onlineconf/onlineconf/admin/go/resolver"
)

// Value encryption, admin side: the feature switch and the reveal policy
// (control parameters, see control.go), the key registry that records which
// key the stored data is encrypted under without holding any key material,
// the promotion of a new key requested by the operator through the key id
// parameter, and the background sweep that re-encrypts stored data under
// it.

var (
	ErrEncryptionNotAllowed   = errors.New("Encryption is not allowed for this content type")
	ErrEncryptionDisabled     = errors.New("Value encryption is disabled on this instance, see " + string(valueEncryptionSwitch))
	ErrEncryptedFieldRequired = errors.New("The parameter is encrypted: the encrypted field must be stated explicitly")
	ErrRevealDenied           = errors.New("Not allowed to reveal this value")
	ErrDecryptionFailed       = errors.New("Failed to decrypt the value")
	ErrActiveKeyMissing       = errors.New("The active encryption key is not in this instance's key file")
)

// isValueEncryptionEnabled reports whether encrypted writes are accepted on
// this instance: the switch is on and a keyring is loaded.
func isValueEncryptionEnabled(ctx context.Context) (bool, error) {
	if Keys == nil {
		return false, nil
	}
	return valueEncryptionSwitch.enabled(ctx)
}

// revealRequiresWrite reads the reveal policy: true unless the policy
// parameter says "ro". A request reads it once, through the memo its
// context carries (withPolicyMemo); a background context reads the database
// each time.
func revealRequiresWrite(ctx context.Context) (bool, error) {
	memo, _ := ctx.Value(policyMemoKey{}).(*policyMemo)
	if memo != nil {
		memo.mu.Lock()
		defer memo.mu.Unlock()
		if memo.read {
			return memo.requireWrite, nil
		}
	}
	value, ok, err := revealPolicy.text(ctx)
	if err != nil {
		return true, err
	}
	requireWrite := !(ok && strings.TrimSpace(value) == "ro")
	if memo != nil {
		memo.read, memo.requireWrite = true, requireWrite
	}
	return requireWrite, nil
}

// canReveal applies the policy to the caller's access to a parameter.
func canReveal(requireWrite bool, rw NullBool) bool {
	return rw.Valid && (rw.Bool || !requireWrite)
}

// --- advisory locks -----------------------------------------------------------

// errAdvisoryLocked reports that another admin instance holds the requested
// advisory lock (e.g. during a rolling deploy); callers skip the work.
var errAdvisoryLocked = errors.New("locked by another instance")

// acquireAdvisoryLock serializes a maintenance operation across admin
// instances with a MySQL advisory lock held on a dedicated connection
// (GET_LOCK is connection-scoped and released by the server if the
// connection dies, so a crashed holder cannot leave it stuck). The name is
// scoped to the current database: advisory locks are server-wide and
// installations may share a server. A negative wait waits for the lock
// without a bound.
func acquireAdvisoryLock(ctx context.Context, name string, wait time.Duration) (release func(), err error) {
	conn, err := DB.Conn(ctx)
	if err != nil {
		return nil, err
	}
	seconds := int(wait / time.Second)
	if wait < 0 {
		seconds = -1 // GET_LOCK's infinite timeout
	}
	var got sql.NullInt64
	row := conn.QueryRowContext(ctx, "SELECT GET_LOCK(CONCAT(DATABASE(), '/', ?), ?)", name, seconds)
	if err := row.Scan(&got); err != nil {
		conn.Close()
		return nil, err
	}
	if !got.Valid || got.Int64 != 1 {
		conn.Close()
		return nil, fmt.Errorf("%s: %w", name, errAdvisoryLocked)
	}
	return func() {
		// ignore errors: closing the connection releases the lock anyway
		conn.ExecContext(ctx, "DO RELEASE_LOCK(CONCAT(DATABASE(), '/', ?))", name)
		conn.Close()
	}, nil
}

// --- key registry -------------------------------------------------------------

// fatalEncryptionError is a startup condition that must stop the server
// (as opposed to a transient database error, which is retried).
type fatalEncryptionError struct{ msg string }

func (e *fatalEncryptionError) Error() string { return e.msg }

const encryptionStartupRetry = 5 * time.Second

// runStartupChecks runs the checks that need the database: the schema
// version (ensureSchemaCurrent) and then value encryption — refuse to run
// without a keyring when live encrypted values exist, verify every loaded
// key against the registry, register unknown keys, and reconcile the wanted
// key, and hand the registry work — the first registration, a promotion,
// the consistency pass — to the sweep goroutine, armed here. A database
// that cannot be reached is retried in the background; the server serves
// in the meantime (the resolver fails closed on its own if the keys are
// wrong), and a verdict against the schema or the keys stops it whenever
// it is reached. The key id parameter is watched from the start through
// the resolver's tree reloads, which trigger the goroutine once it is
// armed.
func runStartupChecks() {
	resolver.Synchronize(string(activeKeyID), keyIDWatch)
	ctx := log.Logger.WithContext(context.Background())
	run := func() bool {
		if !ensureSchemaCurrent() {
			return false
		}
		err := encryptionStartup(ctx)
		var fatal *fatalEncryptionError
		if errors.As(err, &fatal) {
			log.Fatal().Msg(fatal.msg)
		}
		if err != nil {
			log.Warn().Err(err).Msg("encryption startup check failed, will retry")
			return false
		}
		return true
	}
	if run() {
		return
	}
	go func() {
		for !run() {
			time.Sleep(encryptionStartupRetry)
		}
	}()
}

func encryptionStartup(ctx context.Context) error {
	if Keys == nil {
		// what the resolver serves: live rows. A deleted secret and the sealed
		// history need the key too, but only at a restore or a reveal, which
		// then fail with DecryptionFailed; an installation that declassified
		// or deleted every secret may drop its key file
		var n int
		if err := DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM my_config_tree WHERE Encrypted = 1 AND NOT Deleted").Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			return &fatalEncryptionError{"encrypted parameters exist but encryption.key_file is not configured"}
		}
		return nil
	}

	for _, key := range Keys.All() {
		if err := registerKey(ctx, key); err != nil {
			return err
		}
	}
	// the watcher starts from the value this startup acts on, so that the
	// first tree load triggers a reconciliation only if the parameter
	// changed meanwhile
	value, _, err := activeKeyID.text(ctx)
	if err != nil {
		return err
	}
	keyIDWatch.prime(value)
	if err := reconcileActiveKey(ctx); err != nil {
		return err
	}
	// the startup judged; the sweep goroutine does the work, now and at
	// every change of the key id the watcher sees
	armSweep()
	triggerKeyReconcile()
	return nil
}

// --- the active key -----------------------------------------------------------

// wantedKeyID is the key the installation wants new encryptions under: the
// key id control parameter when present and not blank, else the registry's
// active key, else the key file's first key (a fresh installation). The
// file's order therefore stops meaning anything once a key is registered.
// The source is named for the log and the status endpoint.
func wantedKeyID(ctx context.Context) (id, source string, err error) {
	if value, ok, err := activeKeyID.text(ctx); err != nil {
		return "", "", err
	} else if ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value), "parameter", nil
	}
	if id, ok, err := registeredActiveKey(ctx); err != nil {
		return "", "", err
	} else if ok {
		return id, "registry", nil
	}
	return Keys.First().ID, "key file", nil
}

// keyAction is what reconcileActiveKey does about the wanted key.
type keyAction int

const (
	keyActionNone     keyAction = iota // the wanted key is the key in use
	keyActionRegister                  // a fresh installation: record the wanted key as active
	keyActionPromote                   // re-encrypt everything under the wanted key, then record it
	keyActionRefuse                    // the wanted key is not held: say so, change nothing
	keyActionFatal                     // a fresh installation wanting a key the file lacks
)

// planActiveKey is the decision table of reconcileActiveKey: the wanted key
// and its source against the registry's active key (hasActive is false on
// a fresh installation) and whether this instance holds the wanted key. The
// message is logged with the action; the fatal one stops the server.
func planActiveKey(wanted, source, active string, hasActive, held bool) (keyAction, string) {
	switch {
	case !hasActive && !held:
		return keyActionFatal, fmt.Sprintf("the encryption key %s wanted by the %s is not in the key file", wanted, source)
	case !hasActive:
		return keyActionRegister, "encryption key registered as active"
	case wanted == active:
		return keyActionNone, "encryption key in use"
	case !held:
		return keyActionRefuse, "the wanted encryption key is not in this instance's key file: new encryptions stay under the key in use"
	default:
		return keyActionPromote, "encryption key rotation requested, re-encrypting stored values in the background"
	}
}

// reconcileActiveKey judges the wanted key at startup, so that the outcomes
// that must stop the server are fatal before it serves: the registry's
// active key not in the key file, or a fresh installation wanting a key the
// file lacks. The work itself — the registration, the promotion or the
// consistency pass — is the sweep goroutine's, which makes the same plan
// again under the lock (runEncryptionSweep) once the startup has armed and
// triggered it.
func reconcileActiveKey(ctx context.Context) error {
	wanted, source, err := wantedKeyID(ctx)
	if err != nil {
		return err
	}
	active, hasActive, err := registeredActiveKey(ctx)
	if err != nil {
		return err
	}
	if hasActive {
		if _, held := Keys.Key(active); !held {
			return &fatalEncryptionError{fmt.Sprintf("the registered active encryption key %s is not in the key file", active)}
		}
	}
	_, held := Keys.Key(wanted)
	if action, message := planActiveKey(wanted, source, active, hasActive, held); action == keyActionFatal {
		return &fatalEncryptionError{message}
	}
	return nil
}

// activeKeyForWrite resolves the key a write transaction encrypts under: the
// registry's active key, read under a shared lock in the transaction, so
// that a promotion's flip of the mark (setRegisteredActive) waits for every
// in-flight encrypted write under the old key, and every write started
// after the flip sees the new key. The read must follow every row lock the
// write takes — the parameter row, the parent, the path row of a create —
// and precede its updates, so that a transaction holding the shared key
// lock never waits for a tree row. With no key recorded yet, on a fresh
// installation whose startup reconciliation is still retrying, the key
// file's first key is used, the one the reconciliation will record.
func activeKeyForWrite(ctx context.Context, tx *sql.Tx) (string, error) {
	var id string
	err := tx.QueryRowContext(ctx, "SELECT KeyID FROM my_config_encryption_key WHERE Active LOCK IN SHARE MODE").Scan(&id)
	if err == sql.ErrNoRows {
		return Keys.First().ID, nil
	} else if err != nil {
		return "", err
	}
	if _, held := Keys.Key(id); !held {
		return "", fmt.Errorf("%w: %s", ErrActiveKeyMissing, id)
	}
	return id, nil
}

// keyIDWatcher follows the key id parameter through the resolver's tree
// reloads, which every instance performs: the one trigger of a
// reconciliation outside startup, so that a rotation requested through any
// instance is noticed by every instance at its next reload. It compares
// what the reader of the parameter would see (controlParameter.text: a
// plain text, anything else absent), and is primed at startup with the
// value the startup acted on.
type keyIDWatcher struct {
	mu   sync.Mutex
	seen bool
	last string
}

var keyIDWatch = &keyIDWatcher{}

func (w *keyIDWatcher) prime(value string) {
	w.mu.Lock()
	w.seen, w.last = true, value
	w.mu.Unlock()
}

func (w *keyIDWatcher) Update(ctx context.Context, param *resolver.Param) {
	value := ""
	if param != nil && param.ContentType == ContentTypeText && !param.Encrypted && param.Value.Valid {
		value = param.Value.String
	}
	w.mu.Lock()
	changed := !w.seen || value != w.last
	w.seen, w.last = true, value
	w.mu.Unlock()
	if changed {
		triggerKeyReconcile()
	}
}

// registerKey verifies a key against its registry row, or registers it. A
// key whose material differs from the registered one under the same id is
// fatal: it is the classic operator error, and it would surface only as
// unexplained decryption failures otherwise.
func registerKey(ctx context.Context, key *Key) error {
	verifier, err := selectVerifier(ctx, key.ID)
	if err == sql.ErrNoRows {
		v, err := key.Verifier()
		if err != nil {
			return err
		}
		if _, err := DB.ExecContext(ctx, "INSERT IGNORE INTO my_config_encryption_key (KeyID, Verifier) VALUES (?, ?)", key.ID, v); err != nil {
			return err
		}
		// re-read: another instance may have registered the same id first
		verifier, err = selectVerifier(ctx, key.ID)
		if err != nil {
			return err
		}
		log.Info().Str("key", key.ID).Msg("encryption key registered")
	} else if err != nil {
		return err
	}
	if !key.CheckVerifier(verifier) {
		return &fatalEncryptionError{fmt.Sprintf("encryption key %s in the key file differs from the key registered under that id in the database", key.ID)}
	}
	return nil
}

func selectVerifier(ctx context.Context, keyID string) ([]byte, error) {
	var verifier []byte
	err := DB.QueryRowContext(ctx, "SELECT Verifier FROM my_config_encryption_key WHERE KeyID = ?", keyID).Scan(&verifier)
	return verifier, err
}

func registeredActiveKey(ctx context.Context) (id string, ok bool, err error) {
	err = DB.QueryRowContext(ctx, "SELECT KeyID FROM my_config_encryption_key WHERE Active").Scan(&id)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	return id, err == nil, err
}

// setRegisteredActive flips the registry's mark to the key: an update of
// every row, which waits for the shared locks of the in-flight writes that
// read the old key (activeKeyForWrite) and makes every later write use the
// new one.
func setRegisteredActive(ctx context.Context, keyID string) error {
	_, err := DB.ExecContext(ctx, "UPDATE my_config_encryption_key SET Active = (KeyID = ?)", keyID)
	return err
}

// registerActiveKey records the key as the registry's active one on a fresh
// installation, unless a peer did first: a locking read of the active row in
// a transaction, then the flip only when there is none — two instances
// starting together register once, the first writer winning, whatever
// their key files' order. It returns the key recorded: the given one, or
// the peer's.
func registerActiveKey(ctx context.Context, keyID string) (string, error) {
	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var active string
	err = tx.QueryRowContext(ctx, "SELECT KeyID FROM my_config_encryption_key WHERE Active FOR UPDATE").Scan(&active)
	if err == nil {
		return active, nil
	} else if err != sql.ErrNoRows {
		return "", err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE my_config_encryption_key SET Active = (KeyID = ?)", keyID); err != nil {
		return "", err
	}
	return keyID, tx.Commit()
}

// --- sweep --------------------------------------------------------------------

// sweepStatus is the per-instance state of the last sweep, served by the
// status endpoint. done means verified on this instance: the last pass
// walked every encrypted row of both tables under the target and rewrote
// nothing; waiting is the goroutine queued behind another instance's lock,
// before it knows its job; refused a wanted or a recorded key this
// instance's file lacks.
type sweepStatus struct {
	State       string `json:"state"` // idle, waiting, running, done, failed, refused
	Target      string `json:"target,omitempty"`
	Scanned     int    `json:"scanned"`
	Reencrypted int    `json:"reencrypted"`
	Error       string `json:"error,omitempty"`
	StartedAt   string `json:"started_at,omitempty"`
	FinishedAt  string `json:"finished_at,omitempty"`
}

var (
	sweepMu    sync.Mutex
	sweepState = sweepStatus{State: "idle"}
)

// the one sweep goroutine of this instance (triggerKeyReconcile): armed once
// the startup succeeded, running while it waits for the lock or works, marked
// to run again by a trigger that arrived meanwhile, and scheduled again after
// a failure with a growing delay
var (
	sweepArmed, sweepRunning, sweepRerun bool

	sweepRetry      *time.Timer
	sweepRetryDelay = sweepRetryMin
)

// runSweep is runEncryptionSweep; the unit tests replace it.
var runSweep = runEncryptionSweep

// lockRetryPause is the first pause between two attempts at an operation a
// lock conflict failed; it doubles each time. The unit tests shorten it.
var lockRetryPause = time.Second

const (
	sweepBatch        = 500
	sweepMaxPasses    = 5
	sweepRetryMin     = time.Minute
	sweepRetryMax     = 10 * time.Minute
	lockRetryAttempts = 5
)

// armSweep lets triggers start the sweep goroutine. The startup calls it
// once the keys are registered and verified, so that a registration by the
// goroutine finds the verifier rows, and a tree load delivered while the
// startup retries triggers nothing.
func armSweep() {
	sweepMu.Lock()
	sweepArmed = true
	sweepMu.Unlock()
}

// triggerKeyReconcile starts this instance's sweep goroutine, or marks it to
// run again when it is already waiting for the lock or working, so that a
// change of the key id observed during a sweep is never lost on this
// instance. Before the startup armed it, and without a keyring, it does
// nothing. The resolver's watcher of the key id parameter calls it at every
// change, the startup once, and a failed sweep again after a delay.
func triggerKeyReconcile() {
	sweepMu.Lock()
	defer sweepMu.Unlock()
	if !sweepArmed || Keys == nil {
		return
	}
	if sweepRetry != nil {
		sweepRetry.Stop()
		sweepRetry = nil
	}
	if sweepRunning {
		sweepRerun = true
		return
	}
	sweepRunning = true
	go sweepLoop()
}

// sweepLoop runs the sweep, again as long as triggers arrived meanwhile, and
// schedules a retry after a failure: after a minute, then twice as long each
// time up to ten minutes, so that a persistent cause is logged every ten
// minutes at most and a transient one heals without a restart.
func sweepLoop() {
	for {
		ok := runSweep()
		sweepMu.Lock()
		if sweepRerun {
			sweepRerun = false
			sweepMu.Unlock()
			continue
		}
		sweepRunning = false
		if ok {
			sweepRetryDelay = sweepRetryMin
		} else {
			sweepRetry = time.AfterFunc(sweepRetryDelay, triggerKeyReconcile)
			if sweepRetryDelay *= 2; sweepRetryDelay > sweepRetryMax {
				sweepRetryDelay = sweepRetryMax
			}
		}
		sweepMu.Unlock()
		return
	}
}

func setSweepState(s sweepStatus) {
	sweepMu.Lock()
	sweepState = s
	sweepMu.Unlock()
}

// runEncryptionSweep is one run of the sweep goroutine. It takes the
// advisory lock — waiting without a bound behind another instance's sweep —
// and plans under it what the wanted key needs: a promotion re-encrypts
// every stored value under the wanted key, records it as the registry's
// active key and walks both tables again until a pass rewrites nothing;
// with the wanted key recorded already, the passes alone — the consistency
// pass of a startup, and the share of the work left to an instance that
// queued behind the promotion, whose clean pass is this instance's
// verification; on a fresh installation the registration first. A pass
// walks both tables in full and rewrites every envelope under another key,
// with a compare-and-swap so that a concurrent write always wins. The flip
// of the registry's mark after the first pass waits for every in-flight
// encrypted write under the old key and makes every later write use the
// target (activeKeyForWrite). Versions, timestamps and notifications are
// never touched. It reports false after a failure, which the goroutine
// retries later.
func runEncryptionSweep() bool {
	ctx := log.Logger.WithContext(context.Background())
	setSweepState(sweepStatus{State: "waiting"})
	release, err := acquireAdvisoryLock(ctx, "encryption-sweep", -1)
	if err != nil {
		log.Error().Err(err).Msg("failed to acquire the encryption sweep lock")
		setSweepState(sweepStatus{State: "failed", Error: err.Error(), FinishedAt: time.Now().Format(time.RFC3339)})
		return false
	}
	defer release()

	refuse := func(target, message string) bool {
		setSweepState(sweepStatus{State: "refused", Target: target, Error: message, FinishedAt: time.Now().Format(time.RFC3339)})
		return true
	}
	fail := func(target string, err error) bool {
		log.Error().Err(err).Str("key", target).Msg("re-encryption of stored values failed")
		setSweepState(sweepStatus{State: "failed", Target: target, Error: err.Error(), FinishedAt: time.Now().Format(time.RFC3339)})
		return false
	}

	// the plan, made under the lock: what the wanted key needs now
	wanted, source, err := wantedKeyID(ctx)
	if err != nil {
		return fail("", err)
	}
	active, hasActive, err := registeredActiveKey(ctx)
	if err != nil {
		return fail("", err)
	}
	if hasActive {
		if _, held := Keys.Key(active); !held {
			// a peer promoted a key this instance lacks (fatal at startup):
			// nothing to re-encrypt with, and its writes fail closed already
			message := fmt.Sprintf("the registered active encryption key %s is not in the key file", active)
			log.Error().Str("key", active).Msg(message)
			return refuse(active, message)
		}
	}
	target, promotion := active, false
	for {
		_, held := Keys.Key(wanted)
		action, message := planActiveKey(wanted, source, active, hasActive, held)
		if action == keyActionRegister {
			registered, err := registerActiveKey(ctx, wanted)
			if err != nil {
				return fail(wanted, err)
			}
			if registered != wanted {
				// a peer registered first: its key is the one in use, and
				// the plan is made again with it, the wanted key re-read
				// (without the parameter it is now the registry's)
				log.Info().Str("key", registered).Str("wanted", wanted).Msg("encryption key registered as active by a peer")
				if _, held := Keys.Key(registered); !held {
					message := fmt.Sprintf("the registered active encryption key %s is not in the key file", registered)
					log.Error().Str("key", registered).Msg(message)
					return refuse(registered, message)
				}
				active, hasActive, target = registered, true, registered
				if wanted, source, err = wantedKeyID(ctx); err != nil {
					return fail(registered, err)
				}
				continue
			}
			log.Info().Str("key", wanted).Str("source", source).Msg(message)
			target = wanted
			break
		}
		switch action {
		case keyActionFatal, keyActionRefuse:
			log.Error().Str("key", wanted).Str("source", source).Str("active", active).Msg(message)
			return refuse(wanted, message)
		case keyActionNone:
			log.Info().Str("key", active).Msg(message)
		case keyActionPromote:
			log.Info().Str("from", active).Str("to", wanted).Str("source", source).Msg(message)
			target, promotion = wanted, true
		}
		break
	}

	setSweepState(sweepStatus{State: "running", Target: target, StartedAt: time.Now().Format(time.RFC3339)})
	scanned, reencrypted, err := sweepPasses(ctx, target, promotion)
	sweepMu.Lock()
	sweepState.Scanned, sweepState.Reencrypted = scanned, reencrypted
	sweepState.FinishedAt = time.Now().Format(time.RFC3339)
	if err != nil {
		sweepState.State, sweepState.Error = "failed", err.Error()
	} else {
		sweepState.State = "done"
	}
	sweepMu.Unlock()
	if err != nil {
		log.Error().Err(err).Bool("promotion", promotion).Str("key", target).Msg("re-encryption of stored values failed")
		return false
	}
	event := log.Info()
	if reencrypted > 0 && !promotion {
		// a pass without a promotion found rows under another key: a peer
		// wrote them before it learned of the rotation
		event = log.Warn()
	}
	event.Bool("promotion", promotion).Int("scanned", scanned).Int("reencrypted", reencrypted).Str("key", target).Msg("re-encryption of stored values finished")
	return true
}

// sweepPasses re-encrypts under the target: for a promotion the first pass,
// the flip and the passes until one rewrites nothing; otherwise the passes
// alone. Bounded by sweepMaxPasses, which the shared-lock read of the key
// (activeKeyForWrite) makes unreachable.
func sweepPasses(ctx context.Context, target string, promotion bool) (scanned, reencrypted int, err error) {
	if promotion {
		if scanned, reencrypted, err = encryptionSweep(ctx, target); err != nil {
			return
		}
		if err = withLockRetry(ctx, "the flip of the key registry", func() error { return setRegisteredActive(ctx, target) }); err != nil {
			return
		}
	}
	for pass := 0; ; pass++ {
		if pass == sweepMaxPasses {
			return scanned, reencrypted, fmt.Errorf("stored values were still being written under another key after %d passes", sweepMaxPasses)
		}
		s, r, err := encryptionSweep(ctx, target)
		if err != nil {
			return scanned, reencrypted, err
		}
		scanned, reencrypted = scanned+s, reencrypted+r
		if r == 0 {
			return scanned, reencrypted, nil
		}
	}
}

// withLockRetry runs an operation again after a deadlock (1213) or a lock
// wait timeout (1205), a few times with a growing pause. The flip of the key
// registry waits for every in-flight encrypted write and can be the victim
// InnoDB picks to break a cycle — a create holds a gap lock on the path
// index before it reads the key, and a second create under another parent
// can wait for that gap while holding the shared key lock — and the
// re-encryption of a row can wait behind a write that holds the row longer
// than the server's lock wait timeout, sealing a long history.
func withLockRetry(ctx context.Context, what string, op func() error) error {
	pause := lockRetryPause
	for attempt := 1; ; attempt++ {
		err := op()
		if err == nil || !isRetryableLockError(err) || attempt == lockRetryAttempts {
			return err
		}
		log.Ctx(ctx).Warn().Err(err).Int("attempt", attempt).Msg("retrying " + what)
		time.Sleep(pause)
		pause *= 2
	}
}

// isRetryableLockError reports a deadlock or a lock wait timeout.
func isRetryableLockError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && (mysqlErr.Number == 1213 || mysqlErr.Number == 1205)
}

// encryptionSweep is one pass over both tables.
func encryptionSweep(ctx context.Context, target string) (scanned, reencrypted int, err error) {
	for _, table := range []string{"my_config_tree", "my_config_tree_log"} {
		if err = reencryptTable(ctx, table, target, &scanned, &reencrypted); err != nil {
			return
		}
	}
	return
}

type sweepRow struct {
	id    int64
	value NullString
}

// reencryptTable walks the encrypted rows of a table in ID batches and
// rewrites every envelope under a key other than the target, with a
// compare-and-swap so that a concurrent write always wins. The batch query
// says Encrypted = 1 and forces the index: the bare column becomes the range
// Encrypted <> 0, which cannot deliver the rows in ID order, so that every
// batch would read and sort every encrypted row above the last id — a pass
// quadratic in the encrypted rows.
func reencryptTable(ctx context.Context, table, target string, scanned, reencrypted *int) error {
	lastID := int64(0)
	for {
		rows, err := DB.QueryContext(ctx, "SELECT ID, Value FROM "+table+" FORCE INDEX (Encrypted) WHERE Encrypted = 1 AND ID > ? ORDER BY ID LIMIT ?", lastID, sweepBatch)
		if err != nil {
			return err
		}
		batch := make([]sweepRow, 0, sweepBatch)
		for rows.Next() {
			var r sweepRow
			if err := rows.Scan(&r.id, &r.value); err != nil {
				rows.Close()
				return err
			}
			batch = append(batch, r)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
		if len(batch) == 0 {
			return nil
		}
		for _, r := range batch {
			lastID = r.id
			*scanned++
			newValue, changed, err := Reencrypt(r.value.String, target)
			if err != nil {
				return fmt.Errorf("%s row %d: %w", table, r.id, err)
			}
			if !changed {
				continue
			}
			var res sql.Result
			if err := withLockRetry(ctx, "the re-encryption of a row", func() error {
				var err error
				res, err = DB.ExecContext(ctx, "UPDATE "+table+" SET Value = ? WHERE ID = ? AND Value = ?", newValue, r.id, r.value)
				return err
			}); err != nil {
				return err
			}
			if n, _ := res.RowsAffected(); n == 1 {
				*reencrypted++
			}
		}
	}
}

// --- status -------------------------------------------------------------------

type keyStatus struct {
	ID         string `json:"id"`
	Registered bool   `json:"registered"` // present in the registry
	Verified   bool   `json:"verified"`   // material matches the registry
	Recorded   bool   `json:"recorded"`   // the registry's Active mark: the key in use
	CreatedAt  string `json:"created_at,omitempty"`
}

// wantedKeyStatus is the key the installation wants new encryptions under
// (see wantedKeyID) and whether this instance holds it.
type wantedKeyStatus struct {
	ID     string `json:"id"`
	Source string `json:"source"` // parameter, registry, key file
	Held   bool   `json:"held"`
}

type EncryptionStatus struct {
	Configured bool             `json:"configured"`
	Wanted     *wantedKeyStatus `json:"wanted,omitempty"`
	Keys       []keyStatus      `json:"keys"`
	Instance   string           `json:"instance"`
	Sweep      sweepStatus      `json:"sweep"`
}

// instanceName names this instance in the status: the sweep state is this
// instance's, and several instances may answer behind one address.
var instanceName, _ = os.Hostname()

// SelectEncryptionStatus answers the operator's questions "did the rotation
// take?" and "may the old key go now?": the wanted key, the keyring's keys
// with their registry state (recorded marks the key in use), and the sweep
// state of the instance that answers, which it names.
func SelectEncryptionStatus(ctx context.Context) (*EncryptionStatus, error) {
	status := &EncryptionStatus{Configured: Keys != nil, Keys: []keyStatus{}, Instance: instanceName}
	rows, err := DB.QueryContext(ctx, "SELECT KeyID, Verifier, Active, CreatedAt FROM my_config_encryption_key ORDER BY CreatedAt, KeyID")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	registered := map[string]keyStatus{}
	var order []string
	for rows.Next() {
		var ks keyStatus
		var verifier []byte
		if err := rows.Scan(&ks.ID, &verifier, &ks.Recorded, &ks.CreatedAt); err != nil {
			return nil, err
		}
		ks.Registered = true
		if Keys != nil {
			if key, ok := Keys.Key(ks.ID); ok {
				ks.Verified = key.CheckVerifier(verifier)
			}
		}
		registered[ks.ID] = ks
		order = append(order, ks.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if Keys != nil {
		for _, key := range Keys.All() {
			if _, ok := registered[key.ID]; !ok {
				registered[key.ID] = keyStatus{ID: key.ID}
				order = append(order, key.ID)
			}
		}
		id, source, err := wantedKeyID(ctx)
		if err != nil {
			return nil, err
		}
		_, held := Keys.Key(id)
		status.Wanted = &wantedKeyStatus{ID: id, Source: source, Held: held}
	}
	for _, id := range order {
		status.Keys = append(status.Keys, registered[id])
	}
	sweepMu.Lock()
	status.Sweep = sweepState
	sweepMu.Unlock()
	return status, nil
}
