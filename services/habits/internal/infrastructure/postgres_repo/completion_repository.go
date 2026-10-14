package postgres_repo

import (
	"context"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	"github.com/google/uuid"
)

// CompletionRepository реализует ports.CompletionRepository для PostgreSQL.
type CompletionRepository struct {
	pool      postgres.Pool
	opTimeout time.Duration
}

// NewCompletionRepository создаёт новый репозиторий с пулом.
func NewCompletionRepository(pool postgres.Pool) *CompletionRepository {
	// TODO
	return nil
}

// Complete добавляет отметку. ON CONFLICT DO NOTHING — повторная отметка не падает.
func (r *CompletionRepository) Complete(ctx context.Context, completion domain.Completion) error {
	// TODO
	return nil
}

// Uncomplete удаляет отметку за дату. Идемпотентно.
func (r *CompletionRepository) Uncomplete(ctx context.Context, habitID uuid.UUID, date time.Time) error {
	// TODO
	return nil
}

// IsCompleted — проверяет наличие отметки на дату.
func (r *CompletionRepository) IsCompleted(ctx context.Context, habitID uuid.UUID, date time.Time) (bool, error) {
	// TODO: SELECT EXISTS (...)
	return false, nil
}

// ListByHabit возвращает отметки за период [from, to] включительно.
// Сортировка по date ASC.
func (r *CompletionRepository) ListByHabit(ctx context.Context, habitID uuid.UUID, from, to time.Time) ([]domain.Completion, error) {
	// TODO
	return nil, nil
}

// Streak считает текущий стрик привычки.
//
// Логика: идём от today назад, считаем подряд идущие даты с отметками.
// Если сегодня не отмечено, но вчера — стрик не сломан (держится до конца дня).
//
// Реализация для MVP — на Go (выбрать все даты, посчитать в цикле).
// Если станет горячо — SQL-запрос с window function или кэш в Redis.
func (r *CompletionRepository) Streak(ctx context.Context, habitID uuid.UUID, today time.Time) (int, error) {
	// TODO
	return 0, nil
}