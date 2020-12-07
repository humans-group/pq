package pq

import (
	"context"

	"github.com/jackc/pgx/v4"
)

var defaultTxOptions = pgx.TxOptions{
	IsoLevel:       pgx.ReadCommitted,
	AccessMode:     pgx.ReadWrite,
	DeferrableMode: pgx.NotDeferrable,
}

type PgxTxAdapter struct {
	tx pgx.Tx
}

var _ Executor = (*PgxTxAdapter)(nil)

func (p *PgxTxAdapter) Exec(ctx context.Context, sql string, args ...interface{}) (result RowsAffected, err error) {
	return p.tx.Exec(ctx, sql, args...)
}

func (p *PgxTxAdapter) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	return p.tx.Query(ctx, sql, args...)
}

func (p *PgxTxAdapter) QueryRow(ctx context.Context, sql string, args ...interface{}) Row {
	return p.tx.QueryRow(ctx, sql, args...)
}

func (p *PgxTxAdapter) SendBatch(ctx context.Context, batch *pgx.Batch) BatchResults {
	res := p.tx.SendBatch(ctx, batch)
	return batchResultsAdapter{BatchResults:res}
}
