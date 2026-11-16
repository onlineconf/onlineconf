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

// checkSchema turns the result of the version query into what to do: a
// fatal message stops the server, a warning is logged, neither means the
// schema is current. The table appeared with version 8, so a database
// without it (or with no row in it: a half-applied update) is behind. A
// version newer than the binary's is expected during a rolling upgrade —
// the migration is applied before the new binaries roll out — and is only
// reported. A connectivity problem is not judged here: the caller retries.
func checkSchema(version sql.NullInt64, err error) (fatal, warn string) {
	switch {
	case isMissingTable(err) || (err == nil && !version.Valid):
		return fmt.Sprintf("the database schema is behind this version: apply admin/etc/update-%d.sql (and any earlier update-*.sql not applied yet)", schemaVersion), ""
	case err != nil:
		return "", "schema version check failed, will retry: " + err.Error()
	case version.Int64+1 == schemaVersion:
		return fmt.Sprintf("the database schema is at version %d, this version needs %d: apply admin/etc/update-%d.sql", version.Int64, schemaVersion, schemaVersion), ""
	case version.Int64 < schemaVersion:
		return fmt.Sprintf("the database schema is at version %d, this version needs %d: apply admin/etc/update-%d.sql through update-%d.sql", version.Int64, schemaVersion, version.Int64+1, schemaVersion), ""
	case version.Int64 > schemaVersion:
		return "", fmt.Sprintf("the database schema is at version %d, newer than the %d this version is written for: fine during a rolling upgrade", version.Int64, schemaVersion)
	}
	return "", ""
}

// ensureSchemaCurrent aborts startup with an actionable message when the
// database schema is behind the one this version is written against — the
// upgrade mistake that otherwise surfaces as an endless stream of
// unknown-column errors on every request. It reports whether it could
// judge: against a database that cannot be reached the caller retries,
// since a verdict missed at a cold start would never be given.
func ensureSchemaCurrent() bool {
	var version sql.NullInt64
	err := DB.QueryRow("SELECT MAX(Version) FROM my_config_schema").Scan(&version)
	switch fatal, warn := checkSchema(version, err); {
	case fatal != "":
		log.Fatal().Msg(fatal)
	case warn != "":
		log.Warn().Msg(warn)
		return err == nil
	default:
		log.Info().Int64("version", version.Int64).Msg("database schema is current")
	}
	return true
}

// isMissingTable reports MySQL error 1146 (no such table), wrapped or not.
func isMissingTable(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1146
}
