package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
	"github.com/google/uuid"
)

type GetWorkout struct {
	workoutRepo ports.WorkoutRepository
	logger      *logger.Logger
}

func NewGetWorkout(repo ports.WorkoutRepository, log *logger.Logger) *GetWorkout {
	return &GetWorkout{workoutRepo: repo, logger: log}
}

func (uc *GetWorkout) Execute(ctx context.Context, id uuid.UUID) (domain.Workout, error) {
	// TODO: userID from ctx, get from repo
	return domain.Workout{}, nil
}
