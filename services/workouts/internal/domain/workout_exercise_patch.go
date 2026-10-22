package domain

// WorkoutExercisePatch — патч упражнения.
//
// nil = не трогать. Третьего состояния «сбросить в NULL» нет:
// для weight-типа weight/sets/reps обязательны, для duration-типа
// duration обязателен (ValidateForWorkoutExerciseType).
// exercise_id immutable — в патче отсутствует (ADR-025 п.9).
type WorkoutExercisePatch struct {
    Weight    *Weight
    Sets      *Sets
    Reps      *Reps
    Duration  *Duration
    Completed *bool
}

func NewWorkoutExercisePatch(
    weight *Weight,
    sets *Sets,
    reps *Reps,
    duration *Duration,
    completed *bool,
) WorkoutExercisePatch {
    return WorkoutExercisePatch{
        Weight:    weight,
        Sets:      sets,
        Reps:      reps,
        Duration:  duration,
        Completed: completed,
    }
}