package habits

import "time"

// CreateHabitRequest — тело запроса POST /habits.
type CreateHabitRequest struct {
	Name        string `json:"name"        validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
}

// PatchHabitRequest — тело запроса PATCH /habits/{id}.
type PatchHabitRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=1,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

// CompleteHabitRequest — тело запроса POST /habits/{id}/complete.
type CompleteHabitRequest struct {
	Date string `json:"date" validate:"omitempty,datetime=2006-01-02"`
}

// UncompleteHabitRequest — тело запроса DELETE /habits/{id}/complete.
type UncompleteHabitRequest struct {
	Date string `json:"date" validate:"omitempty,datetime=2006-01-02"`
}

// HabitResponse — HTTP-представление привычки.
type HabitResponse struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Streak         int        `json:"streak"`
	CompletedToday bool       `json:"completed_today"`
	IsDeleted      bool       `json:"is_deleted"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
}

// CompletionResponse — HTTP-представление отметки.
type CompletionResponse struct {
	HabitID   string    `json:"habit_id"`
	Date      string    `json:"date"`
	CreatedAt time.Time `json:"created_at"`
}
