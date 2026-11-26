// Package postgres — интерфейсы и адаптеры для работы с PostgreSQL.
package postgres

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
