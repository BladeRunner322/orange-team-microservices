package domain

import "github.com/BladeRunner322/orange-team-microservices/pkg/nullable"

// WorkoutPatch — патч тренировки. nil = не трогать.
type WorkoutPatch struct {
	Status    nullable.Nullable[Status]
	Intensity nullable.Nullable[Intensity]
}

func NewWorkoutPatch(
	status nullable.Nullable[Status],
	intensity nullable.Nullable[Intensity],
) WorkoutPatch {
	return WorkoutPatch{
		Status:    status,
		Intensity: intensity,
	}
}
