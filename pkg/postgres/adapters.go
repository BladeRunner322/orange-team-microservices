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
			return fmt.Errorf("%v: %w", err, ErrViolatesForeignKey)
		case pgxViolatesUniqueErrorCode:
			return fmt.Errorf("%v: %w", err, ErrViolatesUnique)
		}
	}

	return fmt.Errorf("%v: %w", err, ErrUnknown)
}
