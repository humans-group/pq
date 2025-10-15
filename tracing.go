package pq

import (
	"context"

	otel "github.com/humans-group/opentelemetry-go"
	"github.com/humans-group/opentelemetry-go/attribute"
	"github.com/humans-group/opentelemetry-go/trace"
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
	tracer := otel.Tracer("db")
	_, span := tracer.Start(ctx, operationNameTransaction)
	defer span.End()

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
	tracer := otel.Tracer("db")
	_, span := tracer.Start(ctx, operationNameExec)
	defer span.End()
	span.AddEvent("db", trace.WithAttributes(
		attribute.String("db.operation", operationNameExec),
		attribute.String("db.sql", sql),
	))

	rowsAffected, err := ta.Executor.Exec(ctx, sql, args...)

	if err != nil {
		traceErr(err, span)
	}

	return rowsAffected, err
}

func (ta *tracingAdapter) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	tracer := otel.Tracer("db")
	_, span := tracer.Start(ctx, operationNameQuery)
	defer span.End()
	span.AddEvent("db", trace.WithAttributes(
		attribute.String("db.operation", operationNameQuery),
		attribute.String("db.sql", sql),
	))

	rows, err := ta.Executor.Query(ctx, sql, args...)

	if err != nil {
		traceErr(err, span)
	}

	return rows, err
}

func (ta *tracingAdapter) QueryRow(ctx context.Context, sql string, args ...interface{}) Row {
	tracer := otel.Tracer("db")
	_, span := tracer.Start(ctx, operationNameQueryRow)
	defer span.End()
	span.AddEvent("db", trace.WithAttributes(
		attribute.String("db.operation", operationNameQueryRow),
		attribute.String("db.sql", sql),
	))

	row := ta.Executor.QueryRow(ctx, sql, args...)

	return row
}

func traceErr(err error, span trace.Span) {
	span.AddEvent("error", trace.WithAttributes(
		attribute.String("db.error", err.Error()),
	))
}
