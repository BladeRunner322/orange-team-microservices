package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
	"github.com/google/uuid"
)

type DeleteWorkout struct {
	workoutRepo ports.WorkoutRepository
	logger      *logger.Logger
}

func NewDeleteWorkout(repo ports.WorkoutRepository, log *logger.Logger) *DeleteWorkout {
	return &DeleteWorkout{workoutRepo: repo, logger: log}
}

// Execute помечает тренировку удалённой (soft delete).
//
// Идемпотентен: повторный вызов или DELETE несуществующего id
// возвращает nil, а не ошибку (ADR-007 F-3).
// Принадлежность проверяется через user_id в MarkDeleted.
func (uc *DeleteWorkout) Execute(ctx context.Context, id uuid.UUID) error {
	log := uc.logger.With("operation", "DeleteWorkout", "workout_id", id)

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return err
	}

	if err := uc.workoutRepo.MarkDeleted(ctx, userID, id); err != nil {
		if errors.Is(err, domain.ErrWorkoutNotFound) {
			log.Warn("workout not found or already deleted")
			return nil
		}

		log.Error("failed to mark workout deleted", "error", err)
		return fmt.Errorf("mark workout deleted: %w", err)
	}

	return nil
}
