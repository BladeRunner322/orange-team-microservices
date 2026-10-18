package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
)

type GetUserScore struct {
	workoutRepo ports.WorkoutRepository
	logger      *logger.Logger
}

func NewGetUserScore(repo ports.WorkoutRepository, log *logger.Logger) *GetUserScore {
	return &GetUserScore{workoutRepo: repo, logger: log}
}

func (uc *GetUserScore) Execute(ctx context.Context) (int64, error) {
	// TODO: userID from ctx, GetUserScore
	return 0, nil
}
