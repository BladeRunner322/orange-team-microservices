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
	DeletedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

// Scan читает строку из БД в модель.
func (m *HabitModel) Scan(row postgres.Row) error {
	return row.Scan(
		&m.ID,
		&m.UserID,
		&m.Name,
		&m.Description,
		&m.DeletedAt,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
}
