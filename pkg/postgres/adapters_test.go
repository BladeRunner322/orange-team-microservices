package postgres

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestConstraintNameFromError(t *testing.T) {
	t.Run("nil — пустая строка", func(t *testing.T) {
		assert.Empty(t, ConstraintNameFromError(nil))
	})

	t.Run("PgError с ConstraintName", func(t *testing.T) {
		err := &pgconn.PgError{
			Code:           "23505",
			ConstraintName: "idx_auth_users_email",
		}

		assert.Equal(t, "idx_auth_users_email", ConstraintNameFromError(err))
	})

	t.Run("PgError в цепочке через fmt.Errorf %w", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code:           "23505",
			ConstraintName: "idx_exercises_name_active",
		}
		wrapped := fmt.Errorf("insert exercise: %w", pgErr)

		assert.Equal(t, "idx_exercises_name_active", ConstraintNameFromError(wrapped))
	})

	t.Run("PgError без ConstraintName — пустая строка", func(t *testing.T) {
		err := &pgconn.PgError{
			Code:    "42601", // syntax_error
			Message: "syntax error at or near \"SELEC\"",
		}

		assert.Empty(t, ConstraintNameFromError(err))
	})

	t.Run("обычная ошибка без PgError — пустая строка", func(t *testing.T) {
		err := errors.New("network timeout")

		assert.Empty(t, ConstraintNameFromError(err))
	})

	t.Run("результат mapErrors тоже даёт ConstraintName", func(t *testing.T) {
		// mapErrors оборачивает оригинальную ошибку через %v и %w.
		// errors.As должен пройти по цепочке и найти PgError.
		pgErr := &pgconn.PgError{
			Code:           "23505",
			ConstraintName: "idx_auth_users_email",
		}
		mapped := mapErrors(pgErr)

		assert.ErrorIs(t, mapped, ErrViolatesUnique)
		assert.Equal(t, "idx_auth_users_email", ConstraintNameFromError(mapped))
	})
}