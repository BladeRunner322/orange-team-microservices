package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BladeRunner322/orange-team-microservices/pkg/ctxkeys"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
)

func TestGetHabit_Execute(t *testing.T) {
	// setupHabit создаёт привычку в репозитории и возвращает её вместе с ctx.
	setupHabit := func(t *testing.T, repo *mockHabitRepository, userID uuid.UUID, name string) domain.Habit {
		t.Helper()
		n, err := domain.NewName(name)
		require.NoError(t, err)
		d, err := domain.NewDescription("Описание")
		require.NoError(t, err)
		habit := domain.NewHabit(userID, n, d)
		created, err := repo.Create(context.Background(), habit)
		require.NoError(t, err)
		return created
	}

	newCtx := func(userID uuid.UUID) context.Context {
		return ctxkeys.WithUserID(context.Background(), userID.String())
	}

	t.Run("успешно возвращает привычку без стрика", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := setupHabit(t, repo, userID, "Пить воду")

		uc := NewGetHabit(repo, completionRepo, logger.NewTestLogger())

		// Act
		result, err := uc.Execute(newCtx(userID), habit.ID())

		// Assert
		require.NoError(t, err)
		assert.Equal(t, habit.ID(), result.Habit.ID())
		assert.Equal(t, 0, result.Streak)
		assert.False(t, result.CompletedToday)
	})

	t.Run("возвращает стрик и флаг completedToday", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := setupHabit(t, repo, userID, "Пить воду")

		today := dayOnly(time.Now().UTC())
		yesterday := today.AddDate(0, 0, -1)

		// Отмечаем сегодня и вчера.
		for _, d := range []time.Time{today, yesterday} {
			date, _ := domain.NewDate(d)
			require.NoError(t, completionRepo.Complete(context.Background(), domain.NewCompletion(habit.ID(), date)))
		}

		uc := NewGetHabit(repo, completionRepo, logger.NewTestLogger())

		// Act
		result, err := uc.Execute(newCtx(userID), habit.ID())

		// Assert
		require.NoError(t, err)
		assert.Equal(t, 2, result.Streak)
		assert.True(t, result.CompletedToday)
	})

	t.Run("привычка не найдена — ErrHabitNotFound", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		uc := NewGetHabit(repo, completionRepo, logger.NewTestLogger())

		_, err := uc.Execute(newCtx(userID), uuid.New())

		assert.ErrorIs(t, err, domain.ErrHabitNotFound)
	})

	t.Run("чужая привычка — ErrHabitNotFound (не раскрываем существование)", func(t *testing.T) {
		ownerID := uuid.New()
		otherID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := setupHabit(t, repo, ownerID, "Пить воду")

		uc := NewGetHabit(repo, completionRepo, logger.NewTestLogger())

		_, err := uc.Execute(newCtx(otherID), habit.ID())

		assert.ErrorIs(t, err, domain.ErrHabitNotFound)
	})

	t.Run("нет user_id в context — ErrUnauthenticated", func(t *testing.T) {
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		uc := NewGetHabit(repo, completionRepo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), uuid.New())

		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})

	t.Run("ошибка репозитория при GetByID — пробрасывается", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		repo.err = errors.New("db is down")
		completionRepo := newMockCompletionRepository()
		uc := NewGetHabit(repo, completionRepo, logger.NewTestLogger())

		_, err := uc.Execute(newCtx(userID), uuid.New())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})

	t.Run("ошибка completionRepo при Streak — пробрасывается", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := setupHabit(t, repo, userID, "Пить воду")

		completionRepo.err = errors.New("redis is down")
		uc := NewGetHabit(repo, completionRepo, logger.NewTestLogger())

		_, err := uc.Execute(newCtx(userID), habit.ID())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "redis is down")
	})
}
