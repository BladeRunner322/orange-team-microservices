package postgres_repo

import (
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/google/uuid"
)

// ProfileModel — модель для маппинга с таблицей profiles.users.
type ProfileModel struct {
	UserID      uuid.UUID
	Sex         *string
	WeightGrams *int
	BirthDate   *time.Time
	HeightCM    *int
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

// Scan читает строку из БД в модель.
func (m *ProfileModel) Scan(row postgres.Row) error {
	return row.Scan(
		&m.UserID,
		&m.Sex,
		&m.WeightGrams,
		&m.BirthDate,
		&m.HeightCM,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
}
