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

func TestCreateWorkoutExercise_Execute(t *testing.T) {
	newCtx := func(userID uuid.UUID) context.Context {
		return ctxkeys.WithUserID(context.Background(), userID.String())
	}

	// setup создаёт тренировку в репо и возвращает её.
	setupWorkout := func(t *testing.T, repo *mockWorkoutRepository, ctx context.Context, userID uuid.UUID, status domain.Status) domain.Workout {
		t.Helper()
		w := newWorkoutInStatus(t, userID, status)
		_, err := repo.Create(ctx, w)
		require.NoError(t, err)
		return w
	}

	// weightInfo — exerciseInfo для weight-упражнения.
	weightInfo := ports.ExerciseInfo{Type: "weight", Difficulty: 5}
	durationInfo := ports.ExerciseInfo{Type: "duration", Difficulty: 3}

	t.Run("успешно создаёт completed weight-упражнение (score пересчитан)", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w := setupWorkout(t, repo, ctx, userID, domain.StatusInProgress)
		exercisesClient := &mockExercisesClient{info: weightInfo}

		uc := NewCreateWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())
		created, err := uc.Execute(ctx, w.ID(), uuid.New(), ptrWeight(t, 100), ptrSets(t, 4), ptrReps(t, 10), nil, true)

		require.NoError(t, err)
		assert.Equal(t, 100, created.Weight().Int())
		assert.Equal(t, 4000, created.ExerciseLoad()) // 100 × 10 × 4

		// score пересчитан в тренировке
		updated, err := repo.GetByID(ctx, userID, w.ID())
		require.NoError(t, err)
		// effective_load = exercise_load (4000) × difficulty (5) = 20000
		expectedScore := domain.CalculateScore(4000*5, 5)
		assert.Equal(t, expectedScore, updated.WorkoutScore())
	})

	t.Run("успешно создаёт weight-упражнение (не completed, score = 0)", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w := setupWorkout(t, repo, ctx, userID, domain.StatusPlanned)
		exercisesClient := &mockExercisesClient{info: weightInfo}

		uc := NewCreateWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())
		created, err := uc.Execute(ctx, w.ID(), uuid.New(), ptrWeight(t, 100), ptrSets(t, 4), ptrReps(t, 10), nil, false)

		require.NoError(t, err)
		assert.Equal(t, 4000, created.ExerciseLoad())

		updated, err := repo.GetByID(ctx, userID, w.ID())
		require.NoError(t, err)
		assert.Equal(t, 0, updated.WorkoutScore(), "не completed упражнение не влияет на score")
	})

	t.Run("успешно создаёт duration-упражнение", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w := setupWorkout(t, repo, ctx, userID, domain.StatusPlanned)
		exercisesClient := &mockExercisesClient{info: durationInfo}

		uc := NewCreateWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())
		created, err := uc.Execute(ctx, w.ID(), uuid.New(), nil, nil, nil, ptrDuration(t, 600), false)

		require.NoError(t, err)
		assert.Equal(t, 600, created.Duration().Int())
		assert.Equal(t, 600, created.ExerciseLoad())
	})

	t.Run("completed=true в in_progress — ок", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w := setupWorkout(t, repo, ctx, userID, domain.StatusInProgress)
		exercisesClient := &mockExercisesClient{info: weightInfo}

		uc := NewCreateWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())
		created, err := uc.Execute(ctx, w.ID(), uuid.New(), ptrWeight(t, 100), ptrSets(t, 4), ptrReps(t, 10), nil, true)

		require.NoError(t, err)
		assert.True(t, created.Completed())
	})

	t.Run("completed=true в planned — ErrCannotMarkExerciseCompleted", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w := setupWorkout(t, repo, ctx, userID, domain.StatusPlanned)
		exercisesClient := &mockExercisesClient{info: weightInfo}

		uc := NewCreateWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())
		_, err := uc.Execute(ctx, w.ID(), uuid.New(), ptrWeight(t, 100), ptrSets(t, 4), ptrReps(t, 10), nil, true)

		assert.ErrorIs(t, err, domain.ErrCannotMarkExerciseCompleted)
	})

	t.Run("workout не найден — ErrWorkoutNotFound", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockWorkoutRepository()
		exercisesClient := &mockExercisesClient{info: weightInfo}

		uc := NewCreateWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())
		_, err := uc.Execute(ctx, uuid.New(), uuid.New(), ptrWeight(t, 100), ptrSets(t, 4), ptrReps(t, 10), nil, false)

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

		uc := NewCreateWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())
		_, err = uc.Execute(ctx, w.ID(), uuid.New(), ptrWeight(t, 100), ptrSets(t, 4), ptrReps(t, 10), nil, false)

		assert.ErrorIs(t, err, domain.ErrWorkoutNotFound)
	})

	t.Run("workout завершён — ErrCannotModifyWorkoutExercise", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w := setupWorkout(t, repo, ctx, userID, domain.StatusCompleted)
		exercisesClient := &mockExercisesClient{info: weightInfo}

		uc := NewCreateWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())
		_, err := uc.Execute(ctx, w.ID(), uuid.New(), ptrWeight(t, 100), ptrSets(t, 4), ptrReps(t, 10), nil, false)

		assert.ErrorIs(t, err, domain.ErrCannotModifyWorkoutExercise)
	})

	t.Run("ошибка Exercises — пробрасывается", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w := setupWorkout(t, repo, ctx, userID, domain.StatusPlanned)
		exercisesClient := &mockExercisesClient{err: errors.New("exercises is down")}

		uc := NewCreateWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())
		_, err := uc.Execute(ctx, w.ID(), uuid.New(), ptrWeight(t, 100), ptrSets(t, 4), ptrReps(t, 10), nil, false)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "exercises is down")
	})

	t.Run("невалидный type из Exercises — ошибка", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w := setupWorkout(t, repo, ctx, userID, domain.StatusPlanned)
		exercisesClient := &mockExercisesClient{info: ports.ExerciseInfo{Type: "bogus", Difficulty: 5}}

		uc := NewCreateWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())
		_, err := uc.Execute(ctx, w.ID(), uuid.New(), ptrWeight(t, 100), ptrSets(t, 4), ptrReps(t, 10), nil, false)

		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidExerciseType)
	})

	t.Run("поля не соответствуют типу — ошибка NewWorkoutExercise", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w := setupWorkout(t, repo, ctx, userID, domain.StatusPlanned)
		// duration-упражнение, но передали weight-поля
		exercisesClient := &mockExercisesClient{info: durationInfo}

		uc := NewCreateWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())
		_, err := uc.Execute(ctx, w.ID(), uuid.New(), ptrWeight(t, 100), ptrSets(t, 4), ptrReps(t, 10), nil, false)

		assert.ErrorIs(t, err, domain.ErrInvalidWorkoutExerciseForType)
	})

	t.Run("ошибка репозитория упражнений при Create — пробрасывается", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w := setupWorkout(t, repo, ctx, userID, domain.StatusPlanned)
		repo.exerciseRepo.err = errors.New("insert failed")
		exercisesClient := &mockExercisesClient{info: weightInfo}

		uc := NewCreateWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())
		_, err := uc.Execute(ctx, w.ID(), uuid.New(), ptrWeight(t, 100), ptrSets(t, 4), ptrReps(t, 10), nil, false)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "insert failed")
	})

	t.Run("нет user_id в context — ErrUnauthenticated", func(t *testing.T) {
		repo := newMockWorkoutRepository()
		exercisesClient := &mockExercisesClient{info: weightInfo}
		uc := NewCreateWorkoutExercise(repo, repo.exerciseRepo, exercisesClient, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), nil, nil, nil, nil, false)
		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})
}
