package postgres

import (
	"context"
	"time"
)

// Pool — интерфейс пула соединений с базой данных.
type Pool interface {
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Exec(ctx context.Context, sql string, args ...any) (CommandTag, error)
	Close()
	OpTimeout() time.Duration
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

// CommandTag — интерфейс для результата выполнения запроса (например, количество затронутых строк).
type CommandTag interface {
	RowsAffected() int64
}
