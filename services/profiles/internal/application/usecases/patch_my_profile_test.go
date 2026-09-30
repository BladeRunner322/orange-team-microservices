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
	"github.com/BladeRunner322/orange-team-microservices/pkg/nullable"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
)

func TestPatchMyProfile_Execute(t *testing.T) {
	t.Run("успешно устанавливает вес", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		ctx := authctx.WithUserID(context.Background(), userID.String())
		profile := domain.NewEmptyProfile(userID)
		repo := newMockRepository()
		_ = repo.Upsert(ctx, profile)

		weight, err := domain.NewWeightFromKilograms(80.5)
		require.NoError(t, err)

		patch := domain.NewProfilePatch(
			nullable.Nullable[domain.Sex]{},
			nullable.Nullable[domain.Weight]{Set: true, Value: &weight},
			nullable.Nullable[domain.BirthDate]{},
			nullable.Nullable[domain.Height]{},
		)
		uc := NewPatchMyProfile(repo, logger.NewTestLogger())

		// Act
		result, err := uc.Execute(ctx, patch)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result.Weight())
		assert.Equal(t, 80.5, result.Weight().Kilograms())
	})

	t.Run("убирает вес (Set: true, Value: nil)", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		ctx := authctx.WithUserID(context.Background(), userID.String())

		weight, err := domain.NewWeightFromKilograms(80.5)
		require.NoError(t, err)

		profile := domain.NewEmptyProfile(userID)
		profile.ApplyPatch(domain.NewProfilePatch(
			nullable.Nullable[domain.Sex]{},
			nullable.Nullable[domain.Weight]{Set: true, Value: &weight},
			nullable.Nullable[domain.BirthDate]{},
			nullable.Nullable[domain.Height]{},
		))
		require.NotNil(t, profile.Weight())

		repo := newMockRepository()
		_ = repo.Upsert(ctx, profile)

		// патч: убрать вес
		patch := domain.NewProfilePatch(
			nullable.Nullable[domain.Sex]{},
			nullable.Nullable[domain.Weight]{Set: true, Value: nil},
			nullable.Nullable[domain.BirthDate]{},
			nullable.Nullable[domain.Height]{},
		)
		uc := NewPatchMyProfile(repo, logger.NewTestLogger())

		// Act
		result, err := uc.Execute(ctx, patch)

		// Assert
		require.NoError(t, err)
		assert.Nil(t, result.Weight())
	})

	t.Run("профиль не найден — ErrProfileNotFound", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		ctx := authctx.WithUserID(context.Background(), userID.String())
		repo := newMockRepository()
		uc := NewPatchMyProfile(repo, logger.NewTestLogger())

		patch := domain.NewProfilePatch(
			nullable.Nullable[domain.Sex]{},
			nullable.Nullable[domain.Weight]{},
			nullable.Nullable[domain.BirthDate]{},
			nullable.Nullable[domain.Height]{},
		)

		// Act
		_, err := uc.Execute(ctx, patch)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrProfileNotFound)
	})

	t.Run("нет user_id в context — ErrUnauthenticated", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		repo := newMockRepository()
		uc := NewPatchMyProfile(repo, logger.NewTestLogger())

		patch := domain.NewProfilePatch(
			nullable.Nullable[domain.Sex]{},
			nullable.Nullable[domain.Weight]{},
			nullable.Nullable[domain.BirthDate]{},
			nullable.Nullable[domain.Height]{},
		)

		// Act
		_, err := uc.Execute(ctx, patch)

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
		uc := NewPatchMyProfile(repo, logger.NewTestLogger())

		patch := domain.NewProfilePatch(
			nullable.Nullable[domain.Sex]{},
			nullable.Nullable[domain.Weight]{},
			nullable.Nullable[domain.BirthDate]{},
			nullable.Nullable[domain.Height]{},
		)

		// Act
		_, err := uc.Execute(ctx, patch)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}
