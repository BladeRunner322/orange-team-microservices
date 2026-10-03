package postgres_repo

import (
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/google/uuid"
)

// ExerciseModel — модель для маппинга с таблицей exercises.exercises.
type ExerciseModel struct {
	ID          uuid.UUID
	Name        string
	Description string
	Difficulty  int
	Type        string
	DeletedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

// Scan читает строку из БД в модель.
func (m *ExerciseModel) Scan(row postgres.Row) error {
	return row.Scan(
		&m.ID,
		&m.Name,
		&m.Description,
		&m.Difficulty,
		&m.Type,
		&m.DeletedAt,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
}
