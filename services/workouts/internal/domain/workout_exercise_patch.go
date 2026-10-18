package domain

import "github.com/BladeRunner322/orange-team-microservices/pkg/nullable"

// WorkoutExercisePatch — патч упражнения. nil = не трогать.
// exercise_id immutable — в патче отсутствует (ADR-025 п.9).
type WorkoutExercisePatch struct {
	Weight    nullable.Nullable[Weight]
	Sets      nullable.Nullable[Sets]
	Reps      nullable.Nullable[Reps]
	Duration  nullable.Nullable[Duration]
	Completed nullable.Nullable[bool]
}

func NewWorkoutExercisePatch(
	weight nullable.Nullable[Weight],
	sets nullable.Nullable[Sets],
	reps nullable.Nullable[Reps],
	duration nullable.Nullable[Duration],
	completed nullable.Nullable[bool],
) WorkoutExercisePatch {
	return WorkoutExercisePatch{
		Weight:    weight,
		Sets:      sets,
		Reps:      reps,
		Duration:  duration,
		Completed: completed,
	}
}
