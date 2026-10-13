package pgerrors

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// MapErrors преобразует ошибки pgx/pgconn в сентинелы пакета pgerrors.
//
// Использует два %w в одном fmt.Errorf (поддерживается с Go 1.20):
//   - первое %w — оригинал (PgError), чтобы errors.As на результате
//     мог достать ConstraintName через ConstraintNameFromError;
//   - второе %w — сентинел (ErrViolatesUnique и т.п.), чтобы работал
//     errors.Is.
func MapErrors(err error) error {
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