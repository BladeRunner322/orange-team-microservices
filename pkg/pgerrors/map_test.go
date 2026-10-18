package pgerrors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestMapErrors(t *testing.T) {
	t.Run("pgx.ErrNoRows → ErrNoRows", func(t *testing.T) {
		mapped := MapErrors(pgx.ErrNoRows)

		assert.ErrorIs(t, mapped, ErrNoRows)
	})

	t.Run("PgError 23505 (unique) → ErrViolatesUnique", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "23505"}

		mapped := MapErrors(pgErr)

		assert.ErrorIs(t, mapped, ErrViolatesUnique)
	})

	t.Run("PgError 23503 (foreign key) → ErrViolatesForeignKey", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "23503"}

		mapped := MapErrors(pgErr)

		assert.ErrorIs(t, mapped, ErrViolatesForeignKey)
	})

	t.Run("PgError с другим кодом → ErrUnknown", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "42601"} // syntax error

		mapped := MapErrors(pgErr)

		assert.ErrorIs(t, mapped, ErrUnknown)
	})

	t.Run("обычная ошибка → ErrUnknown", func(t *testing.T) {
		mapped := MapErrors(errors.New("network timeout"))

		assert.ErrorIs(t, mapped, ErrUnknown)
	})

	t.Run("PgError обёрнут в fmt.Errorf %w — errors.As проходит", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "23505", ConstraintName: "idx_test"}
		wrapped := fmt.Errorf("insert: %w", pgErr)

		mapped := MapErrors(wrapped)

		assert.ErrorIs(t, mapped, ErrViolatesUnique)
		assert.Equal(t, "idx_test", ConstraintNameFromError(mapped))
	})
}
