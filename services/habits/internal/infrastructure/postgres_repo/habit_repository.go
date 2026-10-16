// Package postgres_repo — реализация ports.HabitRepository и ports.CompletionRepository для PostgreSQL.
package postgres_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/pgerrors"
	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	"github.com/google/uuid"
)

// HabitRepository реализует ports.HabitRepository для PostgreSQL.
type HabitRepository struct {
	pool      postgres.Pool // nil, если репозиторий привязан к транзакции
	tx        postgres.Tx   // nil, если репозиторий не в транзакции
	opTimeout time.Duration
}

// NewHabitRepository создаёт новый репозиторий с пулом.
func NewHabitRepository(pool postgres.Pool) *HabitRepository {
	return &HabitRepository{
		pool:      pool,
		opTimeout: pool.OpTimeout(),
	}
}

// querier возвращает актуальный источник запросов: транзакцию или пул.
func (r *HabitRepository) querier() postgres.Querier {
	if r.tx != nil {
		return r.tx
	}
	return r.pool
}

// WithTx выполняет fn внутри транзакции. Если репозиторий уже привязан
// к транзакции — вложенная не открывается, fn вызывается с текущим.
func (r *HabitRepository) WithTx(ctx context.Context, fn func(ports.HabitRepository) error) error {
	if r.tx != nil {
		return fn(r)
	}
	return r.pool.WithTx(ctx, func(tx postgres.Tx) error {
		txRepo := &HabitRepository{tx: tx, opTimeout: r.opTimeout}
		return fn(txRepo)
	})
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

	row := r.querier().QueryRow(ctx, query, m.ID, m.UserID, m.Name, m.Description)

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
	row := r.querier().QueryRow(ctx, query, id)
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

// GetByIDForUpdate возвращает привычку с блокировкой строки (SELECT ... FOR UPDATE).
// Блокировка держится до COMMIT/ROLLBACK транзакции. Если репозиторий не в tx —
// поведение как у обычного SELECT, блокировка снимается сразу.
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
	row := r.querier().QueryRow(ctx, query, id)
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

	rows, err := r.querier().Query(ctx, query, userID)
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

	row := r.querier().QueryRow(ctx, query, m.ID, m.Name, m.Description)

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

	cmdTag, err := r.querier().Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark deleted: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrHabitNotFound
	}

	return nil
}
