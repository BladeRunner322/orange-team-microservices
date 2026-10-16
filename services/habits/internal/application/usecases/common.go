package usecases

import (
	"context"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/ctxkeys"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	"github.com/google/uuid"
)

// maxBackdateDays — сколько дней назад разрешено отмечать/снимать выполнение.
// Симметрично для CompleteHabit и UncompleteHabit (ADR-023 п.4).
const maxBackdateDays = 7

// userIDFromContext извлекает user_id из context и парсит в uuid.UUID.
// Возвращает domain.ErrUnauthenticated, если user_id отсутствует или невалиден.
func userIDFromContext(ctx context.Context, log *logger.Logger) (uuid.UUID, error) {
	userIDStr, ok := ctxkeys.UserIDFromContext(ctx)
	if !ok {
		log.Warn("user_id is missing in context")
		return uuid.Nil, domain.ErrUnauthenticated
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Warn("invalid user_id format", "user_id", userIDStr, "error", err)
		return uuid.Nil, domain.ErrUnauthenticated
	}

	return userID, nil
}

// validateCompletionDate проверяет бизнес-правила для даты отметки:
// не в будущем (после сегодня в UTC), не старше maxBackdateDays.
//
// Формат даты (zero time, нормализация в UTC 00:00) — забота VO domain.Date.
// Здесь — только диапазон.
func validateCompletionDate(date time.Time, now time.Time) error {
	today := dayOnly(now)
	d := dayOnly(date)

	if d.After(today) {
		return domain.ErrInvalidDate
	}

	earliest := today.AddDate(0, 0, -maxBackdateDays)
	if d.Before(earliest) {
		return domain.ErrInvalidDate
	}

	return nil
}

// dayOnly обрезает время до календарной даты (UTC).
// Дублирует хелпер из postgres_repo — там он приватный, чтобы не тащить
// зависимость usecase → infrastructure.
func dayOnly(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

// mustDate конструирует domain.Date из time.Time, который уже прошёл
// проверку validateCompletionDate. Zero time сюда попасть не может:
// он отсекается на NewDate → ErrInvalidDate.
func mustDate(t time.Time) domain.Date {
	d, _ := domain.NewDate(t)
	return d
}
