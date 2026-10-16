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

func TestCompleteHabit_Execute(t *testing.T) {
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

	today := dayOnly(time.Now().UTC())
	yesterday := today.AddDate(0, 0, -1)
	threeDaysAgo := today.AddDate(0, 0, -3)
	eightDaysAgo := today.AddDate(0, 0, -8)
	tomorrow := today.AddDate(0, 0, 1)

	t.Run("успешно отмечает сегодня", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		uc := NewCompleteHabit(repo, completionRepo, logger.NewTestLogger())

		// Act
		err := uc.Execute(newCtx(userID), habit.ID(), today)

		// Assert
		require.NoError(t, err)

		ok, err := completionRepo.IsCompleted(context.Background(), habit.ID(), today)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("успешно отмечает вчера", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		uc := NewCompleteHabit(repo, completionRepo, logger.NewTestLogger())

		err := uc.Execute(newCtx(userID), habit.ID(), yesterday)

		require.NoError(t, err)
		ok, _ := completionRepo.IsCompleted(context.Background(), habit.ID(), yesterday)
		assert.True(t, ok)
	})

	t.Run("успешно отмечает 7 дней назад (граница)", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		uc := NewCompleteHabit(repo, completionRepo, logger.NewTestLogger())

		err := uc.Execute(newCtx(userID), habit.ID(), today.AddDate(0, 0, -7))

		require.NoError(t, err)
	})

	t.Run("идемпотентность: повторная отметка на ту же дату — не ошибка", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		uc := NewCompleteHabit(repo, completionRepo, logger.NewTestLogger())

		require.NoError(t, uc.Execute(newCtx(userID), habit.ID(), today))
		err := uc.Execute(newCtx(userID), habit.ID(), today)

		assert.NoError(t, err)
	})

	t.Run("будущее — ErrInvalidDate", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		uc := NewCompleteHabit(repo, completionRepo, logger.NewTestLogger())

		err := uc.Execute(newCtx(userID), habit.ID(), tomorrow)

		assert.ErrorIs(t, err, domain.ErrInvalidDate)
	})

	t.Run("старше 7 дней — ErrInvalidDate", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		uc := NewCompleteHabit(repo, completionRepo, logger.NewTestLogger())

		err := uc.Execute(newCtx(userID), habit.ID(), eightDaysAgo)

		assert.ErrorIs(t, err, domain.ErrInvalidDate)
	})

	t.Run("3 дня назад — валидно", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		uc := NewCompleteHabit(repo, completionRepo, logger.NewTestLogger())

		err := uc.Execute(newCtx(userID), habit.ID(), threeDaysAgo)

		require.NoError(t, err)
	})

	t.Run("привычка не найдена — ErrHabitNotFound", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		uc := NewCompleteHabit(repo, completionRepo, logger.NewTestLogger())

		err := uc.Execute(newCtx(userID), uuid.New(), today)

		assert.ErrorIs(t, err, domain.ErrHabitNotFound)
	})

	t.Run("чужая привычка — ErrHabitNotFound", func(t *testing.T) {
		ownerID := uuid.New()
		otherID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, ownerID, "Пить воду")

		uc := NewCompleteHabit(repo, completionRepo, logger.NewTestLogger())

		err := uc.Execute(newCtx(otherID), habit.ID(), today)

		assert.ErrorIs(t, err, domain.ErrHabitNotFound)
	})

	t.Run("удалённая привычка — ErrHabitNotFound", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, userID, "Пить воду")
		require.NoError(t, repo.MarkDeleted(context.Background(), habit.ID()))

		uc := NewCompleteHabit(repo, completionRepo, logger.NewTestLogger())

		err := uc.Execute(newCtx(userID), habit.ID(), today)

		assert.ErrorIs(t, err, domain.ErrHabitNotFound)
	})

	t.Run("нет user_id в context — ErrUnauthenticated", func(t *testing.T) {
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		uc := NewCompleteHabit(repo, completionRepo, logger.NewTestLogger())

		err := uc.Execute(context.Background(), uuid.New(), today)

		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})

	t.Run("ошибка репозитория при GetByID — пробрасывается", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		repo.err = errors.New("db is down")
		completionRepo := newMockCompletionRepository()
		uc := NewCompleteHabit(repo, completionRepo, logger.NewTestLogger())

		err := uc.Execute(newCtx(userID), uuid.New(), today)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})

	t.Run("ошибка completionRepo при Complete — пробрасывается", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		completionRepo.err = errors.New("redis is down")
		uc := NewCompleteHabit(repo, completionRepo, logger.NewTestLogger())

		err := uc.Execute(newCtx(userID), habit.ID(), today)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "redis is down")
	})
}
