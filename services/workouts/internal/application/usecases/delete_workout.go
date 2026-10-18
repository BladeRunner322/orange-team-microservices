package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/google/uuid"
)

type DeleteWorkout struct {
	workoutRepo ports.WorkoutRepository
	logger      *logger.Logger
}

func NewDeleteWorkout(repo ports.WorkoutRepository, log *logger.Logger) *DeleteWorkout {
	return &DeleteWorkout{workoutRepo: repo, logger: log}
}

func (uc *DeleteWorkout) Execute(ctx context.Context, id uuid.UUID) error {
	// TODO: userID from ctx
	//       err := workoutRepo.MarkDeleted(ctx, userID, id)
	//       if errors.Is(err, domain.ErrWorkoutNotFound) { return nil }  // идемпотентность, ADR-007 F-3
	//       return err
	return nil
}
