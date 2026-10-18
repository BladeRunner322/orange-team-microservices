package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	"github.com/google/uuid"
)

// defaultCompletionsWindow — окно по умолчанию, если период не задан.
// ADR-023 п.8: пустой период → последние 30 дней.
const defaultCompletionsWindow = 30 * 24 * time.Hour

type GetCompletions struct {
	habitRepo      ports.HabitRepository
	completionRepo ports.CompletionRepository
	logger         *logger.Logger
}

func NewGetCompletions(habitRepo ports.HabitRepository, completionRepo ports.CompletionRepository, log *logger.Logger) *GetCompletions {
	return &GetCompletions{
		habitRepo:      habitRepo,
		completionRepo: completionRepo,
		logger:         log,
	}
}

// Execute возвращает историю выполнений привычки за период [from, to].
//
// Если период не задан (zero time) — используются последние 30 дней.
// Если задан только from или только to — соответствующая граница
// заменяется на «сегодня».
//
// Проверяет, что привычка принадлежит текущему пользователю.
// Удалённая привычка допустима: история остаётся (ADR-023 п.6).
func (uc *GetCompletions) Execute(ctx context.Context, id uuid.UUID, from, to time.Time) ([]domain.Completion, error) {
	log := uc.logger.With("operation", "GetCompletions", "habit_id", id)

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return nil, err
	}

	habit, err := uc.habitRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrHabitNotFound) {
			log.Warn("habit not found")
			return nil, domain.ErrHabitNotFound
		}

		log.Error("failed to get habit", "error", err)
		return nil, fmt.Errorf("get habit: %w", err)
	}

	if habit.UserID() != userID {
		log.Warn("habit belongs to another user")
		return nil, domain.ErrHabitNotFound
	}

	fromDate, toDate, err := resolvePeriod(from, to, time.Now().UTC())
	if err != nil {
		log.Warn("invalid period", "error", err)
		return nil, err
	}

	completions, err := uc.completionRepo.ListByHabit(ctx, id, fromDate, toDate)
	if err != nil {
		log.Error("failed to list completions", "error", err)
		return nil, fmt.Errorf("list completions: %w", err)
	}

	return completions, nil
}

// resolvePeriod приводит входной период к двум границам [from, to].
//
// Правила:
//   - обе границы zero → последние 30 дней;
//   - только from zero → from = to - 30 дней;
//   - только to zero → to = today;
//   - from > to → ErrInvalidDate.
//
// Возвращает границы, нормализованные до календарной даты (UTC 00:00).
func resolvePeriod(from, to, now time.Time) (time.Time, time.Time, error) {
	fromZero := from.IsZero()
	toZero := to.IsZero()

	var fromDate, toDate time.Time

	switch {
	case fromZero && toZero:
		toDate = dayOnly(now)
		fromDate = toDate.Add(-defaultCompletionsWindow)
	case fromZero:
		toDate = dayOnly(to)
		fromDate = toDate.Add(-defaultCompletionsWindow)
	case toZero:
		fromDate = dayOnly(from)
		toDate = dayOnly(now)
	default:
		fromDate = dayOnly(from)
		toDate = dayOnly(to)
	}

	if fromDate.After(toDate) {
		return time.Time{}, time.Time{}, domain.ErrInvalidDate
	}

	return fromDate, toDate, nil
}
