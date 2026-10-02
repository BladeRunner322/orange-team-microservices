// Package domain — доменные ошибки Exercises-сервиса.
package domain

import "errors"

var (
	ErrExerciseNotFound    = errors.New("exercise not found")
	ErrInvalidName         = errors.New("invalid name")
	ErrInvalidDescription  = errors.New("invalid description")
	ErrInvalidDifficulty   = errors.New("invalid difficulty")
	ErrInvalidExerciseType = errors.New("invalid exercise type")
	ErrExerciseNameExists  = errors.New("exercise with this name already exists")
)
