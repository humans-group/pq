package pq

import (
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/log/zapadapter"
	"go.uber.org/zap"
)

type Config struct {
	Name               string
	ConnString         string
	LogLevel           string
	Logger             *zap.Logger
	Tracing            bool
	Metrics            bool
	MaxConnections     int
	TCPKeepAlivePeriod time.Duration
	AcquireTimeout     time.Duration
}

func (cfg Config) pgxCfg() pgx.ConnConfig {
	c, err := pgx.ParseConnectionString(cfg.ConnString)
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
	c.Dial = dialer.Dial

	if cfg.Logger != nil {
		c.Logger = zapadapter.NewLogger(cfg.Logger)
	}

	return c
}
