package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// pgxTx — адаптер для pgx.Tx.
type pgxTx struct {
	tx pgx.Tx
}

// NewTx — конструктор адаптера (используется в PgxPool).
func NewTx(tx pgx.Tx) Tx {
	return &pgxTx{tx: tx}
}

func (t *pgxTx) Query(ctx context.Context, sql string, args ...any) (Rows, error) {
	rows, err := t.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, mapErrors(err)
	}
	return pgxRows{rows}, nil
}

func (t *pgxTx) QueryRow(ctx context.Context, sql string, args ...any) Row {
	row := t.tx.QueryRow(ctx, sql, args...)
	return pgxRow{row}
}

func (t *pgxTx) Exec(ctx context.Context, sql string, args ...any) (CommandTag, error) {
	tag, err := t.tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, mapErrors(err)
	}
	return pgxCommandTag{tag}, nil
}

func (t *pgxTx) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *pgxTx) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}
