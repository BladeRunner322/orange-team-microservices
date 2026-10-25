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
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
)

func TestPatchWorkoutExercise_Execute(t *testing.T) {
	newCtx := func(userID uuid.UUID) context.Context {
		return ctxkeys.WithUserID(context.Background(), userID.String())
	}

	ptrBool := func(v bool) *bool { return &v }

	// setup создаёт тренировку и упражнение в репо.
	setup := func(
		t *testing.T,
		repo *mockWorkoutRepository,
		ctx context.Context,
		userID uuid.UUID,
		status domain.Status,
		exCompleted bool,
	) (domain.Workout, domain.WorkoutExercise) {
		t.Helper()
		w := newWorkoutInStatus(t, userID, status)
		_, err := repo.Create(ctx, w)
		require.NoError(t, err)

		diff := *ptrDifficulty(t, 5)
		ex, err := domain.NewWorkoutExercise(
			w.ID(), uuid.New(), diff,
			ptrWeight(t, 100), ptrSets(t, 4), ptrReps(t, 10), nil,
			exCompleted, domain.ExerciseTypeWeight,
		)
		require.NoError(t, err)
		_, err = repo.exerciseRepo.Create(ctx, ex)
		require.NoError(t, err)

		return w, ex
	}

	weightInfo := ports.ExerciseInfo{Type: "weight", Difficulty: 5}

	t.Run("успешно меняет weight и пересчитывает load", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w, ex := setup(t, repo, ctx, userID, domain.StatusPlanned, false)
		exercisesClient := &mockExercisesClient{info: weightInfo}

		patch := domain.NewWorkoutExercisePatch(ptrWeight(t, 200), nil, nil, nil, nil)
		uc := NewPatchWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())

		updated, err := uc.Execute(ctx, w.ID(), ex.ID(), patch)
		require.NoError(t, err)
		assert.Equal(t, 200, updated.Weight().Int())
		assert.Equal(t, 8000, updated.ExerciseLoad()) // 200 × 10 × 4

		// score пересчитан (не completed → 0)
		updW, err := repo.GetByID(ctx, userID, w.ID())
		require.NoError(t, err)
		assert.Equal(t, 0, updW.WorkoutScore())
	})

	t.Run("успешно отмечает completed и пересчитывает score", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w, ex := setup(t, repo, ctx, userID, domain.StatusInProgress, false)
		exercisesClient := &mockExercisesClient{info: weightInfo}

		patch := domain.NewWorkoutExercisePatch(nil, nil, nil, nil, ptrBool(true))
		uc := NewPatchWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())

		updated, err := uc.Execute(ctx, w.ID(), ex.ID(), patch)
		require.NoError(t, err)
		assert.True(t, updated.Completed())

		// score пересчитан: effectiveLoad = 4000 (load) × 5 (difficulty) = 20000
		updW, err := repo.GetByID(ctx, userID, w.ID())
		require.NoError(t, err)
		expectedScore := domain.CalculateScore(4000*5, 5)
		assert.Equal(t, expectedScore, updW.WorkoutScore())
	})

	t.Run("нельзя completed=true в planned — ErrCannotMarkExerciseCompleted", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w, ex := setup(t, repo, ctx, userID, domain.StatusPlanned, false)
		exercisesClient := &mockExercisesClient{info: weightInfo}

		patch := domain.NewWorkoutExercisePatch(nil, nil, nil, nil, ptrBool(true))
		uc := NewPatchWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())

		_, err := uc.Execute(ctx, w.ID(), ex.ID(), patch)
		assert.ErrorIs(t, err, domain.ErrCannotMarkExerciseCompleted)
	})

	t.Run("workout не найден — ErrWorkoutNotFound", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockWorkoutRepository()
		exercisesClient := &mockExercisesClient{info: weightInfo}
		uc := NewPatchWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())

		_, err := uc.Execute(ctx, uuid.New(), uuid.New(), domain.NewWorkoutExercisePatch(nil, nil, nil, nil, nil))
		assert.ErrorIs(t, err, domain.ErrWorkoutNotFound)
	})

	t.Run("workout удалён — ErrWorkoutNotFound", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w := newDeletedWorkout(t, userID)
		_, err := repo.Create(ctx, w)
		require.NoError(t, err)
		exercisesClient := &mockExercisesClient{info: weightInfo}
		uc := NewPatchWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())

		_, err = uc.Execute(ctx, w.ID(), uuid.New(), domain.NewWorkoutExercisePatch(nil, nil, nil, nil, nil))
		assert.ErrorIs(t, err, domain.ErrWorkoutNotFound)
	})

	t.Run("workout в терминальном статусе — ErrCannotModifyWorkoutExercise", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w, ex := setup(t, repo, ctx, userID, domain.StatusCompleted, false)
		exercisesClient := &mockExercisesClient{info: weightInfo}
		uc := NewPatchWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())

		_, err := uc.Execute(ctx, w.ID(), ex.ID(), domain.NewWorkoutExercisePatch(nil, nil, nil, nil, nil))
		assert.ErrorIs(t, err, domain.ErrCannotModifyWorkoutExercise)
	})

	t.Run("упражнение не найдено — ErrWorkoutExerciseNotFound", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w, _ := setup(t, repo, ctx, userID, domain.StatusPlanned, false)
		exercisesClient := &mockExercisesClient{info: weightInfo}
		uc := NewPatchWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())

		_, err := uc.Execute(ctx, w.ID(), uuid.New(), domain.NewWorkoutExercisePatch(nil, nil, nil, nil, nil))
		assert.ErrorIs(t, err, domain.ErrWorkoutExerciseNotFound)
	})

	t.Run("ошибка Exercises — пробрасывается", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w, ex := setup(t, repo, ctx, userID, domain.StatusPlanned, false)
		exercisesClient := &mockExercisesClient{err: errors.New("exercises is down")}
		uc := NewPatchWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())

		_, err := uc.Execute(ctx, w.ID(), ex.ID(), domain.NewWorkoutExercisePatch(nil, nil, nil, nil, nil))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exercises is down")
	})

	t.Run("нет user_id в context — ErrUnauthenticated", func(t *testing.T) {
		repo := newMockWorkoutRepository()
		exercisesClient := &mockExercisesClient{info: weightInfo}
		uc := NewPatchWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), domain.NewWorkoutExercisePatch(nil, nil, nil, nil, nil))
		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})
}
