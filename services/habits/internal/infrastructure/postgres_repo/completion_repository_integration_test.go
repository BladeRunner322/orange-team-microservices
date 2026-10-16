//go:build integration
// +build integration

package postgres_repo

import (
	"context"
	"testing"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletionRepository_Integration(t *testing.T) {
	completionRepo := NewCompletionRepository(testPool)
	habitRepo := NewHabitRepository(testPool)
	runCompletionRepositoryTests(t, completionRepo, habitRepo)
}

// mustCompletion создаёт доменную отметку или падает.
func mustCompletion(t *testing.T, habitID uuid.UUID, date time.Time) domain.Completion {
	t.Helper()

	d, err := domain.NewDate(date)
	require.NoError(t, err)

	return domain.NewCompletion(habitID, d)
}

// midnightUTC приводит t к полуночи UTC (как хранит DATE в PostgreSQL).
func midnightUTC(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// daysAgo возвращает дату N дней назад в полночь UTC.
func daysAgo(n int) time.Time {
	return midnightUTC(time.Now()).AddDate(0, 0, -n)
}

func runCompletionRepositoryTests(t *testing.T, repo *CompletionRepository, habitRepo *HabitRepository) {
	ctx := context.Background()

	// newHabit создаёт реальную привычку — на неё опирается FK в completions.
	newHabit := func(t *testing.T) uuid.UUID {
		t.Helper()
		h, err := habitRepo.Create(ctx, mustHabit(t,
			uuid.New(),
			"Habit "+uuid.New().String(),
			"desc",
		))
		require.NoError(t, err)
		return h.ID()
	}

	// ============================================================
	//  Complete
	// ============================================================

	t.Run("Complete добавляет отметку", func(t *testing.T) {
		habitID := newHabit(t)
		today := daysAgo(0)

		require.NoError(t, repo.Complete(ctx, mustCompletion(t, habitID, today)))

		ok, err := repo.IsCompleted(ctx, habitID, today)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("Complete повторно — идемпотентно (ON CONFLICT DO NOTHING)", func(t *testing.T) {
		habitID := newHabit(t)
		today := daysAgo(0)

		require.NoError(t, repo.Complete(ctx, mustCompletion(t, habitID, today)))
		require.NoError(t, repo.Complete(ctx, mustCompletion(t, habitID, today)))

		ok, err := repo.IsCompleted(ctx, habitID, today)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	// ============================================================
	//  Uncomplete
	// ============================================================

	t.Run("Uncomplete удаляет отметку", func(t *testing.T) {
		habitID := newHabit(t)
		today := daysAgo(0)

		require.NoError(t, repo.Complete(ctx, mustCompletion(t, habitID, today)))
		require.NoError(t, repo.Uncomplete(ctx, habitID, today))

		ok, err := repo.IsCompleted(ctx, habitID, today)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("Uncomplete без отметки — идемпотентно (без ошибки)", func(t *testing.T) {
		habitID := newHabit(t)
		err := repo.Uncomplete(ctx, habitID, daysAgo(0))
		require.NoError(t, err)
	})

	// ============================================================
	//  IsCompleted
	// ============================================================

	t.Run("IsCompleted — false, когда отметки нет", func(t *testing.T) {
		habitID := newHabit(t)
		ok, err := repo.IsCompleted(ctx, habitID, daysAgo(0))
		require.NoError(t, err)
		assert.False(t, ok)
	})

	// ============================================================
	//  ListByHabit
	// ============================================================

	t.Run("ListByHabit возвращает отметки за период включительно и сортирует по date ASC", func(t *testing.T) {
		habitID := newHabit(t)
		otherHabitID := newHabit(t)

		d1 := daysAgo(3)
		d2 := daysAgo(2)
		d3 := daysAgo(1)
		dOutBefore := daysAgo(5)
		dOutAfter := daysAgo(0)

		for _, d := range []time.Time{d1, d2, d3, dOutBefore, dOutAfter} {
			require.NoError(t, repo.Complete(ctx, mustCompletion(t, habitID, d)))
		}
		// чужая привычка — не должна попасть
		require.NoError(t, repo.Complete(ctx, mustCompletion(t, otherHabitID, d2)))

		list, err := repo.ListByHabit(ctx, habitID, d2, d3)
		require.NoError(t, err)
		require.Len(t, list, 2)

		assert.Equal(t, d2, list[0].Date().Time())
		assert.Equal(t, d3, list[1].Date().Time())
	})

	t.Run("ListByHabit пустой slice, не nil", func(t *testing.T) {
		habitID := newHabit(t)
		list, err := repo.ListByHabit(ctx, habitID, daysAgo(10), daysAgo(5))
		require.NoError(t, err)
		assert.NotNil(t, list)
		assert.Empty(t, list)
	})

	// ============================================================
	//  Streak
	// ============================================================

	t.Run("Streak — пусто → 0", func(t *testing.T) {
		habitID := newHabit(t)
		s, err := repo.Streak(ctx, habitID, daysAgo(0))
		require.NoError(t, err)
		assert.Equal(t, 0, s)
	})

	t.Run("Streak — только сегодня → 1", func(t *testing.T) {
		habitID := newHabit(t)
		require.NoError(t, repo.Complete(ctx, mustCompletion(t, habitID, daysAgo(0))))

		s, err := repo.Streak(ctx, habitID, daysAgo(0))
		require.NoError(t, err)
		assert.Equal(t, 1, s)
	})

	t.Run("Streak — только вчера → 1 (стрик держится до конца дня)", func(t *testing.T) {
		habitID := newHabit(t)
		require.NoError(t, repo.Complete(ctx, mustCompletion(t, habitID, daysAgo(1))))

		s, err := repo.Streak(ctx, habitID, daysAgo(0))
		require.NoError(t, err)
		assert.Equal(t, 1, s)
	})

	t.Run("Streak — сегодня и вчера → 2", func(t *testing.T) {
		habitID := newHabit(t)
		require.NoError(t, repo.Complete(ctx, mustCompletion(t, habitID, daysAgo(0))))
		require.NoError(t, repo.Complete(ctx, mustCompletion(t, habitID, daysAgo(1))))

		s, err := repo.Streak(ctx, habitID, daysAgo(0))
		require.NoError(t, err)
		assert.Equal(t, 2, s)
	})

	t.Run("Streak — 4 дня подряд → 4", func(t *testing.T) {
		habitID := newHabit(t)
		for i := 0; i < 4; i++ {
			require.NoError(t, repo.Complete(ctx, mustCompletion(t, habitID, daysAgo(i))))
		}

		s, err := repo.Streak(ctx, habitID, daysAgo(0))
		require.NoError(t, err)
		assert.Equal(t, 4, s)
	})

	t.Run("Streak — пропуск в середине обрезает стрик", func(t *testing.T) {
		habitID := newHabit(t)
		// сегодня, вчера, 3 дня назад (2 дня назад пропущен)
		require.NoError(t, repo.Complete(ctx, mustCompletion(t, habitID, daysAgo(0))))
		require.NoError(t, repo.Complete(ctx, mustCompletion(t, habitID, daysAgo(1))))
		require.NoError(t, repo.Complete(ctx, mustCompletion(t, habitID, daysAgo(3))))

		s, err := repo.Streak(ctx, habitID, daysAgo(0))
		require.NoError(t, err)
		assert.Equal(t, 2, s)
	})

	t.Run("Streak — сегодня не отмечено, но вчера и позавчера → 2", func(t *testing.T) {
		habitID := newHabit(t)
		require.NoError(t, repo.Complete(ctx, mustCompletion(t, habitID, daysAgo(1))))
		require.NoError(t, repo.Complete(ctx, mustCompletion(t, habitID, daysAgo(2))))

		s, err := repo.Streak(ctx, habitID, daysAgo(0))
		require.NoError(t, err)
		assert.Equal(t, 2, s)
	})

	t.Run("Streak — сегодня и вчера не отмечено → 0", func(t *testing.T) {
		habitID := newHabit(t)
		require.NoError(t, repo.Complete(ctx, mustCompletion(t, habitID, daysAgo(3))))

		s, err := repo.Streak(ctx, habitID, daysAgo(0))
		require.NoError(t, err)
		assert.Equal(t, 0, s)
	})

	t.Run("Streak — только позавчера → 0", func(t *testing.T) {
		habitID := newHabit(t)
		require.NoError(t, repo.Complete(ctx, mustCompletion(t, habitID, daysAgo(2))))

		s, err := repo.Streak(ctx, habitID, daysAgo(0))
		require.NoError(t, err)
		assert.Equal(t, 0, s)
	})
}