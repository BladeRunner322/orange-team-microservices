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

type GetWorkoutExercises struct {
	workoutRepo  ports.WorkoutRepository
	exerciseRepo ports.WorkoutExerciseRepository
	logger       *logger.Logger
}

func NewGetWorkoutExercises(
	workoutRepo ports.WorkoutRepository,
	exerciseRepo ports.WorkoutExerciseRepository,
	log *logger.Logger,
) *GetWorkoutExercises {
	return &GetWorkoutExercises{
		workoutRepo:  workoutRepo,
		exerciseRepo: exerciseRepo,
		logger:       log,
	}
}

func (uc *GetWorkoutExercises) Execute(ctx context.Context, workoutID uuid.UUID) ([]domain.WorkoutExercise, error) {
	log := uc.logger.With("operation", "GetWorkoutExercises", "workout_id", workoutID)

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return nil, err
	}

	workout, err := uc.workoutRepo.GetByID(ctx, userID, workoutID)
	if err != nil {
		if errors.Is(err, domain.ErrWorkoutNotFound) {
			log.Warn("workout not found")
			return nil, domain.ErrWorkoutNotFound
		}
		log.Error("failed to get workout", "error", err)
		return nil, fmt.Errorf("get workout: %w", err)
	}

	if workout.IsDeleted() {
		log.Warn("workout is deleted")
		return nil, domain.ErrWorkoutNotFound
	}

	exercises, err := uc.exerciseRepo.ListByWorkout(ctx, workoutID)
	if err != nil {
		log.Error("failed to list workout exercises", "error", err)
		return nil, fmt.Errorf("list workout exercises: %w", err)
	}

	return exercises, nil
}
