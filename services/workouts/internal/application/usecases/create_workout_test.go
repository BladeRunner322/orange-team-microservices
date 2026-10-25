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

func TestCreateWorkout_Execute(t *testing.T) {
	newCtx := func(userID uuid.UUID) context.Context {
		return ctxkeys.WithUserID(context.Background(), userID.String())
	}

	strPtr := func(v string) *string { return &v }
	intPtr := func(v int) *int { return &v }

	t.Run("успешно создаёт тренировку с полным профилем", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		profilesClient := &mockProfilesClient{
			profile: ports.UserProfile{
				UserID:      userID,
				Sex:         strPtr("male"),
				WeightGrams: intPtr(80000),
				HeightCM:    intPtr(180),
			},
		}

		uc := NewCreateWorkout(repo, profilesClient, logger.NewTestLogger())
		created, err := uc.Execute(ctx)

		require.NoError(t, err)

		expectedCoeff := domain.CalculatePersonalScoreCoefficient(domain.UserProfileData{
			Sex:         strPtr("male"),
			WeightGrams: intPtr(80000),
			HeightCM:    intPtr(180),
		})

		assert.Equal(t, userID, created.UserID())
		assert.Equal(t, domain.StatusPlanned, created.Status())
		assert.Equal(t, expectedCoeff, created.PersonalScoreCoefficient())

		// сохранено в репозитории
		found, err := repo.GetByID(ctx, userID, created.ID())
		require.NoError(t, err)
		assert.Equal(t, created.ID(), found.ID())
	})

	t.Run("профиль не заполнен — coefficient = 1", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockWorkoutRepository()
		profilesClient := &mockProfilesClient{
			profile: ports.UserProfile{UserID: userID},
		}

		uc := NewCreateWorkout(repo, profilesClient, logger.NewTestLogger())
		created, err := uc.Execute(ctx)

		require.NoError(t, err)
		assert.Equal(t, 1, created.PersonalScoreCoefficient())
	})

	t.Run("нет user_id в context — ErrUnauthenticated", func(t *testing.T) {
		repo := newMockWorkoutRepository()
		profilesClient := &mockProfilesClient{}
		uc := NewCreateWorkout(repo, profilesClient, logger.NewTestLogger())

		_, err := uc.Execute(context.Background())
		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})

	t.Run("ошибка Profiles — пробрасывается", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockWorkoutRepository()
		profilesClient := &mockProfilesClient{err: errors.New("profiles is down")}
		uc := NewCreateWorkout(repo, profilesClient, logger.NewTestLogger())

		_, err := uc.Execute(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "profiles is down")
	})

	t.Run("ошибка репозитория при Create — пробрасывается", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockWorkoutRepository()
		repo.err = errors.New("db is down")
		profilesClient := &mockProfilesClient{profile: ports.UserProfile{}}
		uc := NewCreateWorkout(repo, profilesClient, logger.NewTestLogger())

		_, err := uc.Execute(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}
