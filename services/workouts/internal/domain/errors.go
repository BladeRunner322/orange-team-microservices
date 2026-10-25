package domain

import "errors"

// Lookup — сущность не найдена.
var (
	ErrWorkoutNotFound         = errors.New("workout not found")
	ErrWorkoutExerciseNotFound = errors.New("workout exercise not found")
)

// Workout — бизнес-правила агрегата Workout (state machine, завершение).
var (
	ErrInvalidStatusTransition        = errors.New("invalid status transition")
	ErrCannotCompleteEmptyWorkout     = errors.New("cannot complete workout with no completed exercises")
	ErrCannotModifyWorkoutExercise    = errors.New("cannot modify workout exercise in terminal status")
	ErrCannotMarkExerciseCompleted    = errors.New("can only mark exercise as completed when workout is in progress")
)

// WorkoutExercise — бизнес-правила компонента агрегата.
var (
	ErrInvalidWorkoutExerciseForType = errors.New("workout exercise fields do not match exercise type")
)

// Value objects — ошибки валидации VO.
var (
	ErrInvalidStatus       = errors.New("invalid status")
	ErrInvalidWeight       = errors.New("invalid weight")
	ErrInvalidSets         = errors.New("invalid sets")
	ErrInvalidReps         = errors.New("invalid reps")
	ErrInvalidDuration     = errors.New("invalid duration")
	ErrInvalidDifficulty   = errors.New("invalid difficulty")
	ErrInvalidIntensity    = errors.New("invalid intensity")
	ErrInvalidExerciseType = errors.New("invalid exercise type")
)

// Auth — ошибки аутентификации.
var (
	ErrUnauthenticated = errors.New("unauthenticated")
)
