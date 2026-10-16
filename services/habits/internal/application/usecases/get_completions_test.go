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

func TestGetCompletions_Execute(t *testing.T) {
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

	markComplete := func(t *testing.T, repo *mockCompletionRepository, habitID uuid.UUID, d time.Time) {
		t.Helper()
		date, err := domain.NewDate(d)
		require.NoError(t, err)
		require.NoError(t, repo.Complete(context.Background(), domain.NewCompletion(habitID, date)))
	}

	today := dayOnly(time.Now().UTC())

	t.Run("успешно возвращает отметки за период", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		d1 := today.AddDate(0, 0, -3)
		d2 := today.AddDate(0, 0, -2)
		d3 := today.AddDate(0, 0, -1)
		dOutside := today.AddDate(0, 0, -10)

		for _, d := range []time.Time{d1, d2, d3, dOutside} {
			markComplete(t, completionRepo, habit.ID(), d)
		}

		uc := NewGetCompletions(repo, completionRepo, logger.NewTestLogger())

		result, err := uc.Execute(newCtx(userID), habit.ID(), d2, d3)

		require.NoError(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, d2, result[0].Date().Time())
		assert.Equal(t, d3, result[1].Date().Time())
	})

	t.Run("пустой период — последние 30 дней", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		dInside := today.AddDate(0, 0, -15)
		dOutside := today.AddDate(0, 0, -40)
		markComplete(t, completionRepo, habit.ID(), dInside)
		markComplete(t, completionRepo, habit.ID(), dOutside)

		uc := NewGetCompletions(repo, completionRepo, logger.NewTestLogger())

		result, err := uc.Execute(newCtx(userID), habit.ID(), time.Time{}, time.Time{})

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, dInside, result[0].Date().Time())
	})

	t.Run("только from — to = today", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		markComplete(t, completionRepo, habit.ID(), today)

		from := today.AddDate(0, 0, -5)
		uc := NewGetCompletions(repo, completionRepo, logger.NewTestLogger())

		result, err := uc.Execute(newCtx(userID), habit.ID(), from, time.Time{})

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, today, result[0].Date().Time())
	})

	t.Run("только to — from = to - 30 дней", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		// to = today - 10д → окно [today - 40д, today - 10д].
		// dInside = today - 20д — внутри.
		// dOutside = today - 45д — старше нижней границы.
		to := today.AddDate(0, 0, -10)
		dInside := today.AddDate(0, 0, -20)
		dOutside := today.AddDate(0, 0, -45)

		markComplete(t, completionRepo, habit.ID(), dInside)
		markComplete(t, completionRepo, habit.ID(), dOutside)

		uc := NewGetCompletions(repo, completionRepo, logger.NewTestLogger())

		result, err := uc.Execute(newCtx(userID), habit.ID(), time.Time{}, to)

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, dInside, result[0].Date().Time())
	})

	t.Run("from > to — ErrInvalidDate", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		uc := NewGetCompletions(repo, completionRepo, logger.NewTestLogger())

		from := today
		to := today.AddDate(0, 0, -5)

		_, err := uc.Execute(newCtx(userID), habit.ID(), from, to)

		assert.ErrorIs(t, err, domain.ErrInvalidDate)
	})

	t.Run("пустой результат — не nil", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		uc := NewGetCompletions(repo, completionRepo, logger.NewTestLogger())

		result, err := uc.Execute(newCtx(userID), habit.ID(), today.AddDate(0, 0, -5), today)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("привычка не найдена — ErrHabitNotFound", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		uc := NewGetCompletions(repo, completionRepo, logger.NewTestLogger())

		_, err := uc.Execute(newCtx(userID), uuid.New(), time.Time{}, time.Time{})

		assert.ErrorIs(t, err, domain.ErrHabitNotFound)
	})

	t.Run("чужая привычка — ErrHabitNotFound", func(t *testing.T) {
		ownerID := uuid.New()
		otherID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, ownerID, "Пить воду")

		uc := NewGetCompletions(repo, completionRepo, logger.NewTestLogger())

		_, err := uc.Execute(newCtx(otherID), habit.ID(), time.Time{}, time.Time{})

		assert.ErrorIs(t, err, domain.ErrHabitNotFound)
	})

	t.Run("удалённая привычка — история доступна", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		markComplete(t, completionRepo, habit.ID(), today)
		require.NoError(t, repo.MarkDeleted(context.Background(), habit.ID()))

		uc := NewGetCompletions(repo, completionRepo, logger.NewTestLogger())

		result, err := uc.Execute(newCtx(userID), habit.ID(), time.Time{}, time.Time{})

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, today, result[0].Date().Time())
	})

	t.Run("нет user_id в context — ErrUnauthenticated", func(t *testing.T) {
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		uc := NewGetCompletions(repo, completionRepo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), uuid.New(), time.Time{}, time.Time{})

		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})

	t.Run("ошибка репозитория при GetByID — пробрасывается", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		repo.err = errors.New("db is down")
		completionRepo := newMockCompletionRepository()
		uc := NewGetCompletions(repo, completionRepo, logger.NewTestLogger())

		_, err := uc.Execute(newCtx(userID), uuid.New(), time.Time{}, time.Time{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})

	t.Run("ошибка completionRepo при ListByHabit — пробрасывается", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		completionRepo := newMockCompletionRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		completionRepo.err = errors.New("redis is down")
		uc := NewGetCompletions(repo, completionRepo, logger.NewTestLogger())

		_, err := uc.Execute(newCtx(userID), habit.ID(), time.Time{}, time.Time{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "redis is down")
	})
}
