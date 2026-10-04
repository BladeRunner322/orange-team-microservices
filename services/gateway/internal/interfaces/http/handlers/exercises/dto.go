package exercises

import "time"

// CreateExerciseRequest — тело запроса POST /exercises.
type CreateExerciseRequest struct {
	Name        string `json:"name"        validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"required,min=1,max=1000"`
	Difficulty  int32  `json:"difficulty"  validate:"required,min=1,max=10"`
	Type        string `json:"type"        validate:"required,oneof=weight duration"`
}

// PatchExerciseRequest — тело запроса PATCH /exercises/{id}.
//
// Указатели позволяют отличить «не трогать» (nil) от «установить значение».
// У Exercises все поля NOT NULL, поэтому сбросить в NULL нельзя —
// third state (explicit null) отсутствует по определению (см. ADR-008 п.6).
type PatchExerciseRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=3,max=100"`
	Description *string `json:"description" validate:"omitempty,min=1,max=1000"`
	Difficulty  *int32  `json:"difficulty"  validate:"omitempty,min=1,max=10"`
}

// ExerciseResponse — HTTP-представление упражнения.
type ExerciseResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Difficulty  int32      `json:"difficulty"`
	Type        string     `json:"type"`
	IsDeleted   bool       `json:"is_deleted"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}
