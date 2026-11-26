// Package postgres_repo — реализация ports.HabitRepository и ports.CompletionRepository для PostgreSQL.
package postgres_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/pgerrors"
	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HabitRepository реализует ports.HabitRepository для PostgreSQL.
type HabitRepository struct {
	db        *pgxpool.Pool
	getter    *trmpgx.CtxGetter
	opTimeout time.Duration
}

// NewHabitRepository создаёт репозиторий привычек с доступом к транзакции из контекста.
func NewHabitRepository(db *pgxpool.Pool, getter *trmpgx.CtxGetter, opTimeout time.Duration) *HabitRepository {
	return &HabitRepository{
		db:        db,
		getter:    getter,
		opTimeout: opTimeout,
	}
}

// queryExecutor выбирает транзакцию или пул и подключает обработку ошибок pgx.
func (r *HabitRepository) queryExecutor(ctx context.Context) postgres.QueryExecutor {
	return postgres.WithMappedErrors(r.getter.DefaultTrOrDB(ctx, r.db))
}

// Create сохраняет новую привычку, возвращает из RETURNING.
func (r *HabitRepository) Create(ctx context.Context, habit domain.Habit) (domain.Habit, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		INSERT INTO habits.habits (id, user_id, name, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, name, description, deleted_at, created_at, updated_at
    `

	m := HabitDomainToModel(habit)

	row := r.queryExecutor(ctx).QueryRow(ctx, query, m.ID, m.UserID, m.Name, m.Description)

	if err := m.Scan(row); err != nil {
		if errors.Is(err, pgerrors.ErrViolatesUnique) {
			return domain.Habit{}, domain.ErrHabitNameExists
		}

		return domain.Habit{}, fmt.Errorf("create habit: %w", err)
	}

	created, err := HabitModelToDomain(m)
	if err != nil {
		return domain.Habit{}, fmt.Errorf("habit %s: %w", m.ID, err)
	}

	return created, nil
}

// GetByID возвращает привычку по id (включая удалённые).
func (r *HabitRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Habit, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		SELECT id, user_id, name, description, deleted_at, created_at, updated_at
		FROM habits.habits
		WHERE id = $1
	`

	var m HabitModel
	row := r.queryExecutor(ctx).QueryRow(ctx, query, id)
	if err := m.Scan(row); err != nil {
		if errors.Is(err, pgerrors.ErrNoRows) {
			return domain.Habit{}, domain.ErrHabitNotFound
		}

		return domain.Habit{}, fmt.Errorf("find habit by id: %w", err)
	}

	habit, err := HabitModelToDomain(m)
	if err != nil {
		return domain.Habit{}, fmt.Errorf("habit %s: %w", m.ID, err)
	}

	return habit, nil
}

// GetByIDForUpdate читает привычку с блокировкой строки.
// Вызывается с контекстом транзакции, чтобы блокировка держалась до её завершения.
func (r *HabitRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (domain.Habit, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		SELECT id, user_id, name, description, deleted_at, created_at, updated_at
		FROM habits.habits
		WHERE id = $1
		FOR UPDATE
	`

	var m HabitModel
	row := r.queryExecutor(ctx).QueryRow(ctx, query, id)
	if err := m.Scan(row); err != nil {
		if errors.Is(err, pgerrors.ErrNoRows) {
			return domain.Habit{}, domain.ErrHabitNotFound
		}

		return domain.Habit{}, fmt.Errorf("find habit by id for update: %w", err)
	}

	habit, err := HabitModelToDomain(m)
	if err != nil {
		return domain.Habit{}, fmt.Errorf("habit %s: %w", m.ID, err)
	}

	return habit, nil
}

// ListByUser возвращает активные привычки пользователя.
func (r *HabitRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Habit, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		SELECT id, user_id, name, description, deleted_at, created_at, updated_at
		FROM habits.habits
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`

	rows, err := r.queryExecutor(ctx).Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query habits: %w", err)
	}
	defer rows.Close()

	habits := make([]domain.Habit, 0)
	for rows.Next() {
		var m HabitModel
		if err := m.Scan(rows); err != nil {
			return nil, fmt.Errorf("scan habit: %w", err)
		}

		habit, err := HabitModelToDomain(m)
		if err != nil {
			return nil, fmt.Errorf("habit %s: %w", m.ID, err)
		}

		habits = append(habits, habit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate habits: %w", err)
	}

	return habits, nil
}

// Update обновляет name/description/updated_at.
// Не трогает deleted_at. Возвращает из RETURNING.
func (r *HabitRepository) Update(ctx context.Context, habit domain.Habit) (domain.Habit, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		UPDATE habits.habits
		SET
			name = $2,
			description = $3,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, user_id, name, description, deleted_at, created_at, updated_at
	`

	m := HabitDomainToModel(habit)

	row := r.queryExecutor(ctx).QueryRow(ctx, query, m.ID, m.Name, m.Description)

	if err := m.Scan(row); err != nil {
		if errors.Is(err, pgerrors.ErrNoRows) {
			return domain.Habit{}, domain.ErrHabitNotFound
		}

		if errors.Is(err, pgerrors.ErrViolatesUnique) {
			return domain.Habit{}, domain.ErrHabitNameExists
		}

		return domain.Habit{}, fmt.Errorf("update habit: %w", err)
	}

	updated, err := HabitModelToDomain(m)
	if err != nil {
		return domain.Habit{}, fmt.Errorf("habit %s: %w", m.ID, err)
	}

	return updated, nil
}

// MarkDeleted ставит deleted_at = NOW(), updated_at = NOW().
// WHERE id = $1 AND deleted_at IS NULL — идемпотентность через RowsAffected.
func (r *HabitRepository) MarkDeleted(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		UPDATE habits.habits
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	cmdTag, err := r.queryExecutor(ctx).Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark deleted: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrHabitNotFound
	}

	return nil
}
