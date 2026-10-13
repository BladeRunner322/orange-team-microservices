package pgerrors

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

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