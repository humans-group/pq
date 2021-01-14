package pq

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PgxAdapter struct {
	pool        *pgxpool.Pool
	withMetrics bool
	withTracing bool
	name        string
}

var _ Client = (*PgxAdapter)(nil)

func (p *PgxAdapter) Transaction(ctx context.Context, f func(context.Context, Executor) error) error {
	tx, err := p.pool.BeginTx(ctx, defaultTxOptions)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	var txAdapter Executor = &PgxTxAdapter{tx}
	if p.withTracing {
		txAdapter = &tracingAdapter{Executor: txAdapter}
	}
	if p.withMetrics {
		txAdapter = &metricsAdapter{Executor: txAdapter, name: p.name}
	}

	if txErr := f(ctx, txAdapter); txErr != nil {
		txErr = fmt.Errorf("exec transaction: %w", txErr)

		if err := tx.Rollback(ctx); err != nil {
			txErr = fmt.Errorf("%w: rollback: %v", txErr, err)
		}

		return txErr
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (p *PgxAdapter) Exec(ctx context.Context, sql string, args ...interface{}) (result RowsAffected, err error) {
	return p.pool.Exec(ctx, sql, args...)
}

func (p *PgxAdapter) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	return p.pool.Query(ctx, sql, args...)
}

func (p *PgxAdapter) QueryRow(ctx context.Context, sql string, args ...interface{}) Row {
	return p.pool.QueryRow(ctx, sql, args...)
}

func (p *PgxAdapter) SendBatch(ctx context.Context, batch *pgx.Batch) BatchResults {
	res := p.pool.SendBatch(ctx, batch)
	return batchResultsAdapter{BatchResults: res}
}

func (p PgxAdapter) SetLogLevel(lvl int) error {
	panic("implement me")
}

func NewClient(ctx context.Context, cfg Config) Client {
	cfg = cfg.withDefaults()

	poolCfg, err := pgxpool.ParseConfig(cfg.ConnString)
	if err != nil {
		panic(fmt.Errorf("parse config: %v", err))
	}

	if cfg.TCPKeepAlivePeriod == 0 {
		cfg.TCPKeepAlivePeriod = 5 * time.Minute // that's default value used by pgx internally
	}
	dialer := &net.Dialer{
		Timeout:   cfg.AcquireTimeout,
		KeepAlive: cfg.TCPKeepAlivePeriod,
	}

	poolCfg.ConnConfig.DialFunc = dialer.DialContext
	poolCfg.MaxConns = cfg.MaxConnections
	if cfg.Logger != nil {
		poolCfg.ConnConfig.Logger = newLoggerAdapter(cfg.Logger)
	}
	poolCfg.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		return !conn.IsClosed()
	}

	connPool, err := pgxpool.ConnectConfig(ctx, poolCfg)
	if err != nil {
		securedConnStr := strings.Replace(cfg.ConnString, poolCfg.ConnConfig.Password, "*****", 1)
		panic(fmt.Errorf("connect to postgres %q: %v", securedConnStr, err))
	}

	if err := collector.register(cfg.Name, connPool); err != nil {
		panic(fmt.Errorf("register dbx pool %q: %v", cfg.Name, err))
	}

	var adapter Client = &PgxAdapter{
		pool:        connPool,
		withMetrics: cfg.Metrics,
		withTracing: cfg.Tracing,
		name:        cfg.Name,
	}

	if cfg.Tracing {
		adapter = &tracingAdapter{Executor: adapter, Transactor: adapter}
	}

	if cfg.Metrics {
		adapter = &metricsAdapter{Executor: adapter, Transactor: adapter, name: cfg.Name}
	}

	return adapter
}


type batchResultsAdapter struct {
	pgx.BatchResults
}

func (b batchResultsAdapter) Exec() (RowsAffected, error) {
	return b.BatchResults.Exec()
}

func (b batchResultsAdapter) Query() (Rows, error) {
	return b.BatchResults.Query()
}

func (b batchResultsAdapter) QueryRow() Row {
	return b.BatchResults.QueryRow()
}