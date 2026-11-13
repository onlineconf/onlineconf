//go:build !integration

package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestJudgeSchema(t *testing.T) {
	version := func(v int64) sql.NullInt64 { return sql.NullInt64{Int64: v, Valid: true} }
	cases := []struct {
		name    string
		version sql.NullInt64
		err     error
		fatal   string // substring, "" for no fatal
		warn    string // substring, "" for no warning
	}{
		{"missing table", sql.NullInt64{}, &mysql.MySQLError{Number: 1146}, "apply admin/etc/update-8.sql", ""},
		{"wrapped missing table", sql.NullInt64{}, fmt.Errorf("x: %w", &mysql.MySQLError{Number: 1146}), "update-8.sql", ""},
		{"empty table", sql.NullInt64{}, nil, "update-8.sql", ""},
		{"connectivity", sql.NullInt64{}, errors.New("connection refused"), "", "skipped"},
		{"behind by one", version(7), nil, "apply admin/etc/update-8.sql", ""},
		{"behind by many", version(5), nil, "update-6.sql through update-8.sql", ""},
		{"current", version(8), nil, "", ""},
		{"ahead", version(9), nil, "", "rolling upgrade"},
	}
	for _, c := range cases {
		got := judgeSchema(c.version, c.err)
		if (c.fatal == "") != (got.fatal == "") || !strings.Contains(got.fatal, c.fatal) {
			t.Errorf("%s: fatal = %q, want %q", c.name, got.fatal, c.fatal)
		}
		if (c.warn == "") != (got.warn == "") || !strings.Contains(got.warn, c.warn) {
			t.Errorf("%s: warn = %q, want %q", c.name, got.warn, c.warn)
		}
	}
	if isMissingTable(&mysql.MySQLError{Number: 1054}) || isMissingTable(errors.New("other")) {
		t.Errorf("unrelated error taken for a missing table")
	}
}
