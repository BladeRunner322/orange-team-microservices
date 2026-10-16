package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BladeRunner322/orange-team-microservices/pkg/ctxkeys"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
)

func TestDeleteHabit_Execute(t *testing.T) {
	newCtx := func(userID uuid.UUID) context.Context {
		return ctxkeys.WithUserID(context.Background(), userID.String())
	}

	createHabit := func(t *testing.T, repo *mockHabitRepository, userID uuid.UUID, name string) domain.Habit {
		t.Helper()
		n, err := domain.NewName(name)
		require.NoError(t, err)
		d, err := domain.NewDescription("Описание")
		require.NoError(t, err)
		created, err := repo.Create(context.Background(), domain.NewHabit(userID, n, d))
		require.NoError(t, err)
		return created
	}

	t.Run("успешно помечает привычку удалённой", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		uc := NewDeleteHabit(repo, logger.NewTestLogger())

		err := uc.Execute(newCtx(userID), habit.ID())

		require.NoError(t, err)

		found, err := repo.GetByID(context.Background(), habit.ID())
		require.NoError(t, err)
		assert.True(t, found.IsDeleted())
	})

	t.Run("идемпотентность: повторный вызов — не ошибка", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		habit := createHabit(t, repo, userID, "Пить воду")
		uc := NewDeleteHabit(repo, logger.NewTestLogger())

		require.NoError(t, uc.Execute(newCtx(userID), habit.ID()))
		err := uc.Execute(newCtx(userID), habit.ID())

		assert.NoError(t, err, "повторный DELETE должен быть идемпотентен")
	})

	t.Run("идемпотентность: несуществующий id — не ошибка", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		uc := NewDeleteHabit(repo, logger.NewTestLogger())

		err := uc.Execute(newCtx(userID), uuid.New())

		assert.NoError(t, err, "DELETE несуществующего должен быть идемпотентен")
	})

	t.Run("чужая привычка — ErrHabitNotFound", func(t *testing.T) {
		ownerID := uuid.New()
		otherID := uuid.New()
		repo := newMockHabitRepository()
		habit := createHabit(t, repo, ownerID, "Пить воду")
		uc := NewDeleteHabit(repo, logger.NewTestLogger())

		err := uc.Execute(newCtx(otherID), habit.ID())

		assert.ErrorIs(t, err, domain.ErrHabitNotFound)

		// Привычка не тронута.
		found, err := repo.GetByID(context.Background(), habit.ID())
		require.NoError(t, err)
		assert.False(t, found.IsDeleted())
	})

	t.Run("нет user_id в context — ErrUnauthenticated", func(t *testing.T) {
		repo := newMockHabitRepository()
		uc := NewDeleteHabit(repo, logger.NewTestLogger())

		err := uc.Execute(context.Background(), uuid.New())

		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})

	t.Run("ошибка репозитория при GetByID — пробрасывается", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		repo.err = errors.New("db is down")
		uc := NewDeleteHabit(repo, logger.NewTestLogger())

		err := uc.Execute(newCtx(userID), uuid.New())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}
