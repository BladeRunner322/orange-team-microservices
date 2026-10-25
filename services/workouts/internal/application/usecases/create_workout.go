package usecases

import (
	"context"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
)

type CreateWorkout struct {
	workoutRepo    ports.WorkoutRepository
	profilesClient ports.ProfilesClient
	logger         *logger.Logger
}

func NewCreateWorkout(
	workoutRepo ports.WorkoutRepository,
	profilesClient ports.ProfilesClient,
	log *logger.Logger,
) *CreateWorkout {
	return &CreateWorkout{
		workoutRepo:    workoutRepo,
		profilesClient: profilesClient,
		logger:         log,
	}
}

// Execute создаёт новую тренировку.
// Коэффициент считается из профиля пользователя (ADR-025 п.2)
// и фиксируется как снапшот на момент создания (ADR-001).
func (uc *CreateWorkout) Execute(ctx context.Context) (domain.Workout, error) {
	log := uc.logger.With("operation", "CreateWorkout")

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return domain.Workout{}, err
	}

	profile, err := uc.profilesClient.GetProfile(ctx, userID)
	if err != nil {
		log.Error("failed to get user profile", "error", err)
		return domain.Workout{}, fmt.Errorf("get user profile: %w", err)
	}

	coeff := domain.CalculatePersonalScoreCoefficient(domain.UserProfileData{
		Sex:         profile.Sex,
		WeightGrams: profile.WeightGrams,
		HeightCM:    profile.HeightCM,
	})

	workout := domain.NewWorkout(userID, coeff)

	created, err := uc.workoutRepo.Create(ctx, workout)
	if err != nil {
		log.Error("failed to create workout", "error", err)
		return domain.Workout{}, fmt.Errorf("create workout: %w", err)
	}

	return created, nil
}
