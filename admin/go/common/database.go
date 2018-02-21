package common

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

var DB = OpenDatabase(AdminConfig.Database)

func OpenDatabase(config DatabaseConfig) *sql.DB {
	mysqlConfig := mysql.NewConfig()
	mysqlConfig.User = config.User
	mysqlConfig.Passwd = config.Password
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = config.Host
	mysqlConfig.DBName = config.Base
	mysqlConfig.Params = make(map[string]string)
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
