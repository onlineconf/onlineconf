package admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/nyarla/go-crypt"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

type MySQLAuthenticatorConfig struct {
	DatabaseConfig `yaml:",inline"`
	Table          string
	NameField      string `yaml:"name_field"`
	PasswordField  string `yaml:"password_field"`
	Condition      string
}

type mysqlAuthenticator struct {
	config MySQLAuthenticatorConfig
	db     *sql.DB
}

func NewMySQLAuthenticator(config MySQLAuthenticatorConfig) *mysqlAuthenticator {
	return &mysqlAuthenticator{
		config: config,
		db:     openAuthDatabase(config.DatabaseConfig),
	}
}

func openAuthDatabase(config DatabaseConfig) *sql.DB {
	mysqlConfig := MysqlInitConfig(config)
	mysqlConfig.Params["allowOldPasswords"] = "1"
	db, err := sql.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open authentication database")
	}
	db.SetConnMaxLifetime(time.Duration(config.MaxLifetime) * time.Second)
	return db
}

func (auth *mysqlAuthenticator) Authenticate(req *http.Request) (string, error) {
	if username, password, ok := req.BasicAuth(); ok {
		ok, err := auth.checkUserPassword(req.Context(), username, password)
		if err != nil {
			return "", err
		}
		if ok {
			return username, nil
		}
	}
	return "", nil
}

func (auth *mysqlAuthenticator) checkUserPassword(ctx context.Context, user, password string) (bool, error) {
	query := fmt.Sprintf("SELECT `%s` FROM `%s` WHERE `%s` = ?", auth.config.PasswordField, auth.config.Table, auth.config.NameField)
	if auth.config.Condition != "" {
		query += " AND " + auth.config.Condition
	}
	row := auth.db.QueryRowContext(ctx, query, user)
	var pwd string
	switch err := row.Scan(&pwd); err {
	case sql.ErrNoRows:
		return false, nil
	case nil:
		return crypt.Crypt(password, pwd) == pwd, nil
	default:
		return false, err
	}
}

var termRe = regexp.MustCompile("([%_])")

func (auth *mysqlAuthenticator) SelectUsers(ctx context.Context, term string) ([]string, error) {
	order := "1"
	condition := make([]string, 0)
	bind := make([]interface{}, 0)
	if auth.config.Condition != "" {
		condition = append(condition, auth.config.Condition)
	}
	if term != "" {
		condition = append(condition, fmt.Sprintf("`%s` LIKE ?", auth.config.NameField))
		bind = append(bind, "%"+termRe.ReplaceAllString(term, "\\$1")+"%", term)
		order = fmt.Sprintf("INSTR(`%s`, ?), 1", auth.config.NameField)
	}
	query := fmt.Sprintf("SELECT `%s` FROM `%s`", auth.config.NameField, auth.config.Table)
	if len(condition) > 0 {
		query += " WHERE " + strings.Join(condition, " AND ")
	}
	query += " ORDER BY " + order
	rows, err := auth.db.QueryContext(ctx, query, bind...)
	if err != nil {
		return nil, err
	}
	return ReadStrings(rows)
}
