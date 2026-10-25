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
	"github.com/BladeRunner322/orange-team-microservices/pkg/nullable"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
)

func TestPatchWorkout_Execute(t *testing.T) {
	newCtx := func(userID uuid.UUID) context.Context {
		return ctxkeys.WithUserID(context.Background(), userID.String())
	}

	// setupWorkout создаёт тренировку в репо.
	setupWorkout := func(t *testing.T, repo *mockWorkoutRepository, ctx context.Context, userID uuid.UUID, status domain.Status) domain.Workout {
		t.Helper()
		w := newWorkoutInStatus(t, userID, status)
		_, err := repo.Create(ctx, w)
		require.NoError(t, err)
		return w
	}

	intensity7 := func(t *testing.T) *domain.Intensity {
		t.Helper()
		i, err := domain.NewIntensity(7)
		require.NoError(t, err)
		return &i
	}

	// patchStatus — хелпер для создания патча со статусом.
	patchStatus := func(s domain.Status) domain.WorkoutPatch {
		return domain.NewWorkoutPatch(&s, nullable.Nullable[domain.Intensity]{})
	}

	t.Run("успешно меняет intensity", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w := setupWorkout(t, repo, ctx, userID, domain.StatusPlanned)

		patch := domain.NewWorkoutPatch(nil, nullable.Nullable[domain.Intensity]{Set: true, Value: intensity7(t)})
		uc := NewPatchWorkout(repo, logger.NewTestLogger())

		updated, err := uc.Execute(ctx, w.ID(), patch)
		require.NoError(t, err)
		require.NotNil(t, updated.Intensity())
		assert.Equal(t, 7, updated.Intensity().Int())
	})

	t.Run("успешно переводит planned → in_progress", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w := setupWorkout(t, repo, ctx, userID, domain.StatusPlanned)

		uc := NewPatchWorkout(repo, logger.NewTestLogger())
		updated, err := uc.Execute(ctx, w.ID(), patchStatus(domain.StatusInProgress))

		require.NoError(t, err)
		assert.Equal(t, domain.StatusInProgress, updated.Status())
	})

	t.Run("успешно переводит in_progress → completed с хотя бы одним выполненным упражнением", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w := setupWorkout(t, repo, ctx, userID, domain.StatusInProgress)

		// Добавляем completed упражнение в репо.
		diff := *ptrDifficulty(t, 5)
		ex, err := domain.NewWorkoutExercise(
			w.ID(), uuid.New(), diff,
			ptrWeight(t, 100), ptrSets(t, 4), ptrReps(t, 10), nil,
			true, domain.ExerciseTypeWeight,
		)
		require.NoError(t, err)
		_, err = repo.exerciseRepo.Create(ctx, ex)
		require.NoError(t, err)

		uc := NewPatchWorkout(repo, logger.NewTestLogger())
		updated, err := uc.Execute(ctx, w.ID(), patchStatus(domain.StatusCompleted))

		require.NoError(t, err)
		assert.Equal(t, domain.StatusCompleted, updated.Status())
	})

	t.Run("нельзя завершить тренировку без completed упражнений", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w := setupWorkout(t, repo, ctx, userID, domain.StatusInProgress)

		uc := NewPatchWorkout(repo, logger.NewTestLogger())
		_, err := uc.Execute(ctx, w.ID(), patchStatus(domain.StatusCompleted))

		assert.ErrorIs(t, err, domain.ErrCannotCompleteEmptyWorkout)
	})

	t.Run("невалидный переход — ErrInvalidStatusTransition", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w := setupWorkout(t, repo, ctx, userID, domain.StatusPlanned)

		uc := NewPatchWorkout(repo, logger.NewTestLogger())
		_, err := uc.Execute(ctx, w.ID(), patchStatus(domain.StatusCompleted))

		assert.ErrorIs(t, err, domain.ErrInvalidStatusTransition)
	})

	t.Run("workout не найден — ErrWorkoutNotFound", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockWorkoutRepository()
		uc := NewPatchWorkout(repo, logger.NewTestLogger())

		_, err := uc.Execute(ctx, uuid.New(), patchStatus(domain.StatusInProgress))
		assert.ErrorIs(t, err, domain.ErrWorkoutNotFound)
	})

	t.Run("workout удалён — ErrWorkoutNotFound", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		w := newDeletedWorkout(t, userID)
		_, err := repo.Create(ctx, w)
		require.NoError(t, err)

		uc := NewPatchWorkout(repo, logger.NewTestLogger())
		_, err = uc.Execute(ctx, w.ID(), patchStatus(domain.StatusInProgress))
		assert.ErrorIs(t, err, domain.ErrWorkoutNotFound)
	})

	t.Run("нет user_id в context — ErrUnauthenticated", func(t *testing.T) {
		repo := newMockWorkoutRepository()
		uc := NewPatchWorkout(repo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), uuid.New(), patchStatus(domain.StatusInProgress))
		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})

	t.Run("ошибка репозитория — пробрасывается", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockWorkoutRepository()
		repo.err = errors.New("db is down")
		uc := NewPatchWorkout(repo, logger.NewTestLogger())

		_, err := uc.Execute(ctx, uuid.New(), patchStatus(domain.StatusInProgress))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}