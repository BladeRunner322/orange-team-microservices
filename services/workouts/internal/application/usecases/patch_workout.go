package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
	"github.com/google/uuid"
)

type PatchWorkout struct {
	workoutRepo ports.WorkoutRepository
	logger      *logger.Logger
}

func NewPatchWorkout(repo ports.WorkoutRepository, log *logger.Logger) *PatchWorkout {
	return &PatchWorkout{workoutRepo: repo, logger: log}
}

func (uc *PatchWorkout) Execute(ctx context.Context, id uuid.UUID, patch domain.WorkoutPatch) (domain.Workout, error) {
	// TODO: userID, GetByID, ApplyPatch, Update
	return domain.Workout{}, nil
}
