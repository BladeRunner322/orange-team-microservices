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

func TestGetWorkout_Execute(t *testing.T) {
	newCtx := func(userID uuid.UUID) context.Context {
		return ctxkeys.WithUserID(context.Background(), userID.String())
	}

	t.Run("успешно возвращает тренировку", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()

		w := domain.NewWorkout(userID, 5)
		_, err := repo.Create(ctx, w)
		require.NoError(t, err)

		uc := NewGetWorkout(repo, logger.NewTestLogger())
		result, err := uc.Execute(ctx, w.ID())

		require.NoError(t, err)
		assert.Equal(t, w.ID(), result.ID())
	})

	t.Run("не найдено — ErrWorkoutNotFound", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockWorkoutRepository()
		uc := NewGetWorkout(repo, logger.NewTestLogger())

		_, err := uc.Execute(ctx, uuid.New())
		assert.ErrorIs(t, err, domain.ErrWorkoutNotFound)
	})

	t.Run("чужая тренировка — ErrWorkoutNotFound", func(t *testing.T) {
		ownerID := uuid.New()
		otherID := uuid.New()
		ctx := newCtx(otherID)
		repo := newMockWorkoutRepository()

		w := domain.NewWorkout(ownerID, 5)
		_, err := repo.Create(context.Background(), w)
		require.NoError(t, err)

		uc := NewGetWorkout(repo, logger.NewTestLogger())
		_, err = uc.Execute(ctx, w.ID())

		assert.ErrorIs(t, err, domain.ErrWorkoutNotFound)
	})

	t.Run("нет user_id в context — ErrUnauthenticated", func(t *testing.T) {
		repo := newMockWorkoutRepository()
		uc := NewGetWorkout(repo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), uuid.New())
		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})

	t.Run("ошибка репозитория — пробрасывается", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockWorkoutRepository()
		repo.err = errors.New("db is down")
		uc := NewGetWorkout(repo, logger.NewTestLogger())

		_, err := uc.Execute(ctx, uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}
