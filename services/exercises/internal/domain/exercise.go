package domain

import (
	"time"

	"github.com/google/uuid"
)

// Exercise — доменная сущность упражнения.
type Exercise struct {
	id           uuid.UUID
	name         Name
	description  Description
	difficulty   Difficulty
	exerciseType ExerciseType
	isDeleted    bool
	createdAt    time.Time
	updatedAt    *time.Time
}

// NewExercise создаёт упражнение с указанными полями.
func NewExercise(
	name Name,
	description Description,
	difficulty Difficulty,
	exerciseType ExerciseType,
) Exercise {
	return Exercise{
		id:           uuid.New(),
		name:         name,
		description:  description,
		difficulty:   difficulty,
		exerciseType: exerciseType,
		isDeleted:    false,
		createdAt:    time.Now().UTC(),
		updatedAt:    nil,
	}
}

// RestoreExercise восстанавливает упражнение из БД (для маппинга).
func RestoreExercise(
	id uuid.UUID,
	name Name,
	description Description,
	difficulty Difficulty,
	exerciseType ExerciseType,
	isDeleted bool,
	createdAt time.Time,
	updatedAt *time.Time,
) Exercise {
	return Exercise{
		id:           id,
		name:         name,
		description:  description,
		difficulty:   difficulty,
		exerciseType: exerciseType,
		isDeleted:    isDeleted,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// ApplyPatch применяет патч: каждое поле с не-nil значением устанавливается,
// nil означает «не трогать».
func (e *Exercise) ApplyPatch(patch ExercisePatch) {
	if patch.Name != nil {
		e.name = *patch.Name
	}
	if patch.Description != nil {
		e.description = *patch.Description
	}
	if patch.Difficulty != nil {
		e.difficulty = *patch.Difficulty
	}
}

// Геттеры (публичные)
func (e Exercise) ID() uuid.UUID              { return e.id }
func (e Exercise) Name() Name                 { return e.name }
func (e Exercise) Description() Description   { return e.description }
func (e Exercise) Difficulty() Difficulty     { return e.difficulty }
func (e Exercise) ExerciseType() ExerciseType { return e.exerciseType }
func (e Exercise) IsDeleted() bool            { return e.isDeleted }
func (e Exercise) CreatedAt() time.Time       { return e.createdAt }
func (e Exercise) UpdatedAt() *time.Time      { return e.updatedAt }
