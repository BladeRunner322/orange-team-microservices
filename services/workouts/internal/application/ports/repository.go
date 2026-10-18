// Package ports — интерфейсы, которые usecases требуют от инфраструктуры.
package ports

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
	"github.com/google/uuid"
)

// WorkoutRepository — работа с таблицей workouts.workouts.
type WorkoutRepository interface {
	// WithTx открывает транзакцию и передаёт в callback оба репозитория,
	// привязанных к одной транзакции. См. ADR-025 п.3.1.
	WithTx(ctx context.Context, fn func(WorkoutRepository, WorkoutExerciseRepository) error) error

	Create(ctx context.Context, workout domain.Workout) (domain.Workout, error)
	GetByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (domain.Workout, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Workout, error)
	Update(ctx context.Context, workout domain.Workout) (domain.Workout, error)
	Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
	GetUserScore(ctx context.Context, userID uuid.UUID) (int64, error)
}

// WorkoutExerciseRepository — работа с таблицей workouts.workout_exercises.
type WorkoutExerciseRepository interface {
	Create(ctx context.Context, exercise domain.WorkoutExercise) (domain.WorkoutExercise, error)
	GetByID(ctx context.Context, workoutID uuid.UUID, id uuid.UUID) (domain.WorkoutExercise, error)
	ListByWorkout(ctx context.Context, workoutID uuid.UUID) ([]domain.WorkoutExercise, error)
	Update(ctx context.Context, exercise domain.WorkoutExercise) (domain.WorkoutExercise, error)
	Delete(ctx context.Context, workoutID uuid.UUID, id uuid.UUID) error
}
