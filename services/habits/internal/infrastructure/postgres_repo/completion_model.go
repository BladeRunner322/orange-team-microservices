package postgres_repo

import (
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/google/uuid"
)

// CompletionModel — модель для таблицы habits.habit_completions.
type CompletionModel struct {
	HabitID   uuid.UUID
	Date      time.Time
	CreatedAt time.Time
}

// Scan читает строку из БД в модель.
func (m *CompletionModel) Scan(row postgres.Row) error {
	return row.Scan(
		&m.HabitID,
		&m.Date,
		&m.CreatedAt,
	)
}
