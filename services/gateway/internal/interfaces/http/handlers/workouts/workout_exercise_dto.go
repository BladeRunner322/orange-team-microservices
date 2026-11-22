package workouts

import "time"

// CreateWorkoutExerciseRequest — тело запроса POST /workouts/{workoutId}/exercises.
type CreateWorkoutExerciseRequest struct {
	ExerciseID string `json:"exercise_id" validate:"required"`
	Weight     *int32 `json:"weight"      validate:"omitempty,min=1,max=1000"`
	Sets       *int32 `json:"sets"        validate:"omitempty,min=1,max=100"`
	Reps       *int32 `json:"reps"        validate:"omitempty,min=1,max=1000"`
	Duration   *int32 `json:"duration"    validate:"omitempty,min=1,max=86400"`
	Completed  bool   `json:"completed"`
}

// PatchWorkoutExerciseRequest — тело запроса PATCH для упражнения в тренировке.
type PatchWorkoutExerciseRequest struct {
	Weight    *int32 `json:"weight"    validate:"omitempty,min=1,max=1000"`
	Sets      *int32 `json:"sets"      validate:"omitempty,min=1,max=100"`
	Reps      *int32 `json:"reps"      validate:"omitempty,min=1,max=1000"`
	Duration  *int32 `json:"duration"  validate:"omitempty,min=1,max=86400"`
	Completed *bool  `json:"completed"`
}

// WorkoutExerciseResponse — HTTP-представление упражнения в тренировке.
type WorkoutExerciseResponse struct {
	ID           string     `json:"id"`
	Version      int32      `json:"version"`
	WorkoutID    string     `json:"workout_id"`
	ExerciseID   string     `json:"exercise_id"`
	Difficulty   int32      `json:"difficulty"`
	Weight       *int32     `json:"weight"`
	Sets         *int32     `json:"sets"`
	Reps         *int32     `json:"reps"`
	Duration     *int32     `json:"duration"`
	Completed    bool       `json:"completed"`
	ExerciseLoad int32      `json:"exercise_load"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}
