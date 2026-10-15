package domain

import (
	"time"

	"github.com/google/uuid"
)

// Habit — доменная сущность привычки.
type Habit struct {
	id          uuid.UUID
	userID      uuid.UUID
	name        Name
	description Description
	isDeleted   bool
	createdAt   time.Time
	updatedAt   *time.Time
}

// NewHabit создаёт новую привычку с сгенерированным id.
func NewHabit(userID uuid.UUID, name Name, description Description) Habit {
	return Habit{
		id:          uuid.New(),
		userID:      userID,
		name:        name,
		description: description,
		isDeleted:   false,
		createdAt:   time.Now().UTC(),
		updatedAt:   nil,
	}
}

// RestoreHabit восстанавливает привычку из БД (для маппинга).
func RestoreHabit(
	id uuid.UUID,
	userID uuid.UUID,
	name Name,
	description Description,
	isDeleted bool,
	createdAt time.Time,
	updatedAt *time.Time,
) Habit {
	return Habit{
		id:          id,
		userID:      userID,
		name:        name,
		description: description,
		isDeleted:   isDeleted,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// ApplyPatch применяет патч: каждое поле с не-nil значением устанавливается.
// nil означает «не трогать».
func (h *Habit) ApplyPatch(patch HabitPatch) {
	if patch.Name != nil {
		h.name = *patch.Name
	}
	if patch.Description != nil {
		h.description = *patch.Description
	}
}

// Геттеры (публичные)
func (h Habit) ID() uuid.UUID            { return h.id }
func (h Habit) UserID() uuid.UUID        { return h.userID }
func (h Habit) Name() Name               { return h.name }
func (h Habit) Description() Description { return h.description }
func (h Habit) IsDeleted() bool          { return h.isDeleted }
func (h Habit) CreatedAt() time.Time     { return h.createdAt }
func (h Habit) UpdatedAt() *time.Time    { return h.updatedAt }
