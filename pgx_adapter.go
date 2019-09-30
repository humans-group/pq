package pq

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PGXAdapter struct {
	pool *pgxpool.Pool
}

func (p *PGXAdapter) Exec(ctx context.Context, sql string, args ...interface{}) (result RowsAffected, err error) {
	return p.pool.Exec(ctx, sql, nil, args)
}

func (p *PGXAdapter) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	return p.pool.Query(ctx, sql, nil, args)
}

func (p *PGXAdapter) QueryRow(ctx context.Context, sql string, args ...interface{}) Row {
	return p.pool.QueryRow(ctx, sql, nil, args)
}

func (p PGXAdapter) SetLogLevel(lvl int) error {
	panic("implement me")
}

func NewClient(ctx context.Context, cfg Config) Client {
	cfg.withDefaults()

	connPool, err := pgxpool.ConnectConfig(ctx, &pgxpool.Config{
		ConnConfig: cfg.pgxCfg(),
		MaxConns:   cfg.MaxConnections,
		BeforeAcquire: func(ctx context.Context, conn *pgx.Conn) bool {
			return !conn.IsClosed()
		},
	})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to postgres %s: %v", cfg.ConnString, err))
	}

	if err := collector.register(cfg.Name, connPool); err != nil {
		panic(fmt.Sprintf("failed to register dbx pool %q: %v", cfg.Name, err))
	}

	var adapter Client = &PGXAdapter{connPool}

	if cfg.Tracing {
		adapter = &tracingAdapter{adapter}
	}

	if cfg.Metrics {
		adapter = &metricsAdapter{Client: adapter, name: cfg.Name}
	}

	return adapter
}
