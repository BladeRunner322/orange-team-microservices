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

func TestGetHabits_Execute(t *testing.T) {
	newCtx := func(userID uuid.UUID) context.Context {
		return ctxkeys.WithUserID(context.Background(), userID.String())
	}

	// createHabit создаёт привычку через репозиторий (в обход usecase,
	// чтобы не зависеть от логики CreateHabit).
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

	// markCompleted проставляет отметку за дату.
	markCompleted := func(t *testing.T, repo *mockCompletionRepository, habitID uuid.UUID, d time.Time) {
		t.Helper()
		date, err := domain.NewDate(d)
		require.NoError(t, err)
		require.NoError(t, repo.Complete(context.Background(), domain.NewCompletion(habitID, date)))
	}

	t.Run("пустой список — не nil", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		uc := NewGetHabits(repo, completionRepo, logger.NewTestLogger())

		result, err := uc.Execute(newCtx(userID))

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("возвращает только привычки текущего пользователя", func(t *testing.T) {
		userID := uuid.New()
		otherID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()

		createHabit(t, repo, userID, "Первая")
		createHabit(t, repo, userID, "Вторая")
		createHabit(t, repo, otherID, "Чужая")

		uc := NewGetHabits(repo, completionRepo, logger.NewTestLogger())

		result, err := uc.Execute(newCtx(userID))

		require.NoError(t, err)
		require.Len(t, result, 2)
		for _, hws := range result {
			assert.Equal(t, userID, hws.Habit.UserID())
		}
	})

	t.Run("не включает удалённые привычки", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()

		active := createHabit(t, repo, userID, "Активная")
		toDelete := createHabit(t, repo, userID, "Удалённая")

		require.NoError(t, repo.MarkDeleted(context.Background(), toDelete.ID()))

		uc := NewGetHabits(repo, completionRepo, logger.NewTestLogger())

		result, err := uc.Execute(newCtx(userID))

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, active.ID(), result[0].Habit.ID())
	})

	t.Run("считает стрик и флаг completedToday для каждой привычки", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()

		withStreak := createHabit(t, repo, userID, "С отметками")
		withoutStreak := createHabit(t, repo, userID, "Без отметок")

		today := dayOnly(time.Now().UTC())
		yesterday := today.AddDate(0, 0, -1)

		markCompleted(t, completionRepo, withStreak.ID(), today)
		markCompleted(t, completionRepo, withStreak.ID(), yesterday)

		uc := NewGetHabits(repo, completionRepo, logger.NewTestLogger())

		result, err := uc.Execute(newCtx(userID))

		require.NoError(t, err)
		require.Len(t, result, 2)

		// Находим по id — порядок не гарантирован (map в моке).
		byID := make(map[uuid.UUID]HabitWithStreak, len(result))
		for _, hws := range result {
			byID[hws.Habit.ID()] = hws
		}

		got1 := byID[withStreak.ID()]
		assert.Equal(t, 2, got1.Streak)
		assert.True(t, got1.CompletedToday)

		got2 := byID[withoutStreak.ID()]
		assert.Equal(t, 0, got2.Streak)
		assert.False(t, got2.CompletedToday)
	})

	t.Run("нет user_id в context — ErrUnauthenticated", func(t *testing.T) {
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		uc := NewGetHabits(repo, completionRepo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background())

		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})

	t.Run("ошибка репозитория при ListByUser — пробрасывается", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		repo.err = errors.New("db is down")
		completionRepo := newMockCompletionRepository()
		uc := NewGetHabits(repo, completionRepo, logger.NewTestLogger())

		_, err := uc.Execute(newCtx(userID))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})

	t.Run("ошибка completionRepo при Streak — пробрасывается", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		createHabit(t, repo, userID, "Пить воду")

		completionRepo.err = errors.New("redis is down")
		uc := NewGetHabits(repo, completionRepo, logger.NewTestLogger())

		_, err := uc.Execute(newCtx(userID))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "redis is down")
	})
}
