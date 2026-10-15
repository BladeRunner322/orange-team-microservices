// Package domain — доменная модель Habits-сервиса:
// сущность Habit, Completion, value objects и ошибки.
package domain

import "errors"

var (
	// ErrHabitNotFound — привычка не найдена или удалена.
	ErrHabitNotFound = errors.New("habit not found")

	// ErrInvalidName — name пустой или > 100 символов после trim.
	ErrInvalidName = errors.New("invalid name")

	// ErrInvalidDescription — description > 500 символов после trim.
	ErrInvalidDescription = errors.New("invalid description")

	// ErrInvalidDate — некорректная дата: невалидный формат (VO)
	// или нарушение бизнес-правила (usecase: в будущем, старше 7 дней).
	ErrInvalidDate = errors.New("invalid date")

	// ErrUnauthenticated — нет user_id в context.
	ErrUnauthenticated = errors.New("unauthenticated")
)
