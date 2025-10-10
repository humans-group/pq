package pq

import (
	"context"

	"github.com/humans-group/opentelemetry-go/trace"
	"github.com/humans-group/opentelemetry-go/attribute"

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
	span := trace.SpanFromContext(ctx)
	span.AddEvent("db", trace.WithAttributes(
		attribute.String("db.operation", operationNameTransaction),
	))
	err := ta.Transactor.Transaction(ctx, f)

	if err != nil {
		traceErr(err, span)
	}

	return err
}

func (ta *tracingAdapter) Exec(ctx context.Context, sql string, args ...interface{}) (result RowsAffected, err error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("db", trace.WithAttributes(
		attribute.String("db.operation", operationNameExec),
	))

	rowsAffected, err := ta.Executor.Exec(ctx, sql, args...)

	if err != nil {
		traceErr(err, span)
	}

	return rowsAffected, err
}

func (ta *tracingAdapter) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("db", trace.WithAttributes(
		attribute.String("db.operation", operationNameQuery),
	))

	rows, err := ta.Executor.Query(ctx, sql, args...)

	if err != nil {
		traceErr(err, span)
	}

	return rows, err
}

func (ta *tracingAdapter) QueryRow(ctx context.Context, sql string, args ...interface{}) Row {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("db", trace.WithAttributes(
		attribute.String("db.operation", operationNameQueryRow),
	))

	row := ta.Executor.QueryRow(ctx, sql, args...)

	return row
}

func traceErr(err error, span trace.Span) {
	span.AddEvent("error", trace.WithAttributes(
		attribute.String("db.error", err.Error()),
	))
}