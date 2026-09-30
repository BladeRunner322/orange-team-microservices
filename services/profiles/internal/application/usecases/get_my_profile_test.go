package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BladeRunner322/orange-team-microservices/pkg/grpc/authctx"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
)

func TestGetMyProfile_Execute(t *testing.T) {
	t.Run("успешно возвращает профиль из context", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		ctx := authctx.WithUserID(context.Background(), userID.String())
		profile := domain.NewEmptyProfile(userID)
		repo := newMockRepository()
		_ = repo.Upsert(ctx, profile)
		uc := NewGetMyProfile(repo, logger.NewTestLogger())

		// Act
		result, err := uc.Execute(ctx)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, userID, result.UserID())
	})

	t.Run("lazy-create, если профиля нет", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		ctx := authctx.WithUserID(context.Background(), userID.String())
		repo := newMockRepository()
		uc := NewGetMyProfile(repo, logger.NewTestLogger())

		// Act
		result, err := uc.Execute(ctx)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, userID, result.UserID())
		assert.False(t, result.Completed())
	})

	t.Run("нет user_id в context — ErrUnauthenticated", func(t *testing.T) {
		// Arrange
		ctx := context.Background() // ← без authctx.WithUserID
		repo := newMockRepository()
		uc := NewGetMyProfile(repo, logger.NewTestLogger())

		// Act
		_, err := uc.Execute(ctx)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})

	t.Run("невалидный user_id в context — ErrUnauthenticated", func(t *testing.T) {
		// Arrange
		ctx := authctx.WithUserID(context.Background(), "not-a-uuid")
		repo := newMockRepository()
		uc := NewGetMyProfile(repo, logger.NewTestLogger())

		// Act
		_, err := uc.Execute(ctx)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})

	t.Run("ошибка репозитория — пробрасывается", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		ctx := authctx.WithUserID(context.Background(), userID.String())
		repo := newMockRepository()
		repo.err = errors.New("db is down")
		uc := NewGetMyProfile(repo, logger.NewTestLogger())

		// Act
		_, err := uc.Execute(ctx)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}
