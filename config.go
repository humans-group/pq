package pq

import (
	"time"
)

type Config struct {
	Name               string
	ConnString         string
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
