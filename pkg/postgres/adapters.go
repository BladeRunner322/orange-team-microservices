package postgres

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// pgxRows — адаптер для pgx.Rows.
type pgxRows struct {
	pgx.Rows
}

func (r pgxRows) Scan(dest ...any) error {
	return r.Rows.Scan(dest...)
}

// pgxRow — адаптер для pgx.Row.
type pgxRow struct {
	pgx.Row
}

func (r pgxRow) Scan(dest ...any) error {
	err := r.Row.Scan(dest...)
	if err != nil {
		return mapErrors(err)
	}
	return nil
}

// pgxCommandTag — адаптер для pgconn.CommandTag.
type pgxCommandTag struct {
	pgconn.CommandTag
}

func (t pgxCommandTag) RowsAffected() int64 {
	return t.CommandTag.RowsAffected()
}

// mapErrors преобразует ошибки pgx в общие ошибки пакета postgres.
//
// Использует два %w в одном fmt.Errorf (поддерживается с Go 1.20):
//   - первое %w — оригинал (PgError), чтобы errors.As на результате
//     мог достать ConstraintName через ConstraintNameFromError;
//   - второе %w — сентинел (ErrViolatesUnique и т.п.), чтобы работал
//     errors.Is.
func mapErrors(err error) error {
	const (
		pgxViolatesForeignKeyErrorCode = "23503"
		pgxViolatesUniqueErrorCode     = "23505"
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoRows
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgxViolatesForeignKeyErrorCode:
			return fmt.Errorf("%w: %w", err, ErrViolatesForeignKey)
		case pgxViolatesUniqueErrorCode:
			return fmt.Errorf("%w: %w", err, ErrViolatesUnique)
		}
	}

	return fmt.Errorf("%w: %w", err, ErrUnknown)
}

// ConstraintNameFromError извлекает имя нарушенного constraint из ошибки
// PostgreSQL. Работает для любого кода, где PgError содержит ConstraintName
// (unique violation 23505, foreign key 23503, check 23514, not-null 23502).
//
// Возвращает пустую строку, если:
//   - ошибка не содержит PgError в цепочке (например, сетевая ошибка);
//   - PgError есть, но ConstraintName пустой (например, ошибка синтаксиса SQL).
//
// Используется репозиториями, когда в одной таблице несколько unique-индексов
// и надо понять, какой именно нарушен — чтобы вернуть правильную доменную
// ошибку (см. ADR-007 S-5).
func ConstraintNameFromError(err error) string {
	if err == nil {
		return ""
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName
	}

	return ""
}