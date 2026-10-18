package usecases

import (
	"context"

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
	// TODO: userID, check ownership, ListByWorkout
	return []domain.WorkoutExercise{}, nil
}
