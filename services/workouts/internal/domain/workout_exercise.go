package domain

import (
	"time"

	"github.com/google/uuid"
)

// WorkoutExercise — доменная сущность упражнения внутри тренировки.
type WorkoutExercise struct {
	id           uuid.UUID
	version      int
	workoutID    uuid.UUID
	exerciseID   uuid.UUID
	difficulty   Difficulty
	weight       *Weight
	sets         *Sets
	reps         *Reps
	duration     *Duration
	completed    bool
	exerciseLoad int
	createdAt    time.Time
	updatedAt    time.Time
}

// NewWorkoutExercise создаёт новое упражнение в тренировке.
// exercise_load вычисляется в зависимости от типа упражнения.
//
// Возвращает ошибку, если набор полей не соответствует exerciseType
// (см. validateForWorkoutExerciseType). Уже-валидированные VO
// (Weight, Sets, Reps, Duration) — ответственность вызывающего.
func NewWorkoutExercise(
	workoutID uuid.UUID,
	exerciseID uuid.UUID,
	difficulty Difficulty,
	weight *Weight,
	sets *Sets,
	reps *Reps,
	duration *Duration,
	completed bool,
	exerciseType ExerciseType,
) (WorkoutExercise, error) {
	now := time.Now().UTC()

	e := WorkoutExercise{
		id:         uuid.New(),
		workoutID:  workoutID,
		exerciseID: exerciseID,
		difficulty: difficulty,
		weight:     weight,
		sets:       sets,
		reps:       reps,
		duration:   duration,
		completed:  completed,
		createdAt:  now,
		updatedAt:  now,
	}

	if err := e.validateForWorkoutExerciseType(exerciseType); err != nil {
		return WorkoutExercise{}, err
	}

	e.recalculateExerciseLoad(exerciseType)

	return e, nil
}

// RestoreWorkoutExercise восстанавливает упражнение из БД.
func RestoreWorkoutExercise(
	id uuid.UUID,
	version int,
	workoutID uuid.UUID,
	exerciseID uuid.UUID,
	difficulty Difficulty,
	weight *Weight,
	sets *Sets,
	reps *Reps,
	duration *Duration,
	completed bool,
	exerciseLoad int,
	createdAt time.Time,
	updatedAt time.Time,
) WorkoutExercise {
	return WorkoutExercise{
		id:           id,
		version:      version,
		workoutID:    workoutID,
		exerciseID:   exerciseID,
		difficulty:   difficulty,
		weight:       weight,
		sets:         sets,
		reps:         reps,
		duration:     duration,
		completed:    completed,
		exerciseLoad: exerciseLoad,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// Геттеры.
func (e WorkoutExercise) ID() uuid.UUID          { return e.id }
func (e WorkoutExercise) Version() int           { return e.version }
func (e WorkoutExercise) WorkoutID() uuid.UUID   { return e.workoutID }
func (e WorkoutExercise) ExerciseID() uuid.UUID  { return e.exerciseID }
func (e WorkoutExercise) Difficulty() Difficulty { return e.difficulty }
func (e WorkoutExercise) Weight() *Weight        { return e.weight }
func (e WorkoutExercise) Sets() *Sets            { return e.sets }
func (e WorkoutExercise) Reps() *Reps            { return e.reps }
func (e WorkoutExercise) Duration() *Duration    { return e.duration }
func (e WorkoutExercise) Completed() bool        { return e.completed }
func (e WorkoutExercise) ExerciseLoad() int      { return e.exerciseLoad }
func (e WorkoutExercise) CreatedAt() time.Time   { return e.createdAt }
func (e WorkoutExercise) UpdatedAt() time.Time   { return e.updatedAt }

// ApplyPatch применяет патч и пересчитывает exercise_load:
// каждое поле с не-nil значением устанавливается,
// nil означает «не трогать».
//
// Completed не зависит от типа и применяется всегда.
// После применения — валидация итогового состояния и пересчёт нагрузки:
// иначе usecase с recalculateScore получит score по старым данным.
func (e *WorkoutExercise) ApplyPatch(patch WorkoutExercisePatch, t ExerciseType) error {
	changed := false

	if err := e.validateForWorkoutExerciseType(t); err != nil {
		return err
	}

	if patch.Completed != nil {
		e.completed = *patch.Completed
		changed = true
	}

	switch t {
	case ExerciseTypeWeight:
		if patch.Weight != nil {
			e.weight = patch.Weight
			changed = true
		}
		if patch.Sets != nil {
			e.sets = patch.Sets
			changed = true
		}
		if patch.Reps != nil {
			e.reps = patch.Reps
			changed = true
		}
	case ExerciseTypeDuration:
		if patch.Duration != nil {
			e.duration = patch.Duration
			changed = true
		}
	}

	if changed {
		e.recalculateExerciseLoad(t)
		e.updatedAt = time.Now().UTC()
	}

	return nil
}

// validateForWorkoutExerciseType проверяет, что набор полей
// соответствует типу упражнения (ADR-025 п.9):
//
//	weight:   weight, sets, reps заданы; duration = nil
//	duration: duration задан;          weight, sets, reps = nil
func (e WorkoutExercise) validateForWorkoutExerciseType(t ExerciseType) error {
	switch t {
	case ExerciseTypeWeight:
		if e.weight == nil || e.sets == nil || e.reps == nil || e.duration != nil {
			return ErrInvalidWorkoutExerciseForType
		}
	case ExerciseTypeDuration:
		if e.duration == nil || e.weight != nil || e.sets != nil || e.reps != nil {
			return ErrInvalidWorkoutExerciseForType
		}
	default:
		return ErrInvalidExerciseType
	}

	return nil
}

// recalculateExerciseLoad пересчитывает exercise_load по типу (ADR-025 п.2):
//
//	weight:   weight × reps × sets
//	duration: duration
//
// Вызывается только после validateForWorkoutExerciseType —
// указатели гарантированно не nil, типы не противоречат полям.
func (e *WorkoutExercise) recalculateExerciseLoad(t ExerciseType) {
	switch t {
	case ExerciseTypeWeight:
		e.exerciseLoad = e.weight.Int() * e.reps.Int() * e.sets.Int()
	case ExerciseTypeDuration:
		e.exerciseLoad = e.duration.Int()
	}
}
