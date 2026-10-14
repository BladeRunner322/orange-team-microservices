package domain

import (
	"time"

	"github.com/google/uuid"
)

// Completion — факт выполнения привычки в конкретный день.
// Дата хранится без времени, в UTC.
type Completion struct {
	habitID   uuid.UUID
	date      time.Time
	createdAt time.Time
}

// NewCompletion создаёт новую отметку (для INSERT).
// date нормализуется до UTC, обнуляется время.
func NewCompletion(habitID uuid.UUID, date time.Time) Completion {
	// TODO
	return Completion{}
}

// RestoreCompletion восстанавливает отметку из БД.
func RestoreCompletion(habitID uuid.UUID, date, createdAt time.Time) Completion {
	// TODO
	return Completion{}
}

// Геттеры
func (c Completion) HabitID() uuid.UUID   { return c.habitID }
func (c Completion) Date() time.Time      { return c.date }
func (c Completion) CreatedAt() time.Time { return c.createdAt }
