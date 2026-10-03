// Package postgres_repo — реализация ports.Repository для PostgreSQL.
package postgres_repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/domain"
	"github.com/google/uuid"
)

// Repository реализует интерфейс ports.Repository для PostgreSQL.
type Repository struct {
	pool postgres.Pool
}

// NewRepository создаёт новый репозиторий с пулом соединений.
func NewRepository(pool postgres.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create вставляет упражнение в БД и возвращает его из RETURNING.
func (r *Repository) Create(ctx context.Context, exercise domain.Exercise) (domain.Exercise, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO exercises.exercises (id, name, description, difficulty, type)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, description, difficulty, type, deleted_at, created_at, updated_at
    `
	m := DomainToModel(exercise)

	row := r.pool.QueryRow(ctx, query, m.ID, m.Name, m.Description, m.Difficulty, m.Type)

	if err := m.Scan(row); err != nil {
		if errors.Is(err, postgres.ErrViolatesUnique) {
			return domain.Exercise{}, domain.ErrExerciseNameExists
		}

		return domain.Exercise{}, fmt.Errorf("create exercise: %w", err)
	}

	created, err := ModelToDomain(m)
	if err != nil {
		return domain.Exercise{}, fmt.Errorf("exercise %s: %w", m.ID, err)
	}

	return created, nil
}

// GetExercise возвращает упражнение по id (включая удалённые).
func (r *Repository) GetExercise(ctx context.Context, id uuid.UUID) (domain.Exercise, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
        SELECT id, name, description, difficulty, type, deleted_at, created_at, updated_at
        FROM exercises.exercises
        WHERE id = $1
    `

	var m ExerciseModel
	row := r.pool.QueryRow(ctx, query, id)
	if err := m.Scan(row); err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return domain.Exercise{}, domain.ErrExerciseNotFound
		}

		return domain.Exercise{}, fmt.Errorf("find exercise by id: %w", err)
	}

	exercise, err := ModelToDomain(m)
	if err != nil {
		return domain.Exercise{}, fmt.Errorf("exercise %s: %w", m.ID, err)
	}

	return exercise, nil
}

// GetExercises возвращает список активных упражнений (deleted_at IS NULL).
func (r *Repository) GetExercises(ctx context.Context) ([]domain.Exercise, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
        SELECT id, name, description, difficulty, type, deleted_at, created_at, updated_at
        FROM exercises.exercises
		WHERE deleted_at IS NULL
		ORDER BY id ASC
    `

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query exercises: %w", err)
	}
	defer rows.Close()

	exercises := make([]domain.Exercise, 0)
	for rows.Next() {
		var m ExerciseModel
		if err := m.Scan(rows); err != nil {
			return nil, fmt.Errorf("scan exercise: %w", err)
		}

		exercise, err := ModelToDomain(m)
		if err != nil {
			return nil, fmt.Errorf("exercise %s: %w", m.ID, err)
		}

		exercises = append(exercises, exercise)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exercises: %w", err)
	}

	return exercises, nil
}

// Update обновляет name, description, difficulty, updated_at.
// deleted_at не трогает.
func (r *Repository) Update(ctx context.Context, exercise domain.Exercise) (domain.Exercise, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE exercises.exercises
		SET
			name = $2,
			description = $3,
			difficulty = $4,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, name, description, difficulty, type, deleted_at, created_at, updated_at
	`

	m := DomainToModel(exercise)

	row := r.pool.QueryRow(ctx, query, m.ID, m.Name, m.Description, m.Difficulty)

	if err := m.Scan(row); err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return domain.Exercise{}, domain.ErrExerciseNotFound
		}

		if errors.Is(err, postgres.ErrViolatesUnique) {
			return domain.Exercise{}, domain.ErrExerciseNameExists
		}

		return domain.Exercise{}, fmt.Errorf("update exercise: %w", err)
	}

	updated, err := ModelToDomain(m)
	if err != nil {
		return domain.Exercise{}, fmt.Errorf("exercise %s: %w", m.ID, err)
	}

	return updated, nil
}

// MarkDeleted ставит deleted_at = NOW() и updated_at = NOW().
// Если уже удалено — ErrExerciseNotFound.
func (r *Repository) MarkDeleted(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE exercises.exercises
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark deleted: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrExerciseNotFound
	}

	return nil
}
