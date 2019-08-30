package pg

import (
	"context"
	"fmt"
	"github.com/jackc/pgx"
)

type PGXAdapter struct {
	pool *pgx.ConnPool
}

func (p *PGXAdapter) Exec(ctx context.Context, sql string, args ...interface{}) (result RowsAffected, err error) {
	return p.pool.ExecEx(ctx, sql, nil, args)
}

func (p *PGXAdapter) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	return p.pool.QueryEx(ctx, sql, nil, args)
}

func (p *PGXAdapter) QueryRow(ctx context.Context, sql string, args ...interface{}) Row {
	return p.pool.QueryRowEx(ctx, sql, nil, args)
}

func (p PGXAdapter) SetLogLevel(lvl int) error {
	panic("implement me")
}

func NewClient(cfg Config) Client {
	connPool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     cfg.pgxCfg(),
		MaxConnections: cfg.MaxConnections,
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
