// Package ports — интерфейсы, которые domain требует от инфраструктуры.
package ports

import (
	"context"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	"github.com/google/uuid"
)

// HabitRepository — работа с таблицей habits.habits.
type HabitRepository interface {
	// WithTx выполняет fn в транзакции.
	// Нужен для PatchHabit (read-then-write, по образцу Exercises).
	WithTx(ctx context.Context, fn func(HabitRepository) error) error

	// Create сохраняет новую привычку.
	Create(ctx context.Context, habit domain.Habit) (domain.Habit, error)

	// GetByID возвращает привычку по id (включая удалённые).
	GetByID(ctx context.Context, id uuid.UUID) (domain.Habit, error)

	// GetByIDForUpdate возвращает привычку с блокировкой строки.
	// Вызывается только внутри WithTx.
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (domain.Habit, error)

	// ListByUser возвращает активные привычки пользователя.
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Habit, error)

	// Update обновляет name/description/updated_at.
	Update(ctx context.Context, habit domain.Habit) (domain.Habit, error)

	// MarkDeleted помечает привычку удалённой (soft delete).
	// Идемпотентно.
	MarkDeleted(ctx context.Context, id uuid.UUID) error
}

// CompletionRepository — работа с таблицей habits.habit_completions.
type CompletionRepository interface {
	// Complete добавляет отметку. Если запись уже есть — no-op.
	// Защита через ON CONFLICT DO NOTHING.
	Complete(ctx context.Context, completion domain.Completion) error

	// Uncomplete удаляет отметку за дату. Если нет — no-op.
	Uncomplete(ctx context.Context, habitID uuid.UUID, date time.Time) error

	// IsCompleted — проверяет, отмечено ли за конкретную дату.
	IsCompleted(ctx context.Context, habitID uuid.UUID, date time.Time) (bool, error)

	// ListByHabit возвращает отметки за период [from, to] включительно.
	ListByHabit(ctx context.Context, habitID uuid.UUID, from, to time.Time) ([]domain.Completion, error)

	// Streak считает текущий стрик (сколько подряд идущих дней до today включительно).
	// Если сегодня не отмечено, но вчера было — стрик не сломан (держится до конца дня).
	Streak(ctx context.Context, habitID uuid.UUID, today time.Time) (int, error)
}
