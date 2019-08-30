package pg

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

const (
	operationNameExec     = "pq.Exec"
	operationNameQuery    = "pq.Query"
	operationNameQueryRow = "pq.QueryRow"
	errLogKeyEvent        = "event"
	errLogKeyMessage      = "message"
	errLogValueErr        = "error"
	logKeySql             = "sql"
)

type tracingAdapter struct {
	Client
}

func (ta *tracingAdapter) Exec(ctx context.Context, sql string, args ...interface{}) (result RowsAffected, err error) {
	span, spanCtx := startSpan(ctx, sql, operationNameExec)

	rowsAffected, err := ta.Client.Exec(spanCtx, sql, args...)

	if err != nil {
		traceErr(err, span)
	}

	span.Finish()

	return rowsAffected, err
}

func (ta *tracingAdapter) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	span, spanCtx := startSpan(ctx, sql, operationNameQuery)

	rows, err := ta.Client.Query(spanCtx, sql, args...)

	if err != nil {
		traceErr(err, span)
	}
	span.Finish()

	return rows, err
}

func (ta *tracingAdapter) QueryRow(ctx context.Context, sql string, args ...interface{}) Row {
	span, spanCtx := startSpan(ctx, sql, operationNameQueryRow)

	row := ta.Client.QueryRow(spanCtx, sql, args...)

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

func startSpan(ctx context.Context, sql string, name string) (opentracing.Span, context.Context) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, name)
	span.LogFields(log.String(logKeySql, sql))
	return span, spanCtx
}
