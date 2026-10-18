package usecases

import (
	"context"

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

func (uc *CreateWorkout) Execute(ctx context.Context) (domain.Workout, error) {
	// TODO:
	//  1. userID := userIDFromContext(ctx)
	//  2. profile := profilesClient.GetProfile(ctx, userID)
	//  3. coeff := domain.CalculatePersonalScoreCoefficient(...)
	//  4. workout := domain.NewWorkout(userID, coeff)
	//  5. return workoutRepo.Create(ctx, workout)
	return domain.Workout{}, nil
}
