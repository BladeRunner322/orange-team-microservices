package postgres_repo

import (
	"time"

	"github.com/google/uuid"
)

// UserModel — модель для маппинга с таблицей auth.users.
type UserModel struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FullName     string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}
