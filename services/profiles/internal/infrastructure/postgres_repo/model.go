package postgres_repo

import (
	"time"

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
