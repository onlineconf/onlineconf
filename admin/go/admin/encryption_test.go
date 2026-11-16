//go:build !integration

package admin

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"

	. "github.com/onlineconf/onlineconf/admin/go/common"
	"github.com/onlineconf/onlineconf/admin/go/resolver"
)

func nullBool(valid, value bool) NullBool {
	var b NullBool
	b.Valid, b.Bool = valid, value
	return b
}

func TestCanReveal(t *testing.T) {
	noAccess, readOnly, readWrite := nullBool(false, false), nullBool(true, false), nullBool(true, true)
	cases := []struct {
		requireWrite bool
		rw           NullBool
		want         bool
	}{
		{true, noAccess, false},
		{true, readOnly, false},
		{true, readWrite, true},
		{false, noAccess, false},
		{false, readOnly, true},
		{false, readWrite, true},
	}
	for _, c := range cases {
		if got := canReveal(c.requireWrite, c.rw); got != c.want {
			t.Errorf("canReveal(requireWrite=%v, rw=%+v) = %v, want %v", c.requireWrite, c.rw.NullBool, got, c.want)
		}
	}
}

func TestValidateParameterEncryptionEligibility(t *testing.T) {
	for _, mime := range []string{"application/x-null", "application/x-symlink", ContentTypeCase, ContentTypeTemplate} {
		if err := validateParameter(nil, mime, true, NullStringOf("x")); !errors.Is(err, ErrEncryptionNotAllowed) {
			t.Errorf("%s with the encrypted flag: err = %v, want ErrEncryptionNotAllowed", mime, err)
		}
	}
	for _, mime := range []string{"text/plain", "application/x-list", "application/x-server"} {
		if err := validateParameter(nil, mime, true, NullStringOf("x")); err != nil {
			t.Errorf("%s with the encrypted flag: err = %v", mime, err)
		}
	}
	if err := validateParameter(nil, ContentTypeTemplate, false, NullStringOf("${/x}")); err != nil {
		t.Errorf("plaintext template: %v", err)
	}
	if err := validateParameter(nil, "application/json", true, NullStringOf(`{"a": 1}`)); err != nil {
		t.Errorf("encrypted json: %v", err)
	}
	if err := validateParameter(nil, "application/json", true, NullStringOf(`{broken`)); err == nil {
		t.Errorf("broken encrypted json accepted")
	}
	// a case branch is never encrypted; a stray marker on a branch is an
	// unknown key, ignored like any other
	marked := `[{"server":"a*","mime":"text/plain","encrypted":true,"value":"s"},{"mime":"application/json","value":"1"}]`
	if err := validateParameter(nil, ContentTypeCase, false, NullStringOf(marked)); err != nil {
		t.Errorf("branch with a stray marker: %v", err)
	}
	plain := `[{"server":"a*","mime":"text/plain","value":"s"},{"mime":"application/json","value":"1"}]`
	if err := validateParameter(nil, ContentTypeCase, false, NullStringOf(plain)); err != nil {
		t.Errorf("plain case: %v", err)
	}
	// a branch without a value is refused by the non-null rule of its type
	if err := validateParameter(nil, ContentTypeCase, false, NullStringOf(`[{"mime":"text/plain","value":null}]`)); !errors.Is(err, ErrInvalidValue) {
		t.Errorf("null text branch: err = %v, want ErrInvalidValue", err)
	}
}

// A syntax error in a JSON or YAML value is the bare invalid-value error,
// encrypted or not: the decoder's text quotes the value.
func TestValidateParameterDecoderErrors(t *testing.T) {
	for _, c := range []struct{ mime, value string }{
		{"application/json", `{"zq":`},
		{"application/x-yaml", "zq: [\n"},
	} {
		for _, encrypted := range []bool{true, false} {
			err := validateParameter(nil, c.mime, encrypted, NullStringOf(c.value))
			if !errors.Is(err, ErrInvalidValue) || err.Error() != ErrInvalidValue.Error() {
				t.Errorf("%s encrypted=%v: err = %v, want the bare ErrInvalidValue", c.mime, encrypted, err)
			}
		}
	}
}

func TestEncryptedField(t *testing.T) {
	for _, c := range []struct {
		form  url.Values
		valid bool
		value bool
		err   error
	}{
		{url.Values{}, false, false, nil},
		{url.Values{"encrypted": {"0"}}, true, false, nil},
		{url.Values{"encrypted": {"1"}}, true, true, nil},
		{url.Values{"encrypted": {"true"}}, false, false, errInvalidEncryptedField},
		{url.Values{"encrypted": {"on"}}, false, false, errInvalidEncryptedField},
		{url.Values{"encrypted": {""}}, false, false, errInvalidEncryptedField},
	} {
		got, err := encryptedField(c.form)
		if !errors.Is(err, c.err) || (err == nil && (got.Valid != c.valid || got.Bool != c.value)) {
			t.Errorf("encryptedField(%v) = %+v, %v; want valid=%v value=%v err=%v", c.form, got.NullBool, err, c.valid, c.value, c.err)
		}
	}
}

// Without a keyring a stored secret cannot be opened: a deleted secret's
// restore and a sealed version's reveal fail with that error wrapped as
// DecryptionFailed, while the resolver, which loads live rows only, serves.
func TestDecryptStoredWithoutKeyring(t *testing.T) {
	withKeys(t, nil)
	if _, err := decryptStored(true, NullStringOf("oc1:demo:AAAA")); !errors.Is(err, ErrNoEncryptionKey) {
		t.Errorf("decrypt without a keyring: err = %v, want ErrNoEncryptionKey", err)
	}
	if got, err := decryptStored(false, NullStringOf("plain")); err != nil || got.String != "plain" {
		t.Errorf("plaintext passthrough: %v, %v", got, err)
	}
}

// withKeys installs a keyring for the test's duration.
func withKeys(t *testing.T, kr *Keyring) {
	t.Helper()
	prev := Keys
	Keys = kr
	t.Cleanup(func() { Keys = prev })
}

// Without a keyring the feature is off whatever the switch parameter says:
// an encrypted write is refused before the database is consulted.
func TestEncryptForStorageGate(t *testing.T) {
	withKeys(t, nil)
	if _, _, err := encryptForStorage(context.Background(), nil, true, NullStringOf("s")); !errors.Is(err, ErrEncryptionDisabled) {
		t.Errorf("encrypted write without a keyring: err = %v, want ErrEncryptionDisabled", err)
	}
	// a plaintext value passes through untouched, null included
	for _, v := range []NullString{NullStringOf("p"), {}} {
		if got, _, err := encryptForStorage(context.Background(), nil, false, v); err != nil || got != v {
			t.Errorf("plaintext passthrough of %v: %v, %v", v, got, err)
		}
	}
}

func TestPlanActiveKey(t *testing.T) {
	for _, c := range []struct {
		name            string
		wanted, active  string
		hasActive, held bool
		want            keyAction
	}{
		{"fresh installation", "k1", "", false, true, keyActionRegister},
		{"fresh installation, key not held", "k1", "", false, false, keyActionFatal},
		{"the key in use", "k1", "k1", true, true, keyActionNone},
		{"a key not held", "k2", "k1", true, false, keyActionRefuse},
		{"a held key", "k2", "k1", true, true, keyActionPromote},
		{"a retired key again", "k0", "k1", true, true, keyActionPromote},
	} {
		if got, _ := planActiveKey(c.wanted, "parameter", c.active, c.hasActive, c.held); got != c.want {
			t.Errorf("%s: action = %v, want %v", c.name, got, c.want)
		}
	}
}

// The watcher compares what the reader of the key id parameter would see:
// a plain text, anything else absent. (Without a keyring nothing is
// triggered, which keeps the test off the database.)
func TestKeyIDWatcherValue(t *testing.T) {
	withKeys(t, nil)
	node := func(contentType string, encrypted bool, value string) *resolver.Param {
		return &resolver.Param{ContentType: contentType, Encrypted: encrypted, Value: NullStringOf(value)}
	}
	w := &keyIDWatcher{}
	for _, c := range []struct {
		name  string
		param *resolver.Param
		want  string
	}{
		{"text", node(ContentTypeText, false, "k1"), "k1"},
		{"absent", nil, ""},
		{"encrypted text", node(ContentTypeText, true, "k1"), ""},
		{"a case resolved to absent", node(ContentTypeNull, false, ""), ""},
		{"json", node(ContentTypeJSON, false, `"k1"`), ""},
		{"text again", node(ContentTypeText, false, "k2"), "k2"},
	} {
		w.Update(context.Background(), c.param)
		if !w.seen || w.last != c.want {
			t.Errorf("%s: last = %q, want %q", c.name, w.last, c.want)
		}
	}
	w.prime("k3")
	if !w.seen || w.last != "k3" {
		t.Errorf("primed: last = %q", w.last)
	}
}

func TestIsRetryableLockError(t *testing.T) {
	for number, want := range map[uint16]bool{1213: true, 1205: true, 1062: false} {
		if got := isRetryableLockError(&mysql.MySQLError{Number: number}); got != want {
			t.Errorf("%d: %v, want %v", number, got, want)
		}
	}
	if isRetryableLockError(errors.New("other")) || isRetryableLockError(fmt.Errorf("x: %w", &mysql.MySQLError{Number: 1062})) {
		t.Errorf("an unrelated error taken for a lock conflict")
	}
	if !isRetryableLockError(fmt.Errorf("x: %w", &mysql.MySQLError{Number: 1213})) {
		t.Errorf("a wrapped deadlock missed")
	}
}

func TestWithLockRetry(t *testing.T) {
	prev := lockRetryPause
	lockRetryPause = time.Millisecond
	t.Cleanup(func() { lockRetryPause = prev })
	ctx := context.Background()
	calls := 0
	err := withLockRetry(ctx, "op", func() error {
		calls++
		if calls < 3 {
			return &mysql.MySQLError{Number: 1213}
		}
		return nil
	})
	if err != nil || calls != 3 {
		t.Errorf("two deadlocks then success: err = %v, calls = %d", err, calls)
	}
	calls = 0
	duplicate := &mysql.MySQLError{Number: 1062}
	if err := withLockRetry(ctx, "op", func() error { calls++; return duplicate }); err != duplicate || calls != 1 {
		t.Errorf("a duplicate key: err = %v, calls = %d", err, calls)
	}
	calls = 0
	timeout := &mysql.MySQLError{Number: 1205}
	if err := withLockRetry(ctx, "op", func() error { calls++; return timeout }); err != timeout || calls != lockRetryAttempts {
		t.Errorf("a lasting lock wait: err = %v, calls = %d", err, calls)
	}
}

// withSweepStub replaces the sweep run for a test and resets the goroutine's
// bookkeeping afterwards.
func withSweepStub(t *testing.T, run func() bool) {
	t.Helper()
	kr, err := ParseKeyring(strings.NewReader("k1 " + strings.Repeat("A", 43) + "=\n"))
	if err != nil {
		t.Fatal(err)
	}
	withKeys(t, kr)
	prev := runSweep
	runSweep = run
	t.Cleanup(func() {
		runSweep = prev
		sweepMu.Lock()
		if sweepRetry != nil {
			sweepRetry.Stop()
		}
		sweepArmed, sweepRunning, sweepRerun, sweepRetry, sweepRetryDelay = false, false, false, nil, sweepRetryMin
		sweepMu.Unlock()
	})
}

func sweepIsRunning() bool {
	sweepMu.Lock()
	defer sweepMu.Unlock()
	return sweepRunning
}

func awaitSweepIdle(t *testing.T) {
	t.Helper()
	for i := 0; i < 500 && sweepIsRunning(); i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if sweepIsRunning() {
		t.Fatalf("the sweep goroutine did not end")
	}
}

// The trigger does nothing before the startup armed it; while the sweep
// runs, triggers mark it to run once more, not once per trigger.
func TestSweepTrigger(t *testing.T) {
	runs, release := make(chan struct{}), make(chan struct{})
	count := 0
	withSweepStub(t, func() bool { count++; runs <- struct{}{}; <-release; return true })
	triggerKeyReconcile()
	if sweepIsRunning() {
		t.Fatalf("a trigger before the startup started the sweep")
	}
	armSweep()
	triggerKeyReconcile()
	<-runs
	triggerKeyReconcile()
	triggerKeyReconcile()
	release <- struct{}{}
	<-runs
	release <- struct{}{}
	awaitSweepIdle(t)
	if count != 2 {
		t.Errorf("the sweep ran %d times, want 2: once, and once more for the triggers that arrived meanwhile", count)
	}
}

// A failed sweep is scheduled again with a growing delay; a trigger
// arriving meanwhile cancels the timer and runs it at once.
func TestSweepRetryAfterFailure(t *testing.T) {
	count := 0
	withSweepStub(t, func() bool { count++; return false })
	armSweep()
	triggerKeyReconcile()
	awaitSweepIdle(t)
	sweepMu.Lock()
	timer, delay := sweepRetry, sweepRetryDelay
	sweepMu.Unlock()
	if timer == nil || delay != 2*sweepRetryMin {
		t.Errorf("after a failure: timer %v, next delay %v", timer != nil, delay)
	}
	triggerKeyReconcile()
	awaitSweepIdle(t)
	sweepMu.Lock()
	delay = sweepRetryDelay
	sweepMu.Unlock()
	if count != 2 || delay != 4*sweepRetryMin {
		t.Errorf("after a trigger during the retry delay: runs %d, next delay %v", count, delay)
	}
}
