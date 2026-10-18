package postgres_repo

import (
	"context"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
	"github.com/google/uuid"
)

type WorkoutExerciseRepository struct {
	pool      postgres.Pool
	tx        postgres.Tx
	opTimeout time.Duration
}

func NewWorkoutExerciseRepository(pool postgres.Pool) *WorkoutExerciseRepository {
	return &WorkoutExerciseRepository{
		pool:      pool,
		opTimeout: pool.OpTimeout(),
	}
}

func (r *WorkoutExerciseRepository) querier() postgres.Querier {
	if r.tx != nil {
		return r.tx
	}
	return r.pool
}

func (r *WorkoutExerciseRepository) Create(ctx context.Context, exercise domain.WorkoutExercise) (domain.WorkoutExercise, error) {
	// TODO: INSERT INTO workouts.workout_exercises ... RETURNING ...
	return domain.WorkoutExercise{}, nil
}

func (r *WorkoutExerciseRepository) GetByID(ctx context.Context, workoutID uuid.UUID, id uuid.UUID) (domain.WorkoutExercise, error) {
	// TODO: SELECT ... WHERE workout_id = $1 AND id = $2
	return domain.WorkoutExercise{}, nil
}

func (r *WorkoutExerciseRepository) ListByWorkout(ctx context.Context, workoutID uuid.UUID) ([]domain.WorkoutExercise, error) {
	// TODO: SELECT ... WHERE workout_id = $1 ORDER BY created_at ASC
	return []domain.WorkoutExercise{}, nil
}

func (r *WorkoutExerciseRepository) Update(ctx context.Context, exercise domain.WorkoutExercise) (domain.WorkoutExercise, error) {
	// TODO: UPDATE ... SET ..., version = version + 1 WHERE id = $1 AND version = $2 RETURNING ...
	return domain.WorkoutExercise{}, nil
}

func (r *WorkoutExerciseRepository) Delete(ctx context.Context, workoutID uuid.UUID, id uuid.UUID) error {
	// TODO: DELETE FROM workouts.workout_exercises WHERE workout_id = $1 AND id = $2
	return nil
}
