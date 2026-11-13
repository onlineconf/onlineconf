package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

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

// encryptedRowsWhere matches rows holding an encrypted leaf: the node-level
// marker, or a case with an encrypted branch (its node-level marker is off;
// the storage boundary serializes branch markers compactly).
const encryptedRowsWhere = `(Encrypted OR (ContentType = '` + ContentTypeCase + `' AND Value LIKE '%"encrypted":true%'))`

// isValueEncryptionEnabled reports whether encrypted writes are accepted on
// this instance: the switch is on and a keyring is loaded.
func isValueEncryptionEnabled(ctx context.Context) (bool, error) {
	if Keys == nil {
		return false, nil
	}
	return valueEncryptionSwitch.enabled(ctx)
}

// revealRequiresWrite reads the reveal policy as the resolver's tree
// reloads deliver it: true unless the policy parameter says "ro".
func revealRequiresWrite(ctx context.Context) (bool, error) {
	value, ok, err := revealPolicyWatch.text(ctx)
	if err != nil {
		return true, err
	}
	return !(ok && strings.TrimSpace(value) == "ro"), nil
}

// revealPolicyName is the policy as served by /ui-config.
func revealPolicyName(requireWrite bool) string {
	if requireWrite {
		return "rw"
	}
	return "ro"
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
// installations may share a server.
func acquireAdvisoryLock(ctx context.Context, name string, wait time.Duration) (release func(), err error) {
	conn, err := DB.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var got sql.NullInt64
	row := conn.QueryRowContext(ctx, "SELECT GET_LOCK(CONCAT(DATABASE(), '/', ?), ?)", name, int(wait/time.Second))
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

// initializeEncryption runs the startup checks: refuse to run without
// a keyring when encrypted values exist, verify every loaded key against the
// registry, register unknown keys, and reconcile the wanted key with the
// registered one (the consistency pass, or a promotion). Database errors
// are retried in the background; the server serves in the meantime (the
// resolver fails closed on its own if the keys are wrong). The key id
// parameter is watched from then on through the resolver's tree reloads.
func initializeEncryption() {
	resolver.Synchronize(string(activeKeyID), &keyIDWatcher{})
	ctx := log.Logger.WithContext(context.Background())
	run := func() bool {
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
		var n int
		if err := DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM my_config_tree WHERE "+encryptedRowsWhere).Scan(&n); err != nil {
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
	return reconcileActiveKey(ctx, true)
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
	if id, _, ok, err := registeredActiveKey(ctx); err != nil {
		return "", "", err
	} else if ok {
		return id, "registry", nil
	}
	return Keys.First().ID, "key file", nil
}

// reconcileActiveKey compares the wanted key with the registry's active key
// and acts: a fresh installation registers the wanted key; new encryptions
// adopt the registered one; a wanted key that is newer and held is promoted
// by the background sweep; the operator's mistakes are reported and change
// nothing. It runs at startup (then with the consistency pass), after the
// key id parameter was written through this instance, and when a tree
// reload shows it changed.
func reconcileActiveKey(ctx context.Context, startup bool) error {
	wanted, source, err := wantedKeyID(ctx)
	if err != nil {
		return err
	}
	activeID, activeSeq, hasActive, err := registeredActiveKey(ctx)
	if err != nil {
		return err
	}
	if !hasActive {
		if !Keys.SetActive(wanted) {
			return &fatalEncryptionError{fmt.Sprintf("the encryption key %s wanted by the %s is not in the key file", wanted, source)}
		}
		if err := setRegisteredActive(ctx, wanted); err != nil {
			return err
		}
		log.Info().Str("key", wanted).Str("source", source).Msg("encryption key registered as active")
		return nil
	}
	if !Keys.SetActive(activeID) {
		return &fatalEncryptionError{fmt.Sprintf("the registered active encryption key %s is not in the key file", activeID)}
	}
	if wanted == activeID {
		if startup {
			go runEncryptionSweep(false, activeID)
		}
		return nil
	}
	if _, held := Keys.Key(wanted); !held {
		log.Error().Str("key", wanted).Str("source", source).Str("active", activeID).
			Msg("the wanted encryption key is not in this instance's key file: new encryptions stay under the active key")
		return nil
	}
	wantedSeq, err := keySeq(ctx, wanted)
	if err != nil {
		return err
	}
	if wantedSeq < activeSeq {
		log.Error().Str("key", wanted).Str("active", activeID).
			Msg("the wanted encryption key is older than the active one: a key never becomes active twice, rotate forward")
		return nil
	}
	log.Info().Str("from", activeID).Str("to", wanted).Str("source", source).
		Msg("encryption key rotation requested, re-encrypting stored values in the background")
	go runEncryptionSweep(true, wanted)
	return nil
}

// triggerKeyReconcile runs the reconciliation outside startup: after the
// key id parameter was written through this instance, or a tree reload
// showed it changed. Nothing here may stop a serving instance; every
// outcome is logged.
func triggerKeyReconcile(reason string) {
	if Keys == nil {
		return
	}
	go func() {
		ctx := log.Logger.WithContext(context.Background())
		if err := reconcileActiveKey(ctx, false); err != nil {
			log.Error().Err(err).Str("reason", reason).Msg("encryption key reconciliation failed")
		}
	}()
}

// adoptRegisteredKey makes new encryptions use the registry's active key,
// which a promotion on another instance may have changed since this one
// looked: one point query per encrypted write.
func adoptRegisteredKey(ctx context.Context) error {
	if Keys == nil {
		return nil
	}
	id, _, ok, err := registeredActiveKey(ctx)
	if err != nil || !ok {
		return err
	}
	if !Keys.SetActive(id) {
		return fmt.Errorf("%w: %s", ErrActiveKeyMissing, id)
	}
	return nil
}

// keyIDWatcher follows the key id parameter through the resolver's tree
// reloads, which every instance performs, so that a rotation requested
// through another instance is noticed within seconds.
type keyIDWatcher struct {
	mu   sync.Mutex
	seen bool
	last string
}

func (w *keyIDWatcher) Update(ctx context.Context, param *resolver.Param) {
	value := ""
	if param != nil && param.Value.Valid {
		value = param.Value.String
	}
	w.mu.Lock()
	changed := !w.seen || value != w.last
	w.seen, w.last = true, value
	w.mu.Unlock()
	if changed {
		triggerKeyReconcile("tree reload")
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

func keySeq(ctx context.Context, keyID string) (int64, error) {
	var seq int64
	err := DB.QueryRowContext(ctx, "SELECT Seq FROM my_config_encryption_key WHERE KeyID = ?", keyID).Scan(&seq)
	return seq, err
}

func registeredActiveKey(ctx context.Context) (id string, seq int64, ok bool, err error) {
	err = DB.QueryRowContext(ctx, "SELECT KeyID, Seq FROM my_config_encryption_key WHERE Active").Scan(&id, &seq)
	if err == sql.ErrNoRows {
		return "", 0, false, nil
	}
	return id, seq, err == nil, err
}

func setRegisteredActive(ctx context.Context, keyID string) error {
	_, err := DB.ExecContext(ctx, "UPDATE my_config_encryption_key SET Active = (KeyID = ?)", keyID)
	return err
}

// --- sweep --------------------------------------------------------------------

// sweepStatus is the per-instance state of the last sweep, served by the
// status endpoint.
type sweepStatus struct {
	State       string `json:"state"` // idle, running, done, failed
	Target      string `json:"target,omitempty"`
	Full        bool   `json:"full"`
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

const (
	sweepBatch    = 500
	sweepLockWait = 10 * time.Second
)

// runEncryptionSweep re-encrypts stored values under the target key. A full
// sweep (a promotion) walks the whole tree and log, flips the registry's
// Active to the target — from then on every instance writes under it — and
// walks the tree once more for what was written under the old key while
// the first pass ran. The consistency pass (same key) walks the tree only
// and follows up the log rows of the nodes it had to fix.
func runEncryptionSweep(full bool, target string) {
	ctx := log.Logger.WithContext(context.Background())
	release, err := acquireAdvisoryLock(ctx, "encryption-sweep", sweepLockWait)
	if errors.Is(err, errAdvisoryLocked) {
		log.Info().Msg("another instance is re-encrypting stored values; skipping")
		return
	} else if err != nil {
		log.Error().Err(err).Msg("failed to acquire the encryption sweep lock")
		return
	}
	defer release()

	sweepMu.Lock()
	sweepState = sweepStatus{State: "running", Target: target, Full: full, StartedAt: time.Now().Format(time.RFC3339)}
	sweepMu.Unlock()

	scanned, reencrypted, err := encryptionSweep(ctx, full, target)
	if err == nil && full {
		if err = setRegisteredActive(ctx, target); err == nil {
			Keys.SetActive(target)
			var s, r int
			s, r, err = encryptionSweep(ctx, false, target)
			scanned, reencrypted = scanned+s, reencrypted+r
		}
	}

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
		log.Error().Err(err).Bool("full", full).Str("key", target).Msg("re-encryption of stored values failed")
		return
	}
	event := log.Info().Bool("full", full).Int("scanned", scanned).Int("reencrypted", reencrypted)
	if reencrypted > 0 && !full {
		event = log.Warn().Bool("full", full).Int("scanned", scanned).Int("reencrypted", reencrypted)
	}
	event.Str("key", target).Msg("re-encryption of stored values finished")
}

func encryptionSweep(ctx context.Context, full bool, target string) (scanned, reencrypted int, err error) {
	nodeIDs, err := reencryptTable(ctx, "my_config_tree", "", target, &scanned, &reencrypted)
	if err != nil {
		return
	}
	if full {
		_, err = reencryptTable(ctx, "my_config_tree_log", "", target, &scanned, &reencrypted)
		return
	}
	if len(nodeIDs) == 0 {
		return
	}
	// the consistency pass follows up only the fixed nodes' history
	marks := make([]string, len(nodeIDs))
	for i := range marks {
		marks[i] = fmt.Sprint(nodeIDs[i])
	}
	_, err = reencryptTable(ctx, "my_config_tree_log", "NodeID IN ("+strings.Join(marks, ", ")+")", target, &scanned, &reencrypted)
	return
}

type sweepRow struct {
	id          int64
	nodeID      int64
	contentType string
	encrypted   bool
	value       NullString
}

// reencryptTable walks a table in ID batches and rewrites every value that
// holds an envelope under a key other than the target, with a
// compare-and-swap so a concurrent write always wins. It returns the node
// IDs of the rows it rewrote.
func reencryptTable(ctx context.Context, table, extraWhere, target string, scanned, reencrypted *int) ([]int64, error) {
	nodeColumn := "ID"
	if table == "my_config_tree_log" {
		nodeColumn = "NodeID"
	}
	where := encryptedRowsWhere
	if extraWhere != "" {
		where += " AND " + extraWhere
	}
	var nodeIDs []int64
	seen := map[int64]bool{}
	lastID := int64(0)
	for {
		rows, err := DB.QueryContext(ctx, "SELECT ID, "+nodeColumn+", ContentType, Encrypted, Value FROM "+table+
			" WHERE ID > ? AND "+where+" ORDER BY ID LIMIT ?", lastID, sweepBatch)
		if err != nil {
			return nodeIDs, err
		}
		batch := make([]sweepRow, 0, sweepBatch)
		for rows.Next() {
			var r sweepRow
			if err := rows.Scan(&r.id, &r.nodeID, &r.contentType, &r.encrypted, &r.value); err != nil {
				rows.Close()
				return nodeIDs, err
			}
			batch = append(batch, r)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nodeIDs, err
		}
		rows.Close()
		if len(batch) == 0 {
			return nodeIDs, nil
		}
		for _, r := range batch {
			lastID = r.id
			*scanned++
			newValue, changed, err := ReencryptStoredValue(r.contentType, r.encrypted, r.value, target)
			if err != nil {
				return nodeIDs, fmt.Errorf("%s row %d: %w", table, r.id, err)
			}
			if !changed {
				continue
			}
			res, err := DB.ExecContext(ctx, "UPDATE "+table+" SET Value = ? WHERE ID = ? AND Value = ?", newValue, r.id, r.value)
			if err != nil {
				return nodeIDs, err
			}
			if n, _ := res.RowsAffected(); n == 1 {
				*reencrypted++
				if !seen[r.nodeID] {
					seen[r.nodeID] = true
					nodeIDs = append(nodeIDs, r.nodeID)
				}
			}
		}
	}
}

// --- status -------------------------------------------------------------------

type keyStatus struct {
	ID         string `json:"id"`
	Current    bool   `json:"current"`    // the key this instance encrypts with
	Registered bool   `json:"registered"` // present in the registry
	Verified   bool   `json:"verified"`   // material matches the registry
	Recorded   bool   `json:"recorded"`   // the registry's Active mark
	Seq        int64  `json:"seq,omitempty"`
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
	Sweep      sweepStatus      `json:"sweep"`
}

// SelectEncryptionStatus answers the operator's questions "did the rotation
// take?" and "may the old key go now?": the wanted key, the keyring's keys
// with their registry state, and the sweep state.
func SelectEncryptionStatus(ctx context.Context) (*EncryptionStatus, error) {
	status := &EncryptionStatus{Configured: Keys != nil, Keys: []keyStatus{}}
	rows, err := DB.QueryContext(ctx, "SELECT KeyID, Seq, Verifier, Active, CreatedAt FROM my_config_encryption_key ORDER BY Seq")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	registered := map[string]keyStatus{}
	var order []string
	for rows.Next() {
		var ks keyStatus
		var verifier []byte
		if err := rows.Scan(&ks.ID, &ks.Seq, &verifier, &ks.Recorded, &ks.CreatedAt); err != nil {
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
			ks, ok := registered[key.ID]
			if !ok {
				ks = keyStatus{ID: key.ID}
			}
			ks.Current = key.ID == Keys.Active().ID
			registered[key.ID] = ks
			if !ok {
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
