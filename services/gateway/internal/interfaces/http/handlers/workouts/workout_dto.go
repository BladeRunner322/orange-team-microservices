package workouts

import (
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/nullable"
)

// PatchWorkoutRequest — тело запроса PATCH /workouts/{workoutId}.
type PatchWorkoutRequest struct {
	Status    *string                  `json:"status" validate:"omitempty,oneof=planned in_progress completed cancelled"`
	Intensity nullable.Nullable[int32] `json:"intensity"`
}

// WorkoutResponse — HTTP-представление тренировки.
type WorkoutResponse struct {
	ID                       string     `json:"id"`
	Version                  int32      `json:"version"`
	UserID                   string     `json:"user_id"`
	Status                   string     `json:"status"`
	StartedAt                *time.Time `json:"started_at,omitempty"`
	CompletedAt              *time.Time `json:"completed_at,omitempty"`
	WorkoutScore             int32      `json:"workout_score"`
	Intensity                *int32     `json:"intensity"`
	PersonalScoreCoefficient int32      `json:"personal_score_coefficient"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                *time.Time `json:"updated_at,omitempty"`
	IsDeleted                bool       `json:"is_deleted"`
}
