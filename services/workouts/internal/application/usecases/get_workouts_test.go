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
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
)

func TestGetWorkouts_Execute(t *testing.T) {
	newCtx := func(userID uuid.UUID) context.Context {
		return ctxkeys.WithUserID(context.Background(), userID.String())
	}

	t.Run("возвращает только активные тренировки текущего пользователя", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()

		// 2 активные
		w1 := domain.NewWorkout(userID, 5)
		_, err := repo.Create(ctx, w1)
		require.NoError(t, err)

		w2 := domain.NewWorkout(userID, 5)
		_, err = repo.Create(ctx, w2)
		require.NoError(t, err)

		// 1 удалённая
		w3 := domain.NewWorkout(userID, 5)
		_, err = repo.Create(ctx, w3)
		require.NoError(t, err)
		err = repo.MarkDeleted(ctx, userID, w3.ID())
		require.NoError(t, err)

		// 1 чужая
		other := domain.NewWorkout(uuid.New(), 5)
		_, err = repo.Create(ctx, other)
		require.NoError(t, err)

		uc := NewGetWorkouts(repo, logger.NewTestLogger())
		result, err := uc.Execute(ctx)

		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("пустой список — не nil", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockWorkoutRepository()
		uc := NewGetWorkouts(repo, logger.NewTestLogger())

		result, err := uc.Execute(ctx)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("ошибка репозитория — пробрасывается", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockWorkoutRepository()
		repo.err = errors.New("db is down")
		uc := NewGetWorkouts(repo, logger.NewTestLogger())

		_, err := uc.Execute(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}
