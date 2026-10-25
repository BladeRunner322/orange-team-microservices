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
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
)

func TestGetUserScore_Execute(t *testing.T) {
	newCtx := func(userID uuid.UUID) context.Context {
		return ctxkeys.WithUserID(context.Background(), userID.String())
	}

	// newCompletedWorkout — completed-тренировка с заданным score.
	newCompletedWorkout := func(userID uuid.UUID, score int) domain.Workout {
		now := time.Now().UTC()
		return domain.RestoreWorkout(
			uuid.New(), 0, userID, domain.StatusCompleted,
			nil, nil, score, nil, 5,
			false, now, now,
		)
	}

	t.Run("суммирует score только completed тренировок", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()

		_, err := repo.Create(ctx, newCompletedWorkout(userID, 100))
		require.NoError(t, err)
		_, err = repo.Create(ctx, newCompletedWorkout(userID, 200))
		require.NoError(t, err)

		// planned — не должна учитываться
		_, err = repo.Create(ctx, domain.NewWorkout(userID, 5))
		require.NoError(t, err)

		uc := NewGetUserScore(repo, logger.NewTestLogger())
		score, err := uc.Execute(ctx)

		require.NoError(t, err)
		assert.Equal(t, int64(300), score)
	})

	t.Run("удалённые тренировки не учитываются", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()

		w := newCompletedWorkout(userID, 100)
		_, err := repo.Create(ctx, w)
		require.NoError(t, err)
		err = repo.MarkDeleted(ctx, userID, w.ID())
		require.NoError(t, err)

		uc := NewGetUserScore(repo, logger.NewTestLogger())
		score, err := uc.Execute(ctx)

		require.NoError(t, err)
		assert.Equal(t, int64(0), score)
	})

	t.Run("нет user_id в context — ErrUnauthenticated", func(t *testing.T) {
		repo := newMockWorkoutRepository()
		uc := NewGetUserScore(repo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background())
		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})

	t.Run("ошибка репозитория — пробрасывается", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockWorkoutRepository()
		repo.err = errors.New("db is down")
		uc := NewGetUserScore(repo, logger.NewTestLogger())

		_, err := uc.Execute(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}
