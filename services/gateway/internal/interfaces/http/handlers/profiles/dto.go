package profiles

import (
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/nullable"
)

// PatchUserRequest — тело запроса PATCH /users/me.
// Nullable-поля позволяют отличить «не трогать» от «сбросить в NULL».
type PatchUserRequest struct {
	Sex       nullable.Nullable[string]  `json:"sex"`
	WeightKg  nullable.Nullable[float64] `json:"weight_kg"`
	BirthDate nullable.Nullable[string]  `json:"birth_date"`
	HeightCm  nullable.Nullable[int32]   `json:"height_cm"`
}

// UserProfileResponse — ответ с профилем пользователя.
type UserProfileResponse struct {
	UserID           string     `json:"user_id"`
	Sex              string     `json:"sex,omitempty"`
	WeightKg         float64    `json:"weight_kg,omitempty"`
	BirthDate        string     `json:"birth_date,omitempty"`
	HeightCm         int32      `json:"height_cm,omitempty"`
	ProfileCompleted bool       `json:"profile_completed"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}
