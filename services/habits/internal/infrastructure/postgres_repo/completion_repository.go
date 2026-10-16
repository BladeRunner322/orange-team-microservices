package postgres_repo

import (
	"context"
	"fmt"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	"github.com/google/uuid"
)

// CompletionRepository реализует ports.CompletionRepository для PostgreSQL.
type CompletionRepository struct {
	pool      postgres.Pool // nil, если репозиторий привязан к транзакции
	tx        postgres.Tx   // nil, если репозиторий не в транзакции
	opTimeout time.Duration
}

// NewCompletionRepository создаёт новый репозиторий с пулом.
func NewCompletionRepository(pool postgres.Pool) *CompletionRepository {
	return &CompletionRepository{
		pool:      pool,
		opTimeout: pool.OpTimeout(),
	}
}

// querier возвращает актуальный источник запросов: транзакцию или пул.
func (r *CompletionRepository) querier() postgres.Querier {
	if r.tx != nil {
		return r.tx
	}
	return r.pool
}

// WithTx выполняет fn внутри транзакции. Если репозиторий уже привязан
// к транзакции — вложенная не открывается, fn вызывается с текущим.
func (r *CompletionRepository) WithTx(ctx context.Context, fn func(ports.CompletionRepository) error) error {
	if r.tx != nil {
		return fn(r)
	}
	return r.pool.WithTx(ctx, func(tx postgres.Tx) error {
		txRepo := &CompletionRepository{tx: tx, opTimeout: r.opTimeout}
		return fn(txRepo)
	})
}

// Complete добавляет отметку. ON CONFLICT DO NOTHING — повторная отметка не падает.
func (r *CompletionRepository) Complete(ctx context.Context, completion domain.Completion) error {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		INSERT INTO habits.habit_completions (habit_id, date)
		VALUES ($1, $2)
		ON CONFLICT (habit_id, date) DO NOTHING
	`

	m := CompletionDomainToModel(completion)

	if _, err := r.querier().Exec(ctx, query, m.HabitID, m.Date); err != nil {
		return fmt.Errorf("complete habit: %w", err)
	}

	return nil
}

// Uncomplete удаляет отметку за дату. Идемпотентно.
func (r *CompletionRepository) Uncomplete(ctx context.Context, habitID uuid.UUID, date time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		DELETE FROM habits.habit_completions
		WHERE habit_id = $1 AND date = $2
	`

	if _, err := r.querier().Exec(ctx, query, habitID, date); err != nil {
		return fmt.Errorf("uncomplete habit: %w", err)
	}

	return nil
}

// IsCompleted — проверяет наличие отметки на дату.
func (r *CompletionRepository) IsCompleted(ctx context.Context, habitID uuid.UUID, date time.Time) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM habits.habit_completions
			WHERE habit_id = $1 AND date = $2
		)
	`

	var exists bool
	if err := r.querier().QueryRow(ctx, query, habitID, date).Scan(&exists); err != nil {
		return false, fmt.Errorf("is completed: %w", err)
	}

	return exists, nil
}

// ListByHabit возвращает отметки за период [from, to] включительно.
// Сортировка по date ASC.
func (r *CompletionRepository) ListByHabit(ctx context.Context, habitID uuid.UUID, from, to time.Time) ([]domain.Completion, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		SELECT habit_id, date, created_at
		FROM habits.habit_completions
		WHERE habit_id = $1 AND date BETWEEN $2 AND $3
		ORDER BY date ASC
	`

	rows, err := r.querier().Query(ctx, query, habitID, from, to)
	if err != nil {
		return nil, fmt.Errorf("query completions: %w", err)
	}
	defer rows.Close()

	completions := make([]domain.Completion, 0)
	for rows.Next() {
		var m CompletionModel
		if err := m.Scan(rows); err != nil {
			return nil, fmt.Errorf("scan completion: %w", err)
		}

		completion, err := CompletionModelToDomain(m)
		if err != nil {
			return nil, fmt.Errorf("completion %s/%s: %w", m.HabitID, m.Date, err)
		}

		completions = append(completions, completion)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate completions: %w", err)
	}

	return completions, nil
}

// Streak считает текущий стрик привычки.
//
// Логика: идём от today назад, считаем подряд идущие даты с отметками.
// Если сегодня не отмечено, но вчера — стрик не сломан (держится до конца дня).
//
// Реализация для MVP — на Go (выбрать все даты, посчитать в цикле).
// Если станет горячо — SQL-запрос с window function или кэш в Redis.
func (r *CompletionRepository) Streak(ctx context.Context, habitID uuid.UUID, today time.Time) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		SELECT date
		FROM habits.habit_completions
		WHERE habit_id = $1 AND date <= $2
		ORDER BY date DESC
	`

	rows, err := r.querier().Query(ctx, query, habitID, today)
	if err != nil {
		return 0, fmt.Errorf("query streak dates: %w", err)
	}
	defer rows.Close()

	dates := make([]time.Time, 0)
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			return 0, fmt.Errorf("scan streak date: %w", err)
		}
		dates = append(dates, d)
	}

	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate streak dates: %w", err)
	}

	if len(dates) == 0 {
		return 0, nil
	}

	todayDay := dayOnly(today)
	yesterdayDay := todayDay.AddDate(0, 0, -1)
	firstDay := dayOnly(dates[0])

	// Определяем точку старта: сегодня, вчера, или стрик сломан.
	var expected time.Time
	switch {
	case firstDay.Equal(todayDay):
		expected = todayDay
	case firstDay.Equal(yesterdayDay):
		expected = yesterdayDay
	default:
		return 0, nil
	}

	streak := 0
	for _, d := range dates {
		if !dayOnly(d).Equal(expected) {
			break
		}
		streak++
		expected = expected.AddDate(0, 0, -1)
	}

	return streak, nil
}

// dayOnly обрезает время до календарной даты (UTC).
func dayOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
