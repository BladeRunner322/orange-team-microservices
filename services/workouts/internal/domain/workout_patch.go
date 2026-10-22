package domain

import "github.com/BladeRunner322/orange-team-microservices/pkg/nullable"

// WorkoutPatch — патч тренировки.
//
// Status — NOT NULL в БД, два состояния: не трогать (nil) / установить.
// Intensity — nullable в БД, три состояния: не трогать (Set=false) /
// сбросить в NULL (Set=true, Value=nil) / установить (Set=true, Value=&v).
type WorkoutPatch struct {
	Status    *Status
	Intensity nullable.Nullable[Intensity]
}

func NewWorkoutPatch(
	status *Status,
	intensity nullable.Nullable[Intensity],
) WorkoutPatch {
	return WorkoutPatch{
		Status:    status,
		Intensity: intensity,
	}
}
