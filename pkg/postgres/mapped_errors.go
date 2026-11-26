package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/BladeRunner322/orange-team-microservices/pkg/pgerrors"
)

// QueryExecutor — методы выполнения SQL, общие для пула pgx и транзакции.
type QueryExecutor interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type mappedErrors struct {
	db QueryExecutor
}

// WithMappedErrors возвращает исполнитель SQL с преобразованием ошибок pgx.
func WithMappedErrors(db QueryExecutor) QueryExecutor {
	return mappedErrors{db: db}
}

func (m mappedErrors) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	rows, err := m.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, pgerrors.MapErrors(err)
	}
	return rows, nil
}

func (m mappedErrors) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return pgxRow{m.db.QueryRow(ctx, sql, args...)}
}

func (m mappedErrors) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	tag, err := m.db.Exec(ctx, sql, args...)
	if err != nil {
		return tag, pgerrors.MapErrors(err)
	}
	return tag, nil
}
