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
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
)

func TestDeleteMyProfile_Execute(t *testing.T) {
	t.Run("успешно удаляет профиль", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		ctx := ctxkeys.WithUserID(context.Background(), userID.String())
		profile := domain.NewEmptyProfile(userID)
		repo := newMockRepository()
		_ = repo.Upsert(ctx, profile)
		uc := NewDeleteMyProfile(repo, logger.NewTestLogger())

		// Act
		err := uc.Execute(ctx)

		// Assert
		require.NoError(t, err)

		// профиль реально удалён из репозитория
		_, err = repo.GetByUserID(ctx, userID)
		assert.ErrorIs(t, err, domain.ErrProfileNotFound)
	})

	t.Run("профиль не найден — ErrProfileNotFound", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		ctx := ctxkeys.WithUserID(context.Background(), userID.String())
		repo := newMockRepository()
		uc := NewDeleteMyProfile(repo, logger.NewTestLogger())

		// Act
		err := uc.Execute(ctx)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrProfileNotFound)
	})

	t.Run("нет user_id в context — ErrUnauthenticated", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		repo := newMockRepository()
		uc := NewDeleteMyProfile(repo, logger.NewTestLogger())

		// Act
		err := uc.Execute(ctx)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})

	t.Run("ошибка репозитория — пробрасывается", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		ctx := ctxkeys.WithUserID(context.Background(), userID.String())
		repo := newMockRepository()
		repo.err = errors.New("db is down")
		uc := NewDeleteMyProfile(repo, logger.NewTestLogger())

		// Act
		err := uc.Execute(ctx)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}
