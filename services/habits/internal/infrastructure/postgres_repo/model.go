// Package postgres_repo — реализация ports.HabitRepository и ports.CompletionRepository для PostgreSQL.
package postgres_repo

import (
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/google/uuid"
)

// HabitModel — модель для таблицы habits.habits.
type HabitModel struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	DeletedAt   *time.Time
}

// Scan читает строку из БД в модель.
func (m *HabitModel) Scan(row postgres.Row) error {
	// TODO
	return nil
}

// CompletionModel — модель для таблицы habits.habit_completions.
type CompletionModel struct {
	HabitID   uuid.UUID
	Date      time.Time
	CreatedAt time.Time
}

// Scan читает строку из БД в модель.
func (m *CompletionModel) Scan(row postgres.Row) error {
	// TODO
	return nil
}