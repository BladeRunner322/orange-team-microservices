package postgres_repo

import (
	"context"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
	"github.com/google/uuid"
)

type WorkoutRepository struct {
	pool      postgres.Pool
	tx        postgres.Tx
	opTimeout time.Duration
}

func NewWorkoutRepository(pool postgres.Pool) *WorkoutRepository {
	return &WorkoutRepository{
		pool:      pool,
		opTimeout: pool.OpTimeout(),
	}
}

func (r *WorkoutRepository) querier() postgres.Querier {
	if r.tx != nil {
		return r.tx
	}
	return r.pool
}

// WithTx открывает транзакцию и передаёт оба репозитория (ADR-025 п.3.1).
func (r *WorkoutRepository) WithTx(
	ctx context.Context,
	fn func(ports.WorkoutRepository, ports.WorkoutExerciseRepository) error,
) error {
	// TODO: r.pool.WithTx → создать txWorkoutRepo + txExerciseRepo, вызвать fn
	return nil
}

func (r *WorkoutRepository) Create(ctx context.Context, workout domain.Workout) (domain.Workout, error) {
	// TODO: INSERT INTO workouts.workouts ... RETURNING ...
	return domain.Workout{}, nil
}

func (r *WorkoutRepository) GetByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (domain.Workout, error) {
	// TODO: SELECT ... WHERE user_id = $1 AND id = $2
	return domain.Workout{}, nil
}

func (r *WorkoutRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Workout, error) {
	// TODO: SELECT ... WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC
	return []domain.Workout{}, nil
}

func (r *WorkoutRepository) Update(ctx context.Context, workout domain.Workout) (domain.Workout, error) {
	// TODO: UPDATE ... SET ..., version = version + 1
	//       WHERE id = $1 AND version = $2 AND deleted_at IS NULL
	//       RETURNING ...
	return domain.Workout{}, nil
}

func (r *WorkoutRepository) MarkDeleted(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	// TODO: UPDATE workouts.workouts
	//       SET deleted_at = NOW(), updated_at = NOW()
	//       WHERE user_id = $1 AND id = $2 AND deleted_at IS NULL
	//       если RowsAffected == 0 → ErrWorkoutNotFound
	return nil
}

func (r *WorkoutRepository) GetUserScore(ctx context.Context, userID uuid.UUID) (int64, error) {
	// TODO: SELECT COALESCE(SUM(workout_score), 0) FROM workouts.workouts
	//       WHERE user_id = $1 AND status = 'completed' AND deleted_at IS NULL
	return 0, nil
}
