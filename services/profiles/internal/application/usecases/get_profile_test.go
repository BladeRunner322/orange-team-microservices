package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProfile_Execute(t *testing.T) {
	t.Run("успешно возвращает существующий профиль", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		userID := uuid.New()
		profile := domain.NewEmptyProfile(userID)
		repo := newMockRepository()
		_ = repo.Upsert(ctx, profile)
		uc := NewGetProfile(repo, logger.NewTestLogger())

		// Act
		result, err := uc.Execute(ctx, userID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, userID, result.UserID())
	})

	t.Run("создаёт пустой, если профиля нет", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		userID := uuid.New()
		repo := newMockRepository()
		uc := NewGetProfile(repo, logger.NewTestLogger())

		// Act
		result, err := uc.Execute(ctx, userID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, userID, result.UserID())
		assert.False(t, result.Completed())
	})

	t.Run("ошибка репозитория — пробрасывается", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		userID := uuid.New()
		repo := newMockRepository()
		repo.err = errors.New("db is down")
		uc := NewGetProfile(repo, logger.NewTestLogger())

		// Act
		_, err := uc.Execute(ctx, userID)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}
