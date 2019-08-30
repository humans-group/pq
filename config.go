package pg

import (
	"fmt"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/log/zapadapter"
	"net/url"
	"strconv"
)

type Config struct {
	MaxConnections int
	Name       string
	ConnString string
	LogLevel   string
	Logger     *zap.Logger
	Tracing    bool
	Metrics    bool
}

func (cfg Config) pgxCfg() pgx.ConnConfig {
	u, err := url.Parse(cfg.ConnString)
	if err != nil {
		panic(fmt.Sprintf("failed to parse postgres conn string %s: %v", cfg.ConnString, err))
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		panic(fmt.Sprintf("failed to parse postgres port from conn string %s: %v", cfg.ConnString, err))
	}

	values, err := url.ParseQuery(cfg.ConnString)
	username := values["username"]
	if len(username) != 1 {
		panic(fmt.Sprintf("invalid user name %+v from postgres conn string %s", username, cfg.ConnString))
	}

	password := values["password"]
	if len(password) != 1 {
		panic(fmt.Sprintf("invalid password %+v from postgres conn string %s", password, cfg.ConnString))
	}

	pgxConfig := pgx.ConnConfig{
		Host:     u.Host,
		Port:     uint16(port),
		Database: u.Path,
		User:     username[0],
		Password: password[0],
	}

	if cfg.Logger != nil {
		pgxConfig.Logger = zapadapter.NewLogger(cfg.Logger)
	}

	return pgxConfig
}
