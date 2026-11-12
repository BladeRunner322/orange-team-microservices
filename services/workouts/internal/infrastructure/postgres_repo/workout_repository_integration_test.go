//go:build integration
// +build integration

package postgres_repo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BladeRunner322/orange-team-microservices/pkg/nullable"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
)

func TestWorkoutRepository_Integration(t *testing.T) {
	repo := NewWorkoutRepository(testPool)
	runWorkoutRepositoryTests(t, repo)
}

// mustWorkout создаёт доменную тренировку или падает.
func mustWorkout(t *testing.T, userID uuid.UUID, coefficient int) domain.Workout {
	t.Helper()
	return domain.NewWorkout(userID, coefficient)
}

func runWorkoutRepositoryTests(t *testing.T, repo *WorkoutRepository) {
	ctx := context.Background()

	// ============================================================
	//  Create
	// ============================================================

	t.Run("Create сохраняет тренировку и возвращает её из RETURNING", func(t *testing.T) {
		userID := uuid.New()
		w := mustWorkout(t, userID, 5)

		created, err := repo.Create(ctx, w)
		require.NoError(t, err)

		assert.Equal(t, w.ID(), created.ID())
		assert.Equal(t, userID, created.UserID())
		assert.Equal(t, domain.StatusPlanned, created.Status())
		assert.Equal(t, 5, created.PersonalScoreCoefficient())
		assert.Equal(t, 0, created.WorkoutScore())
		assert.False(t, created.IsDeleted())
		assert.False(t, created.CreatedAt().IsZero())
		assert.Nil(t, created.UpdatedAt())
	})

	// ============================================================
	//  GetByID
	// ============================================================

	t.Run("GetByID возвращает активную тренировку", func(t *testing.T) {
		userID := uuid.New()
		created, err := repo.Create(ctx, mustWorkout(t, userID, 5))
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, userID, created.ID())
		require.NoError(t, err)

		assert.Equal(t, created.ID(), found.ID())
		assert.Equal(t, userID, found.UserID())
		assert.False(t, found.IsDeleted())
	})

	t.Run("GetByID не найдено — ErrWorkoutNotFound", func(t *testing.T) {
		_, err := repo.GetByID(ctx, uuid.New(), uuid.New())
		assert.ErrorIs(t, err, domain.ErrWorkoutNotFound)
	})

	t.Run("GetByID чужая тренировка — ErrWorkoutNotFound", func(t *testing.T) {
		ownerID := uuid.New()
		otherID := uuid.New()
		created, err := repo.Create(ctx, mustWorkout(t, ownerID, 5))
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, otherID, created.ID())
		assert.ErrorIs(t, err, domain.ErrWorkoutNotFound)
	})

	t.Run("GetByID возвращает удалённое с IsDeleted=true", func(t *testing.T) {
		userID := uuid.New()
		created, err := repo.Create(ctx, mustWorkout(t, userID, 5))
		require.NoError(t, err)
		require.NoError(t, repo.MarkDeleted(ctx, userID, created.ID()))

		found, err := repo.GetByID(ctx, userID, created.ID())
		require.NoError(t, err)
		assert.True(t, found.IsDeleted())
	})

	// ============================================================
	//  GetByUserID
	// ============================================================

	t.Run("GetByUserID возвращает только активные тренировки текущего user", func(t *testing.T) {
		userID := uuid.New()

		w1, err := repo.Create(ctx, mustWorkout(t, userID, 5))
		require.NoError(t, err)
		w2, err := repo.Create(ctx, mustWorkout(t, userID, 5))
		require.NoError(t, err)

		// чужая — не должна попасть
		_, err = repo.Create(ctx, mustWorkout(t, uuid.New(), 5))
		require.NoError(t, err)

		// удалённая — не должна попасть
		deleted, err := repo.Create(ctx, mustWorkout(t, userID, 5))
		require.NoError(t, err)
		require.NoError(t, repo.MarkDeleted(ctx, userID, deleted.ID()))

		list, err := repo.GetByUserID(ctx, userID)
		require.NoError(t, err)
		require.Len(t, list, 2)

		ids := []uuid.UUID{list[0].ID(), list[1].ID()}
		assert.Contains(t, ids, w1.ID())
		assert.Contains(t, ids, w2.ID())

		for _, w := range list {
			assert.Equal(t, userID, w.UserID())
			assert.False(t, w.IsDeleted())
		}

		// порядок по created_at DESC (новые сверху)
		for i := 1; i < len(list); i++ {
			assert.False(t, list[i].CreatedAt().After(list[i-1].CreatedAt()))
		}
	})

	t.Run("GetByUserID пустой slice, не nil", func(t *testing.T) {
		list, err := repo.GetByUserID(ctx, uuid.New())
		require.NoError(t, err)
		assert.NotNil(t, list)
		assert.Empty(t, list)
	})

	// ============================================================
	//  Update + optimistic locking
	// ============================================================

	t.Run("Update меняет поля и возвращает свежий updated_at", func(t *testing.T) {
		userID := uuid.New()
		created, err := repo.Create(ctx, mustWorkout(t, userID, 5))
		require.NoError(t, err)

		intensity, err := domain.NewIntensity(8)
		require.NoError(t, err)
		status := domain.StatusInProgress
		require.NoError(t, created.ApplyPatch(domain.NewWorkoutPatch(
			&status,
			nullableIntensitySet(&intensity),
		)))

		updated, err := repo.Update(ctx, created)
		require.NoError(t, err)

		assert.Equal(t, domain.StatusInProgress, updated.Status())
		require.NotNil(t, updated.Intensity())
		assert.Equal(t, 8, updated.Intensity().Int())
		assert.Equal(t, 1, updated.Version(), "version должен инкрементироваться")
		require.NotNil(t, updated.UpdatedAt())
		require.NotNil(t, updated.StartedAt(), "переход в in_progress ставит started_at")
	})

	t.Run("Update с устаревшей version — ErrConflict", func(t *testing.T) {
		userID := uuid.New()
		created, err := repo.Create(ctx, mustWorkout(t, userID, 5))
		require.NoError(t, err)

		_, err = repo.Update(ctx, created)
		require.NoError(t, err)

		_, err = repo.Update(ctx, created)
		assert.ErrorIs(t, err, domain.ErrConflict)
	})

	t.Run("Update несуществующего — ErrConflict", func(t *testing.T) {
		_, err := repo.Update(ctx, mustWorkout(t, uuid.New(), 5))
		assert.ErrorIs(t, err, domain.ErrConflict)
	})

	t.Run("Update удалённого — ErrConflict", func(t *testing.T) {
		userID := uuid.New()
		created, err := repo.Create(ctx, mustWorkout(t, userID, 5))
		require.NoError(t, err)
		require.NoError(t, repo.MarkDeleted(ctx, userID, created.ID()))

		_, err = repo.Update(ctx, created)
		assert.ErrorIs(t, err, domain.ErrConflict)
	})

	// ============================================================
	//  MarkDeleted
	// ============================================================

	t.Run("MarkDeleted помечает тренировку удалённой", func(t *testing.T) {
		userID := uuid.New()
		created, err := repo.Create(ctx, mustWorkout(t, userID, 5))
		require.NoError(t, err)

		require.NoError(t, repo.MarkDeleted(ctx, userID, created.ID()))

		found, err := repo.GetByID(ctx, userID, created.ID())
		require.NoError(t, err)
		assert.True(t, found.IsDeleted())
		require.NotNil(t, found.UpdatedAt())
	})

	t.Run("MarkDeleted повторно — ErrWorkoutNotFound", func(t *testing.T) {
		userID := uuid.New()
		created, err := repo.Create(ctx, mustWorkout(t, userID, 5))
		require.NoError(t, err)
		require.NoError(t, repo.MarkDeleted(ctx, userID, created.ID()))

		err = repo.MarkDeleted(ctx, userID, created.ID())
		assert.ErrorIs(t, err, domain.ErrWorkoutNotFound)
	})

	t.Run("MarkDeleted несуществующего — ErrWorkoutNotFound", func(t *testing.T) {
		err := repo.MarkDeleted(ctx, uuid.New(), uuid.New())
		assert.ErrorIs(t, err, domain.ErrWorkoutNotFound)
	})

	t.Run("MarkDeleted чужого — ErrWorkoutNotFound", func(t *testing.T) {
		ownerID := uuid.New()
		otherID := uuid.New()
		created, err := repo.Create(ctx, mustWorkout(t, ownerID, 5))
		require.NoError(t, err)

		err = repo.MarkDeleted(ctx, otherID, created.ID())
		assert.ErrorIs(t, err, domain.ErrWorkoutNotFound)

		found, err := repo.GetByID(ctx, ownerID, created.ID())
		require.NoError(t, err)
		assert.False(t, found.IsDeleted())
	})

	// ============================================================
	//  GetUserScore
	// ============================================================

	t.Run("GetUserScore суммирует только completed", func(t *testing.T) {
		userID := uuid.New()

		w1 := newCompletedWorkout(t, repo, ctx, userID, 100)
		w2 := newCompletedWorkout(t, repo, ctx, userID, 200)
		_, _ = w1, w2 // проверяем только сумму

		// planned — не должна учитываться
		_, err := repo.Create(ctx, mustWorkout(t, userID, 5))
		require.NoError(t, err)

		score, err := repo.GetUserScore(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, int64(300), score)
	})

	t.Run("GetUserScore удалённые не учитываются", func(t *testing.T) {
		userID := uuid.New()
		w := newCompletedWorkout(t, repo, ctx, userID, 100)

		require.NoError(t, repo.MarkDeleted(ctx, userID, w.ID()))

		score, err := repo.GetUserScore(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), score)
	})

	t.Run("GetUserScore для пустого user — 0", func(t *testing.T) {
		score, err := repo.GetUserScore(ctx, uuid.New())
		require.NoError(t, err)
		assert.Equal(t, int64(0), score)
	})

	// ============================================================
	//  WithTx
	// ============================================================

	t.Run("WithTx commit — изменения сохранены", func(t *testing.T) {
		userID := uuid.New()

		err := repo.WithTx(ctx, func(wRepo ports.WorkoutRepository, eRepo ports.WorkoutExerciseRepository) error {
			w := mustWorkout(t, userID, 5)
			_, err := wRepo.Create(ctx, w)
			return err
		})
		require.NoError(t, err)

		list, err := repo.GetByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("WithTx rollback — изменения откатились", func(t *testing.T) {
		userID := uuid.New()
		sentinel := assert.AnError

		err := repo.WithTx(ctx, func(wRepo ports.WorkoutRepository, eRepo ports.WorkoutExerciseRepository) error {
			w := mustWorkout(t, userID, 5)
			if _, err := wRepo.Create(ctx, w); err != nil {
				return err
			}
			return sentinel
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, sentinel)

		list, err := repo.GetByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Empty(t, list, "изменения должны были откатиться")
	})
}

// ============================================================
//  Хелперы для nullable.Nullable[domain.Intensity]
// ============================================================

func nullableIntensityNone() nullable.Nullable[domain.Intensity] {
	return nullable.Nullable[domain.Intensity]{}
}

func nullableIntensitySet(v *domain.Intensity) nullable.Nullable[domain.Intensity] {
	return nullable.Nullable[domain.Intensity]{Set: true, Value: v}
}

// newCompletedWorkout создаёт тренировку в статусе completed с заданным score.
// Проходит state machine: planned → in_progress → completed (два Update).
func newCompletedWorkout(
	t *testing.T,
	repo *WorkoutRepository,
	ctx context.Context,
	userID uuid.UUID,
	score int,
) domain.Workout {
	t.Helper()

	w, err := repo.Create(ctx, mustWorkout(t, userID, 5))
	require.NoError(t, err)

	// planned → in_progress
	inProgress := domain.StatusInProgress
	require.NoError(t, w.ApplyPatch(domain.NewWorkoutPatch(&inProgress, nullableIntensityNone())))
	w, err = repo.Update(ctx, w)
	require.NoError(t, err)

	// in_progress → completed + score
	completed := domain.StatusCompleted
	require.NoError(t, w.ApplyPatch(domain.NewWorkoutPatch(&completed, nullableIntensityNone())))
	w.SetWorkoutScore(score)
	w, err = repo.Update(ctx, w)
	require.NoError(t, err)

	return w
}
