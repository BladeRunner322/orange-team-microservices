package postgres_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/pgerrors"
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

// Create вставляет упражнение в БД и возвращает его из RETURNING.
func (r *WorkoutExerciseRepository) Create(ctx context.Context, exercise domain.WorkoutExercise) (domain.WorkoutExercise, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		INSERT INTO workouts.workout_exercises (
			id, version, workout_id, exercise_id, difficulty, weight, sets, reps, duration, completed, exercise_load, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, version, workout_id, exercise_id, difficulty, weight, sets, reps, duration, completed, exercise_load, created_at, updated_at
	`
	m := WorkoutExerciseDomainToModel(exercise)

	row := r.querier().QueryRow(ctx, query,
		m.ID,
		m.Version,
		m.WorkoutID,
		m.ExerciseID,
		m.Difficulty,
		m.Weight,
		m.Sets,
		m.Reps,
		m.Duration,
		m.Completed,
		m.ExerciseLoad,
		m.CreatedAt,
	)

	if err := m.Scan(row); err != nil {
		return domain.WorkoutExercise{}, fmt.Errorf("create workout exercise: %w", err)
	}

	created, err := WorkoutExerciseModelToDomain(m)
	if err != nil {
		return domain.WorkoutExercise{}, fmt.Errorf("workout exercise %s: %w", m.ID, err)
	}

	return created, nil
}

// GetByID возвращает упражнение по id в рамках тренировки.
// Не найдено → ErrWorkoutExerciseNotFound.
func (r *WorkoutExerciseRepository) GetByID(ctx context.Context, workoutID uuid.UUID, id uuid.UUID) (domain.WorkoutExercise, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		SELECT id, version, workout_id, exercise_id, difficulty, weight, sets, reps, duration, completed, exercise_load, created_at, updated_at
		FROM workouts.workout_exercises
		WHERE workout_id = $1 AND id = $2
	`

	var m WorkoutExerciseModel
	row := r.querier().QueryRow(ctx, query, workoutID, id)
	if err := m.Scan(row); err != nil {
		if errors.Is(err, pgerrors.ErrNoRows) {
			return domain.WorkoutExercise{}, domain.ErrWorkoutExerciseNotFound
		}

		return domain.WorkoutExercise{}, fmt.Errorf("find workout exercise by id: %w", err)
	}

	exercise, err := WorkoutExerciseModelToDomain(m)
	if err != nil {
		return domain.WorkoutExercise{}, fmt.Errorf("workout exercise %s: %w", m.ID, err)
	}

	return exercise, nil
}

// ListByWorkout возвращает все упражнения тренировки.
// Порядок — created_at ASC, id ASC (детерминированный).
// Пустой список → пустой слайс (не nil).
func (r *WorkoutExerciseRepository) ListByWorkout(ctx context.Context, workoutID uuid.UUID) ([]domain.WorkoutExercise, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		SELECT id, version, workout_id, exercise_id, difficulty, weight, sets, reps, duration, completed, exercise_load, created_at, updated_at
		FROM workouts.workout_exercises
		WHERE workout_id = $1
		ORDER BY created_at ASC, id ASC
	`

	rows, err := r.querier().Query(ctx, query, workoutID)
	if err != nil {
		return nil, fmt.Errorf("query workout exercises: %w", err)
	}
	defer rows.Close()

	exercises := make([]domain.WorkoutExercise, 0)
	for rows.Next() {
		var m WorkoutExerciseModel
		if err := m.Scan(rows); err != nil {
			return nil, fmt.Errorf("scan workout exercise: %w", err)
		}

		exercise, err := WorkoutExerciseModelToDomain(m)
		if err != nil {
			return nil, fmt.Errorf("workout exercise %s: %w", m.ID, err)
		}

		exercises = append(exercises, exercise)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workout exercises: %w", err)
	}

	return exercises, nil
}

// Update обновляет упражнение и возвращает его из RETURNING.
//
// Optimistic locking: WHERE ... AND version = $N. Если строка не найдена —
// ErrConflict (usecase вызывает Update только после успешного GetByID,
// поэтому «не найдено» = конкурентное изменение или удаление).
//
// Не трогает: workout_id (принадлежность), exercise_id (immutable, ADR-025 п.9),
// difficulty (снапшот, ADR-008), created_at.
func (r *WorkoutExerciseRepository) Update(ctx context.Context, exercise domain.WorkoutExercise) (domain.WorkoutExercise, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		UPDATE workouts.workout_exercises
		SET
			version = version + 1,
			weight = $4,
			sets = $5,
			reps = $6,
			duration = $7,
			completed = $8,
			exercise_load = $9,
			updated_at = NOW()
		WHERE id = $1 AND workout_id = $2 AND version = $3
		RETURNING id, version, workout_id, exercise_id, difficulty, weight, sets, reps, duration, completed, exercise_load, created_at, updated_at
	`

	m := WorkoutExerciseDomainToModel(exercise)

	row := r.querier().QueryRow(ctx, query,
		m.ID,
		m.WorkoutID,
		m.Version,
		m.Weight,
		m.Sets,
		m.Reps,
		m.Duration,
		m.Completed,
		m.ExerciseLoad,
	)

	if err := m.Scan(row); err != nil {
		if errors.Is(err, pgerrors.ErrNoRows) {
			return domain.WorkoutExercise{}, domain.ErrConflict
		}

		return domain.WorkoutExercise{}, fmt.Errorf("update workout exercise: %w", err)
	}

	updated, err := WorkoutExerciseModelToDomain(m)
	if err != nil {
		return domain.WorkoutExercise{}, fmt.Errorf("workout exercise %s: %w", m.ID, err)
	}

	return updated, nil
}

// Delete удаляет упражнение (hard delete, ADR-025 п.3.2).
// Не найдено → ErrWorkoutExerciseNotFound.
// Идемпотентность DELETE обеспечивается в usecase (ADR-007 F-3):
// он перехватывает ErrWorkoutExerciseNotFound и возвращает nil.
func (r *WorkoutExerciseRepository) Delete(ctx context.Context, workoutID uuid.UUID, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		DELETE FROM workouts.workout_exercises
		WHERE workout_id = $1 AND id = $2
	`

	cmdTag, err := r.querier().Exec(ctx, query, workoutID, id)
	if err != nil {
		return fmt.Errorf("delete workout exercise: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrWorkoutExerciseNotFound
	}

	return nil
}
