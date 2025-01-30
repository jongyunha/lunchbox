package postgresotel

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jongyunha/lunchbox/internal/postgres"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type tracedDB struct {
	db pgx.Tx
}

func (t tracedDB) Exec(ctx context.Context, query string, args ...interface{}) (ct pgconn.CommandTag, err error) {
	span := trace.SpanFromContext(ctx)
	defer func(started time.Time) {
		span.AddEvent("ExecContext", trace.WithAttributes(
			attribute.String("Query", query),
			attribute.Float64("Took", time.Since(started).Seconds()),
		))
		t.recordError(span, err)
	}(time.Now())

	return t.db.Exec(ctx, query, args...)
}

func (t tracedDB) Query(ctx context.Context, query string, args ...interface{}) (rows pgx.Rows, err error) {
	span := trace.SpanFromContext(ctx)
	defer func(started time.Time) {
		span.AddEvent("QueryContext", trace.WithAttributes(
			attribute.String("Query", query),
			attribute.Float64("Took", time.Since(started).Seconds()),
		))
		t.recordError(span, err)
	}(time.Now())

	return t.db.Query(ctx, query, args...)
}

func (t tracedDB) QueryRow(ctx context.Context, query string, args ...interface{}) (row pgx.Row) {
	span := trace.SpanFromContext(ctx)

	defer func(started time.Time) {
		span.AddEvent("QueryRowContext", trace.WithAttributes(
			attribute.String("Query", query),
			attribute.Float64("Took", time.Since(started).Seconds()),
		))
	}(time.Now())

	return t.db.QueryRow(ctx, query, args...)
}

var _ postgres.DBTX = (*tracedDB)(nil)

func Trace(db pgx.Tx) postgres.DBTX {
	return tracedDB{db: db}
}

func (t tracedDB) recordError(span trace.Span, err error) {
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			span.AddEvent("Database Error", trace.WithAttributes(
				attribute.String("Error", err.Error()),
				attribute.String("Code", pgErr.Code),
				attribute.String("Severity", pgErr.Severity),
				attribute.String("Message", pgErr.Message),
				attribute.String("Detail", pgErr.Detail),
			))
		} else {
			span.AddEvent("Database Error", trace.WithAttributes(
				attribute.String("Error", err.Error()),
			))
		}
	}
}
