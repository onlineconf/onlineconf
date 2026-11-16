//go:build integration

package integration

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestSchemaVersionGuard restarts the admin against a schema version it must
// refuse (7: behind) and one it must only warn about (9: ahead), reads the
// outcome from the container's log, and restores version 8. Each case
// restarts the admin container.
func TestSchemaVersionGuard(t *testing.T) {
	setVersion := func(v int) {
		dbQuery(t, fmt.Sprintf("UPDATE my_config_schema SET Version = %d", v))
	}
	defer func() {
		setVersion(8)
		restartAdmin(t)
		if err := waitReady(2 * time.Minute); err != nil {
			t.Fatal(err)
		}
	}()

	setVersion(7)
	since := time.Now()
	restartAdmin(t)
	if !adminLogsSince(t, since, "apply admin/etc/update-8.sql", 60*time.Second) {
		t.Errorf("version 7: the admin did not name update-8.sql")
	}

	setVersion(9)
	since = time.Now()
	restartAdmin(t)
	if err := waitReady(2 * time.Minute); err != nil {
		t.Fatalf("version 9: the admin did not serve: %v", err)
	}
	if !adminLogsSince(t, since, "newer than the 8", 30*time.Second) {
		t.Errorf("version 9: no warning logged")
	}
}

// TestColdStartLog reads what the suite's cold start logged: the database
// is not reachable at the admin's first attempt (the MySQL container's
// healthcheck passes before it listens on the network), so the startup
// checks must retry until they judge the schema.
func TestColdStartLog(t *testing.T) {
	if !strings.Contains(coldStartLog, `"message":"encryption key registered as active"`) {
		t.Skip("not a cold start on a fresh database: the log shows no registration")
	}
	if !strings.Contains(coldStartLog, "database schema is current") {
		t.Errorf("the cold start never judged the schema:\n%s", coldStartLog)
	}
	// the first registration happens once — the startup's trigger, not a
	// tree load racing it — and the pass after it verifies the fresh data
	if n := strings.Count(coldStartLog, `"message":"encryption key registered as active"`); n != 1 {
		t.Errorf("the key was registered as active %d times, want once:\n%s", n, coldStartLog)
	}
	if n := strings.Count(coldStartLog, "re-encryption of stored values finished"); n != 1 {
		t.Errorf("%d sweeps finished at the cold start, want the one after the registration:\n%s", n, coldStartLog)
	}
	if strings.Contains(coldStartLog, "re-encryption of stored values failed") {
		t.Errorf("a sweep failed at the cold start:\n%s", coldStartLog)
	}
}

func restartAdmin(t *testing.T) {
	t.Helper()
	if out, err := compose("restart", "onlineconf-admin"); err != nil {
		t.Fatalf("compose restart: %v\n%s", err, out)
	}
}

// adminLogsSince is adminLogsContain restricted to the log written after
// the given moment, so that an earlier run of the same container cannot
// satisfy the search.
func adminLogsSince(t *testing.T, since time.Time, needle string, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		logs, _ := compose("logs", "--since", since.UTC().Format(time.RFC3339), "onlineconf-admin")
		if strings.Contains(logs, needle) {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(2 * time.Second)
	}
}
