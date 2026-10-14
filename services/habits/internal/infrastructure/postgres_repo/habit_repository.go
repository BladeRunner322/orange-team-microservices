package postgres_repo

import (
	"context"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	"github.com/google/uuid"
)

// HabitRepository реализует ports.HabitRepository для PostgreSQL.
type HabitRepository struct {
	pool      postgres.Pool
	tx        postgres.Tx
	opTimeout time.Duration
}

// NewHabitRepository создаёт новый репозиторий с пулом.
func NewHabitRepository(pool postgres.Pool) *HabitRepository {
	// TODO
	return nil
}

// querier возвращает актуальный источник запросов: транзакцию или пул.
func (r *HabitRepository) querier() postgres.Querier {
	// TODO
	return nil
}

// WithTx выполняет fn внутри транзакции (по образцу Exercises).
func (r *HabitRepository) WithTx(ctx context.Context, fn func(ports.HabitRepository) error) error {
	// TODO
	return nil
}

// Create сохраняет новую привычку, возвращает из RETURNING.
func (r *HabitRepository) Create(ctx context.Context, habit domain.Habit) (domain.Habit, error) {
	// TODO
	return domain.Habit{}, nil
}

// GetByID возвращает привычку по id (включая удалённые).
func (r *HabitRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Habit, error) {
	// TODO
	return domain.Habit{}, nil
}

// GetByIDForUpdate возвращает привычку с блокировкой (SELECT ... FOR UPDATE).
// Вызывается только внутри WithTx.
func (r *HabitRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (domain.Habit, error) {
	// TODO
	return domain.Habit{}, nil
}

// ListByUser возвращает активные привычки пользователя.
func (r *HabitRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Habit, error) {
	// TODO: make([]domain.Habit, 0), сортировка по created_at ASC
	return nil, nil
}

// Update обновляет name/description/updated_at.
// Не трогает deleted_at. Возвращает из RETURNING.
func (r *HabitRepository) Update(ctx context.Context, habit domain.Habit) (domain.Habit, error) {
	// TODO
	return domain.Habit{}, nil
}

// MarkDeleted ставит deleted_at = NOW(), updated_at = NOW().
// WHERE id = $1 AND deleted_at IS NULL — идемпотентность через RowsAffected.
func (r *HabitRepository) MarkDeleted(ctx context.Context, id uuid.UUID) error {
	// TODO
	return nil
}