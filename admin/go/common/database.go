package common

import (
	"database/sql"
	"net"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

var DB *sql.DB

func MysqlInitConfig(config DatabaseConfig) *mysql.Config {
	mysqlConfig := mysql.NewConfig()
	mysqlConfig.Collation = "utf8mb4_general_ci"
	mysqlConfig.User = config.User
	mysqlConfig.Passwd = config.Password
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = net.JoinHostPort(config.Host, "3306")
	mysqlConfig.DBName = config.Base
	return mysqlConfig
}

func OpenDatabase(config DatabaseConfig) *sql.DB {
	mysqlConfig := MysqlInitConfig(config)
	db, err := sql.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}
	db.SetConnMaxLifetime(time.Duration(config.MaxLifetime) * time.Second)
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
