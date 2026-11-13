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
