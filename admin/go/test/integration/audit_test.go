//go:build integration

package integration

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/onlineconf/onlineconf/admin/go/admin"
)

// TestAuditKinds keeps the Go kind constants and the my_config_audit.Kind
// enum in step: a constant added without its enum member (its update-N.sql
// not written) does not survive the round trip through the table.
func TestAuditKinds(t *testing.T) {
	kinds := []string{admin.AuditRevealView, admin.AuditRevealHistory}
	p := uniqPath("audit-kinds")
	mkParam(t, p, "text/plain", "x", "mk")
	nodeID := dbInt(t, "SELECT ID FROM my_config_tree WHERE Path = "+sqlQuote(p))
	for _, kind := range kinds {
		dbQuery(t, fmt.Sprintf("INSERT INTO my_config_audit (NodeID, Path, User, Kind) VALUES (%d, %s, 'it', %s)", nodeID, sqlQuote(p), sqlQuote(kind)))
	}
	rows := dbQuery(t, "SELECT Kind FROM my_config_audit WHERE Path = "+sqlQuote(p)+" ORDER BY ID")
	if len(rows) != len(kinds) {
		t.Fatalf("audit rows = %v, want %d", rows, len(kinds))
	}
	for i, kind := range kinds {
		if rows[i][0] != kind {
			t.Errorf("kind %q stored as %q: the enum lacks it", kind, rows[i][0])
		}
	}
}

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
