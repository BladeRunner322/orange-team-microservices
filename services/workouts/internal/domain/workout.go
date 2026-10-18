package domain

import (
	"time"

	"github.com/google/uuid"
)

// Workout — доменная сущность тренировки.
type Workout struct {
	id                       uuid.UUID
	version                  int
	userID                   uuid.UUID
	status                   Status
	startedAt                *time.Time
	completedAt              *time.Time
	workoutScore             int
	intensity                *Intensity
	personalScoreCoefficient int
	createdAt                time.Time
	updatedAt                time.Time
}

// NewWorkout создаёт новую тренировку (status = planned).
func NewWorkout(userID uuid.UUID, personalScoreCoefficient int) Workout {
	// TODO: validate userID != uuid.Nil, coefficient in [MinScoreCoefficient, MaxScoreCoefficient]
	now := time.Now().UTC()
	return Workout{
		id:                       uuid.New(),
		version:                  0,
		userID:                   userID,
		status:                   StatusPlanned,
		workoutScore:             0,
		personalScoreCoefficient: personalScoreCoefficient,
		createdAt:                now,
		updatedAt:                now,
	}
}

// RestoreWorkout восстанавливает тренировку из БД.
func RestoreWorkout(
	id uuid.UUID,
	version int,
	userID uuid.UUID,
	status Status,
	startedAt *time.Time,
	completedAt *time.Time,
	workoutScore int,
	intensity *Intensity,
	personalScoreCoefficient int,
	createdAt time.Time,
	updatedAt time.Time,
) Workout {
	return Workout{
		id:                       id,
		version:                  version,
		userID:                   userID,
		status:                   status,
		startedAt:                startedAt,
		completedAt:              completedAt,
		workoutScore:             workoutScore,
		intensity:                intensity,
		personalScoreCoefficient: personalScoreCoefficient,
		createdAt:                createdAt,
		updatedAt:                updatedAt,
	}
}

// Геттеры.
func (w Workout) ID() uuid.UUID                 { return w.id }
func (w Workout) Version() int                  { return w.version }
func (w Workout) UserID() uuid.UUID             { return w.userID }
func (w Workout) Status() Status                { return w.status }
func (w Workout) StartedAt() *time.Time         { return w.startedAt }
func (w Workout) CompletedAt() *time.Time       { return w.completedAt }
func (w Workout) WorkoutScore() int             { return w.workoutScore }
func (w Workout) Intensity() *Intensity         { return w.intensity }
func (w Workout) PersonalScoreCoefficient() int { return w.personalScoreCoefficient }
func (w Workout) CreatedAt() time.Time          { return w.createdAt }
func (w Workout) UpdatedAt() time.Time          { return w.updatedAt }

// CanModifyWorkoutExercise сообщает, можно ли менять упражнения в тренировке.
// Возможно только в planned или in_progress (ADR-025 п.9).
func (w Workout) CanModifyWorkoutExercise() bool {
	// TODO: status != completed && status != cancelled
	return false
}

// ApplyPatch применяет патч. Включает state machine и инварианты started_at / completed_at.
func (w *Workout) ApplyPatch(patch WorkoutPatch) error {
	// TODO: state machine + инварианты:
	//  - planned → in_progress → completed, без прыжков
	//  - из completed / cancelled переходить нельзя
	//  - при in_progress проставить startedAt = NOW, если пусто
	//  - при completed проставить completedAt = NOW, если пусто
	return nil
}

// SetWorkoutScore обновляет score (вызывается из recalculateScore в usecases).
func (w *Workout) SetWorkoutScore(score int) {
	w.workoutScore = score
	w.updatedAt = time.Now().UTC()
}

// SetVersion нужен для optimistic locking (используется в маппере при восстановлении).
func (w *Workout) SetVersion(v int) { w.version = v }
