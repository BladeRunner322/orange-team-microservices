package postgres

import (
	"github.com/BladeRunner322/orange-team-microservices/pkg/pgerrors"
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
		return pgerrors.MapErrors(err)
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
