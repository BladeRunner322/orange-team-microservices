//go:build integration
// +build integration

package postgres_repo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
)

func TestWorkoutExerciseRepository_Integration(t *testing.T) {
	workoutRepo := NewWorkoutRepository(testPool)
	exerciseRepo := NewWorkoutExerciseRepository(testPool)
	runWorkoutExerciseRepositoryTests(t, exerciseRepo, workoutRepo)
}

// mustWeightExercise создаёт валидное weight-упражнение или падает.
func mustWeightExercise(t *testing.T, workoutID uuid.UUID) domain.WorkoutExercise {
	t.Helper()

	difficulty, err := domain.NewDifficulty(5)
	require.NoError(t, err)

	weight, err := domain.NewWeight(100)
	require.NoError(t, err)

	sets, err := domain.NewSets(4)
	require.NoError(t, err)

	reps, err := domain.NewReps(10)
	require.NoError(t, err)

	e, err := domain.NewWorkoutExercise(
		workoutID, uuid.New(), difficulty,
		&weight, &sets, &reps, nil,
		false, domain.ExerciseTypeWeight,
	)
	require.NoError(t, err)

	return e
}

// mustDurationExercise создаёт валидное duration-упражнение или падает.
func mustDurationExercise(t *testing.T, workoutID uuid.UUID) domain.WorkoutExercise {
	t.Helper()

	difficulty, err := domain.NewDifficulty(3)
	require.NoError(t, err)

	duration, err := domain.NewDuration(600)
	require.NoError(t, err)

	e, err := domain.NewWorkoutExercise(
		workoutID, uuid.New(), difficulty,
		nil, nil, nil, &duration,
		false, domain.ExerciseTypeDuration,
	)
	require.NoError(t, err)

	return e
}

func runWorkoutExerciseRepositoryTests(
	t *testing.T,
	repo *WorkoutExerciseRepository,
	workoutRepo *WorkoutRepository,
) {
	ctx := context.Background()

	// newWorkout создаёт реальную тренировку (нужна для FK) и возвращает её ID.
	newWorkout := func(t *testing.T) uuid.UUID {
		t.Helper()
		w, err := workoutRepo.Create(ctx, mustWorkout(t, uuid.New(), 5))
		require.NoError(t, err)
		return w.ID()
	}

	// ============================================================
	//  Create
	// ============================================================

	t.Run("Create сохраняет weight-упражнение и возвращает из RETURNING", func(t *testing.T) {
		workoutID := newWorkout(t)
		e := mustWeightExercise(t, workoutID)

		created, err := repo.Create(ctx, e)
		require.NoError(t, err)

		assert.Equal(t, e.ID(), created.ID())
		assert.Equal(t, workoutID, created.WorkoutID())
		assert.Equal(t, e.ExerciseID(), created.ExerciseID())
		assert.Equal(t, 5, created.Difficulty().Int())
		require.NotNil(t, created.Weight())
		assert.Equal(t, 100, created.Weight().Int())
		require.NotNil(t, created.Sets())
		assert.Equal(t, 4, created.Sets().Int())
		require.NotNil(t, created.Reps())
		assert.Equal(t, 10, created.Reps().Int())
		assert.Nil(t, created.Duration())
		assert.False(t, created.Completed())
		assert.Equal(t, 4000, created.ExerciseLoad()) // 100×10×4
		assert.False(t, created.CreatedAt().IsZero())
		assert.Nil(t, created.UpdatedAt())
	})

	t.Run("Create сохраняет duration-упражнение", func(t *testing.T) {
		workoutID := newWorkout(t)
		e := mustDurationExercise(t, workoutID)

		created, err := repo.Create(ctx, e)
		require.NoError(t, err)

		assert.Nil(t, created.Weight())
		assert.Nil(t, created.Sets())
		assert.Nil(t, created.Reps())
		require.NotNil(t, created.Duration())
		assert.Equal(t, 600, created.Duration().Int())
		assert.Equal(t, 600, created.ExerciseLoad())
	})

	t.Run("Create с несуществующим workout_id — FK violation", func(t *testing.T) {
		e := mustWeightExercise(t, uuid.New()) // workout с таким ID нет

		_, err := repo.Create(ctx, e)
		require.Error(t, err)
		// не проверяем конкретный sentinel — FK-ошибка может быть обёрнута
		// pgerrors.ErrViolatesForeignKey, а может прийти как ErrUnknown
	})

	// ============================================================
	//  GetByID
	// ============================================================

	t.Run("GetByID возвращает упражнение", func(t *testing.T) {
		workoutID := newWorkout(t)
		created, err := repo.Create(ctx, mustWeightExercise(t, workoutID))
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, workoutID, created.ID())
		require.NoError(t, err)

		assert.Equal(t, created.ID(), found.ID())
		assert.Equal(t, workoutID, found.WorkoutID())
	})

	t.Run("GetByID не найдено — ErrWorkoutExerciseNotFound", func(t *testing.T) {
		_, err := repo.GetByID(ctx, uuid.New(), uuid.New())
		assert.ErrorIs(t, err, domain.ErrWorkoutExerciseNotFound)
	})

	t.Run("GetByID с чужим workout_id — ErrWorkoutExerciseNotFound", func(t *testing.T) {
		workoutID := newWorkout(t)
		otherWorkoutID := newWorkout(t)
		created, err := repo.Create(ctx, mustWeightExercise(t, workoutID))
		require.NoError(t, err)

		// упражнение существует, но в другой тренировке
		_, err = repo.GetByID(ctx, otherWorkoutID, created.ID())
		assert.ErrorIs(t, err, domain.ErrWorkoutExerciseNotFound)
	})

	// ============================================================
	//  ListByWorkout
	// ============================================================

	t.Run("ListByWorkout возвращает все упражнения тренировки", func(t *testing.T) {
		workoutID := newWorkout(t)
		otherWorkoutID := newWorkout(t)

		e1, err := repo.Create(ctx, mustWeightExercise(t, workoutID))
		require.NoError(t, err)
		e2, err := repo.Create(ctx, mustDurationExercise(t, workoutID))
		require.NoError(t, err)

		// чужое упражнение — не должно попасть
		_, err = repo.Create(ctx, mustWeightExercise(t, otherWorkoutID))
		require.NoError(t, err)

		list, err := repo.ListByWorkout(ctx, workoutID)
		require.NoError(t, err)
		require.Len(t, list, 2)

		ids := []uuid.UUID{list[0].ID(), list[1].ID()}
		assert.Contains(t, ids, e1.ID())
		assert.Contains(t, ids, e2.ID())
	})

	t.Run("ListByWorkout пустой slice, не nil", func(t *testing.T) {
		list, err := repo.ListByWorkout(ctx, uuid.New())
		require.NoError(t, err)
		assert.NotNil(t, list)
		assert.Empty(t, list)
	})

	// ============================================================
	//  Update
	// ============================================================

	t.Run("Update меняет поля и инкрементирует version", func(t *testing.T) {
		workoutID := newWorkout(t)
		created, err := repo.Create(ctx, mustWeightExercise(t, workoutID))
		require.NoError(t, err)

		// патчим weight 100 → 200
		newWeight, err := domain.NewWeight(200)
		require.NoError(t, err)
		require.NoError(t, created.ApplyPatch(
			domain.NewWorkoutExercisePatch(&newWeight, nil, nil, nil, nil),
			domain.ExerciseTypeWeight,
		))

		updated, err := repo.Update(ctx, created)
		require.NoError(t, err)

		require.NotNil(t, updated.Weight())
		assert.Equal(t, 200, updated.Weight().Int())
		assert.Equal(t, 8000, updated.ExerciseLoad()) // 200×10×4
		assert.Equal(t, 1, updated.Version())
		require.NotNil(t, updated.UpdatedAt())
	})

	t.Run("Update помечает completed", func(t *testing.T) {
		workoutID := newWorkout(t)
		created, err := repo.Create(ctx, mustWeightExercise(t, workoutID))
		require.NoError(t, err)

		completed := true
		require.NoError(t, created.ApplyPatch(
			domain.NewWorkoutExercisePatch(nil, nil, nil, nil, &completed),
			domain.ExerciseTypeWeight,
		))

		updated, err := repo.Update(ctx, created)
		require.NoError(t, err)
		assert.True(t, updated.Completed())
	})

	t.Run("Update с устаревшей version — ErrConflict", func(t *testing.T) {
		workoutID := newWorkout(t)
		created, err := repo.Create(ctx, mustWeightExercise(t, workoutID))
		require.NoError(t, err)

		_, err = repo.Update(ctx, created)
		require.NoError(t, err)

		_, err = repo.Update(ctx, created)
		assert.ErrorIs(t, err, domain.ErrConflict)
	})

	t.Run("Update несуществующего — ErrConflict", func(t *testing.T) {
		workoutID := newWorkout(t)
		// создаём упражнение в памяти, но не вставляем в БД
		e := mustWeightExercise(t, workoutID)

		_, err := repo.Update(ctx, e)
		assert.ErrorIs(t, err, domain.ErrConflict)
	})

	t.Run("Update не трогает difficulty и exercise_id", func(t *testing.T) {
		workoutID := newWorkout(t)
		created, err := repo.Create(ctx, mustWeightExercise(t, workoutID))
		require.NoError(t, err)

		origDifficulty := created.Difficulty().Int()
		origExerciseID := created.ExerciseID()

		newWeight, err := domain.NewWeight(150)
		require.NoError(t, err)
		require.NoError(t, created.ApplyPatch(
			domain.NewWorkoutExercisePatch(&newWeight, nil, nil, nil, nil),
			domain.ExerciseTypeWeight,
		))

		updated, err := repo.Update(ctx, created)
		require.NoError(t, err)

		assert.Equal(t, origDifficulty, updated.Difficulty().Int())
		assert.Equal(t, origExerciseID, updated.ExerciseID())
	})

	// ============================================================
	//  Delete
	// ============================================================

	t.Run("Delete удаляет упражнение", func(t *testing.T) {
		workoutID := newWorkout(t)
		created, err := repo.Create(ctx, mustWeightExercise(t, workoutID))
		require.NoError(t, err)

		require.NoError(t, repo.Delete(ctx, workoutID, created.ID()))

		_, err = repo.GetByID(ctx, workoutID, created.ID())
		assert.ErrorIs(t, err, domain.ErrWorkoutExerciseNotFound)
	})

	t.Run("Delete повторно — ErrWorkoutExerciseNotFound", func(t *testing.T) {
		workoutID := newWorkout(t)
		created, err := repo.Create(ctx, mustWeightExercise(t, workoutID))
		require.NoError(t, err)
		require.NoError(t, repo.Delete(ctx, workoutID, created.ID()))

		err = repo.Delete(ctx, workoutID, created.ID())
		assert.ErrorIs(t, err, domain.ErrWorkoutExerciseNotFound)
	})

	t.Run("Delete несуществующего — ErrWorkoutExerciseNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, uuid.New(), uuid.New())
		assert.ErrorIs(t, err, domain.ErrWorkoutExerciseNotFound)
	})

	t.Run("Delete с чужим workout_id — ErrWorkoutExerciseNotFound", func(t *testing.T) {
		workoutID := newWorkout(t)
		otherWorkoutID := newWorkout(t)
		created, err := repo.Create(ctx, mustWeightExercise(t, workoutID))
		require.NoError(t, err)

		err = repo.Delete(ctx, otherWorkoutID, created.ID())
		assert.ErrorIs(t, err, domain.ErrWorkoutExerciseNotFound)

		// упражнение не тронуто
		found, err := repo.GetByID(ctx, workoutID, created.ID())
		require.NoError(t, err)
		assert.Equal(t, created.ID(), found.ID())
	})
}
