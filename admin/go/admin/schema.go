package admin

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

// schemaVersion is the database schema this binary is written against: the
// number of the last admin/etc/update-N.sql it needs applied. Every update
// file records itself in my_config_schema as its last statement, and
// schema.sql records the version it creates.
const schemaVersion = 8

// schemaVerdict is what the startup check concluded: a fatal message stops
// the server, a warning is logged, neither means the schema is current.
type schemaVerdict struct {
	fatal, warn string
}

// judgeSchema turns the result of the version query into a verdict. The
// table appeared with version 8, so a database without it (or with no row
// in it: a half-applied update) is behind. A version newer than the
// binary's is expected during a rolling upgrade — the migration is applied
// before the new binaries roll out — and is only reported. Connectivity
// problems are deliberately not judged here.
func judgeSchema(version sql.NullInt64, err error) schemaVerdict {
	switch {
	case isMissingTable(err) || (err == nil && !version.Valid):
		return schemaVerdict{fatal: fmt.Sprintf("the database schema is behind this version: apply admin/etc/update-%d.sql (and any earlier update-*.sql not applied yet)", schemaVersion)}
	case err != nil:
		return schemaVerdict{warn: "schema version check skipped: " + err.Error()}
	case version.Int64 < schemaVersion:
		return schemaVerdict{fatal: fmt.Sprintf("the database schema is at version %d, this version needs %d: apply %s", version.Int64, schemaVersion, updatesToApply(version.Int64+1, schemaVersion))}
	case version.Int64 > schemaVersion:
		return schemaVerdict{warn: fmt.Sprintf("the database schema is at version %d, newer than the %d this version is written for: fine during a rolling upgrade", version.Int64, schemaVersion)}
	}
	return schemaVerdict{}
}

func updatesToApply(from, to int64) string {
	if from == to {
		return fmt.Sprintf("admin/etc/update-%d.sql", to)
	}
	return fmt.Sprintf("admin/etc/update-%d.sql through update-%d.sql", from, to)
}

// ensureSchemaCurrent aborts startup with an actionable message when the
// database schema is behind the one this version is written against — the
// upgrade mistake that otherwise surfaces as an endless stream of
// unknown-column errors on every request.
func ensureSchemaCurrent() {
	var version sql.NullInt64
	err := DB.QueryRow("SELECT MAX(Version) FROM my_config_schema").Scan(&version)
	verdict := judgeSchema(version, err)
	switch {
	case verdict.fatal != "":
		log.Fatal().Msg(verdict.fatal)
	case verdict.warn != "":
		log.Warn().Msg(verdict.warn)
	default:
		log.Info().Int64("version", version.Int64).Msg("database schema is current")
	}
}

// isMissingTable reports MySQL error 1146 (no such table), wrapped or not.
func isMissingTable(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1146
}
