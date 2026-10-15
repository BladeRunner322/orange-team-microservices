package domain

import (
	"time"

	"github.com/google/uuid"
)

// Completion — факт выполнения привычки в конкретный день.
type Completion struct {
	habitID   uuid.UUID
	date      Date
	createdAt time.Time
}

// NewCompletion создаёт новую отметку (для INSERT).
// date уже нормализована VO Date (UTC 00:00).
func NewCompletion(habitID uuid.UUID, date Date) Completion {
	return Completion{
		habitID:   habitID,
		date:      date,
		createdAt: time.Now().UTC(),
	}
}

// RestoreCompletion восстанавливает отметку из БД.
func RestoreCompletion(habitID uuid.UUID, date Date, createdAt time.Time) Completion {
	return Completion{
		habitID: habitID,
		date: date,
		createdAt: createdAt,
	}
}

// Геттеры
func (c Completion) HabitID() uuid.UUID   { return c.habitID }
func (c Completion) Date() Date           { return c.date }
func (c Completion) CreatedAt() time.Time { return c.createdAt }
