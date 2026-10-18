package domain

import "errors"

var (
	ErrWorkoutNotFound             = errors.New("workout not found")
	ErrWorkoutExerciseNotFound     = errors.New("workout exercise not found")
	ErrInvalidStatus               = errors.New("invalid status")
	ErrInvalidStatusTransition     = errors.New("invalid status transition")
	ErrCannotCompleteEmptyWorkout  = errors.New("cannot complete workout with no completed exercises")
	ErrCannotModifyWorkoutExercise = errors.New("cannot modify workout exercise in terminal status")

	ErrInvalidWeight       = errors.New("invalid weight")
	ErrInvalidSets         = errors.New("invalid sets")
	ErrInvalidReps         = errors.New("invalid reps")
	ErrInvalidDuration     = errors.New("invalid duration")
	ErrInvalidDifficulty   = errors.New("invalid difficulty")
	ErrInvalidIntensity    = errors.New("invalid intensity")
	ErrInvalidExerciseType = errors.New("invalid exercise type")

	ErrInvalidWorkoutExerciseForType = errors.New("workout exercise fields do not match exercise type")

	ErrUnauthenticated = errors.New("unauthenticated")
)
