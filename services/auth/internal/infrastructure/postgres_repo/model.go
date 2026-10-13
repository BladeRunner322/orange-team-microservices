package postgres_repo

import (
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/google/uuid"
)

// UserModel — модель для маппинга с таблицей auth.users.
type UserModel struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FullName     string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

// Scan читает строку из БД в модель.
func (m *UserModel) Scan(row postgres.Row) error {
	return row.Scan(
		&m.ID,
		&m.Email,
		&m.PasswordHash,
		&m.FullName,
		&m.Role,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
}
