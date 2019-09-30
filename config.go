package pq

import (
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"go.uber.org/zap"
)

type Config struct {
	Name               string
	ConnString         string
	LogLevel           string
	Logger             *zap.Logger `key:"-"`
	Tracing            bool
	Metrics            bool
	MaxConnections     int32
	TCPKeepAlivePeriod time.Duration
	AcquireTimeout     time.Duration
}

func (c Config) withDefaults() Config {
	if c.AcquireTimeout == 0 {
		c.AcquireTimeout = time.Second
	}

	if c.MaxConnections == 0 {
		c.MaxConnections = 4
	}

	return c
}

func (cfg Config) pgxCfg() *pgx.ConnConfig {
	c, err := pgx.ParseConfig(cfg.ConnString)
	if err != nil {
		panic(fmt.Sprintf("failed to parce conn string %s: %v", cfg.ConnString, err))
	}

	if cfg.TCPKeepAlivePeriod == 0 {
		cfg.TCPKeepAlivePeriod = 5 * time.Minute // that's default value used by pgx internally
	}
	dialer := &net.Dialer{
		Timeout:   cfg.AcquireTimeout,
		KeepAlive: cfg.TCPKeepAlivePeriod,
	}
	c.Config = pgconn.Config{DialFunc: dialer.DialContext}

	if cfg.Logger != nil {
		c.Logger = zapadapter.NewLogger(cfg.Logger)
	}

	return c
}
