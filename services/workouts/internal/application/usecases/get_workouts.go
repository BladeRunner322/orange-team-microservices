package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
)

type GetWorkouts struct {
	workoutRepo ports.WorkoutRepository
	logger      *logger.Logger
}

func NewGetWorkouts(repo ports.WorkoutRepository, log *logger.Logger) *GetWorkouts {
	return &GetWorkouts{workoutRepo: repo, logger: log}
}

func (uc *GetWorkouts) Execute(ctx context.Context) ([]domain.Workout, error) {
	// TODO: userID from ctx, list from repo
	return []domain.Workout{}, nil
}
