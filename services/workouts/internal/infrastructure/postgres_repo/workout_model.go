package postgres_repo

import (
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/google/uuid"
)

type WorkoutModel struct {
	ID                       uuid.UUID
	Version                  int
	UserID                   uuid.UUID
	Status                   string
	StartedAt                *time.Time
	CompletedAt              *time.Time
	WorkoutScore             int
	Intensity                *int
	PersonalScoreCoefficient int
	DeletedAt                *time.Time
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

func (m *WorkoutModel) Scan(row postgres.Row) error {
	return row.Scan(
		&m.ID,
		&m.Version,
		&m.UserID,
		&m.Status,
		&m.StartedAt,
		&m.CompletedAt,
		&m.WorkoutScore,
		&m.Intensity,
		&m.PersonalScoreCoefficient,
		&m.DeletedAt,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
}
