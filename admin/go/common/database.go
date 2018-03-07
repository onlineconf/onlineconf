package common

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
	"net"
	"time"
)

var DB = OpenDatabase(AdminConfig.Database)

// github.com/go-sql-driver/mysql v1.3 compatibility
func mysqlNewConfig() *mysql.Config {
	return &mysql.Config{
		Collation:            "utf8_general_ci",
		Loc:                  time.UTC,
		MaxAllowedPacket:     4 << 20,
		AllowNativePasswords: true,
	}
}

func mysqlInitConfig(config DatabaseConfig) *mysql.Config {
	mysqlConfig := mysqlNewConfig()
	mysqlConfig.User = config.User
	mysqlConfig.Passwd = config.Password
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = net.JoinHostPort(config.Host, "3306")
	mysqlConfig.DBName = config.Base
	mysqlConfig.Params = make(map[string]string)
	return mysqlConfig
}

func OpenDatabase(config DatabaseConfig) *sql.DB {
	mysqlConfig := mysqlInitConfig(config)
	mysqlConfig.Params["charset"] = "utf8mb4"
	mysqlConfig.Params["collation"] = "utf8mb4_general_ci"
	db, err := sql.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}
	return db
}

func ReadStrings(rows *sql.Rows) ([]string, error) {
	list := make([]string, 0)
	defer rows.Close()
	for rows.Next() {
		var str string
		err := rows.Scan(&str)
		if err != nil {
			return nil, err
		}
		list = append(list, str)
	}
	return list, nil
}
