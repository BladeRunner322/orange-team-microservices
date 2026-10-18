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
	// TODO: валидация типа-специфичных полей через ValidateForWorkoutExerciseType
	// TODO: расчёт exercise_load:
	//   - weight: weight × reps × sets
	//   - duration: duration
	now := time.Now().UTC()
	return WorkoutExercise{
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
	}, nil
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

// ApplyPatch применяет патч и пересчитывает exercise_load.
func (e *WorkoutExercise) ApplyPatch(patch WorkoutExercisePatch, exerciseType ExerciseType) error {
	// TODO: применить поля, пересчитать exercise_load
	return nil
}

// ValidateForWorkoutExerciseType проверяет, что поля соответствуют типу упражнения.
func (e WorkoutExercise) ValidateForWorkoutExerciseType(t ExerciseType) error {
	// TODO: weight требует weight/sets/reps, duration требует duration
	return nil
}

// recalculateExerciseLoad пересчитывает exercise_load по типу.
func (e *WorkoutExercise) recalculateExerciseLoad(t ExerciseType) {
	// TODO: weight → weight × reps × sets, duration → duration
}

// SetVersion нужен для optimistic locking.
func (e *WorkoutExercise) SetVersion(v int) { e.version = v }
