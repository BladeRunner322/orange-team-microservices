// Package postgres — интерфейсы и адаптеры для работы с PostgreSQL.
package postgres

import (
	"context"
	"time"
)

// Querier — общий интерфейс для Pool и Tx по методам Query/QueryRow/Exec.
// Позволяет репозиториям работать одинаково поверх пула и поверх транзакции.
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Exec(ctx context.Context, sql string, args ...any) (CommandTag, error)
}

// Pool — интерфейс пула соединений с базой данных.
type Pool interface {
	Querier
	BeginTx(ctx context.Context) (Tx, error)
	WithTx(ctx context.Context, fn func(Tx) error) error
	Close()
	OpTimeout() time.Duration
}

// Tx — интерфейс транзакции.
type Tx interface {
	Querier
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// Rows — интерфейс для работы со строками результата запроса.
type Rows interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...any) error
}

// Row — интерфейс для одной строки результата запроса.
type Row interface {
	Scan(dest ...any) error
}

// CommandTag — интерфейс для результата выполнения запроса.
type CommandTag interface {
	RowsAffected() int64
}