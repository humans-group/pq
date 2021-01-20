package pq

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

const (
	operationNameExec        = "pq.Exec"
	operationNameQuery       = "pq.Query"
	operationNameQueryRow    = "pq.QueryRow"
	operationNameTransaction = "pq.Transaction"
	errLogKeyEvent           = "event"
	errLogKeyMessage         = "message"
	errLogValueErr           = "error"
)

type tracingAdapter struct {
	Transactor
	Executor
}

var _ Client = &tracingAdapter{}

func (ta *tracingAdapter) Transaction(ctx context.Context, f func(context.Context, Executor) error) error {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, operationNameTransaction)

	err := ta.Transactor.Transaction(spanCtx, f)

	if err != nil {
		traceErr(err, span)
	}

	span.Finish()

	return err
}

func (ta *tracingAdapter) Exec(ctx context.Context, sql string, args ...interface{}) (result RowsAffected, err error) {
	span, spanCtx := startSpan(ctx, operationNameExec)

	rowsAffected, err := ta.Executor.Exec(spanCtx, sql, args...)

	if err != nil {
		traceErr(err, span)
	}

	span.Finish()

	return rowsAffected, err
}

func (ta *tracingAdapter) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	span, spanCtx := startSpan(ctx, operationNameQuery)

	rows, err := ta.Executor.Query(spanCtx, sql, args...)

	if err != nil {
		traceErr(err, span)
	}
	span.Finish()

	return rows, err
}

func (ta *tracingAdapter) QueryRow(ctx context.Context, sql string, args ...interface{}) Row {
	span, spanCtx := startSpan(ctx, operationNameQueryRow)

	row := ta.Executor.QueryRow(spanCtx, sql, args...)

	span.Finish()

	return row
}

func traceErr(err error, span opentracing.Span) {
	ext.Error.Set(span, true)
	span.LogFields(
		log.String(errLogKeyEvent, errLogValueErr),
		log.String(errLogKeyMessage, err.Error()),
	)
}

func startSpan(ctx context.Context, name string) (opentracing.Span, context.Context) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, name)
	return span, spanCtx
}
