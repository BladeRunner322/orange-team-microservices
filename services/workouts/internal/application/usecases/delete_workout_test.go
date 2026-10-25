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

func TestDeleteWorkout_Execute(t *testing.T) {
	newCtx := func(userID uuid.UUID) context.Context {
		return ctxkeys.WithUserID(context.Background(), userID.String())
	}

	t.Run("успешно помечает удалённой", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()

		w := domain.NewWorkout(userID, 5)
		_, err := repo.Create(ctx, w)
		require.NoError(t, err)

		uc := NewDeleteWorkout(repo, logger.NewTestLogger())
		err = uc.Execute(ctx, w.ID())
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, userID, w.ID())
		require.NoError(t, err)
		assert.True(t, found.IsDeleted())
	})

	t.Run("идемпотентность: повторный вызов не падает", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()

		w := domain.NewWorkout(userID, 5)
		_, err := repo.Create(ctx, w)
		require.NoError(t, err)

		uc := NewDeleteWorkout(repo, logger.NewTestLogger())
		require.NoError(t, uc.Execute(ctx, w.ID()))
		err = uc.Execute(ctx, w.ID())

		assert.NoError(t, err, "повторный DELETE должен быть идемпотентен")
	})

	t.Run("идемпотентность: несуществующий id не падает", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockWorkoutRepository()
		uc := NewDeleteWorkout(repo, logger.NewTestLogger())

		err := uc.Execute(ctx, uuid.New())
		assert.NoError(t, err)
	})

	t.Run("нет user_id в context — ErrUnauthenticated", func(t *testing.T) {
		repo := newMockWorkoutRepository()
		uc := NewDeleteWorkout(repo, logger.NewTestLogger())

		err := uc.Execute(context.Background(), uuid.New())
		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})

	t.Run("ошибка репозитория — пробрасывается", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockWorkoutRepository()
		repo.err = errors.New("db is down")
		uc := NewDeleteWorkout(repo, logger.NewTestLogger())

		err := uc.Execute(ctx, uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}
