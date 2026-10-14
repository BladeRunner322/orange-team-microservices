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
	createdAt   time.Time
	updatedAt   *time.Time
	deletedAt   *time.Time
}

// NewHabit создаёт новую привычку с сгенерированным id.
func NewHabit(userID uuid.UUID, name Name, description Description) Habit {
	// TODO: uuid.New(), time.Now().UTC()
	return Habit{}
}

// RestoreHabit восстанавливает привычку из БД (для маппинга).
func RestoreHabit(
	id, userID uuid.UUID,
	name Name,
	description Description,
	createdAt time.Time,
	updatedAt *time.Time,
	deletedAt *time.Time,
) Habit {
	// TODO
	return Habit{}
}

// ApplyPatch применяет патч: каждое поле с не-nil значением устанавливается.
func (h *Habit) ApplyPatch(patch HabitPatch) {
	// TODO
}

// IsDeleted — true, если deletedAt != nil.
func (h Habit) IsDeleted() bool {
	// TODO
	return false
}

// Геттеры (публичные)
func (h Habit) ID() uuid.UUID            { return h.id }
func (h Habit) UserID() uuid.UUID        { return h.userID }
func (h Habit) Name() Name               { return h.name }
func (h Habit) Description() Description { return h.description }
func (h Habit) CreatedAt() time.Time     { return h.createdAt }
func (h Habit) UpdatedAt() *time.Time    { return h.updatedAt }
func (h Habit) DeletedAt() *time.Time    { return h.deletedAt }
