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

type GetWorkout struct {
	workoutRepo ports.WorkoutRepository
	logger      *logger.Logger
}

func NewGetWorkout(repo ports.WorkoutRepository, log *logger.Logger) *GetWorkout {
	return &GetWorkout{workoutRepo: repo, logger: log}
}

// Execute возвращает тренировку по id.
// Проверяет, что тренировка принадлежит текущему пользователю:
// чужая тренировка → ErrWorkoutNotFound (существование не раскрываем).
func (uc *GetWorkout) Execute(ctx context.Context, id uuid.UUID) (domain.Workout, error) {
	log := uc.logger.With("operation", "GetWorkout", "workout_id", id)

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return domain.Workout{}, err
	}

	workout, err := uc.workoutRepo.GetByID(ctx, userID, id)
	if err != nil {
		if errors.Is(err, domain.ErrWorkoutNotFound) {
			log.Warn("workout not found")
			return domain.Workout{}, domain.ErrWorkoutNotFound
		}
		
		log.Error("failed to get workout", "error", err)
		return domain.Workout{}, fmt.Errorf("get workout: %w", err)
	}

	return workout, nil
}
