//go:build integration

package integration

import (
	"fmt"
	"testing"

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
