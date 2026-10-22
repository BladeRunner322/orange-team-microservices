package domain

import (
	"fmt"
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
	isDeleted                bool
	createdAt                time.Time
	updatedAt                time.Time
}

// NewWorkout создаёт новую тренировку (status = planned).
func NewWorkout(userID uuid.UUID, personalScoreCoefficient int) Workout {
	now := time.Now().UTC()
	return Workout{
		id:                       uuid.New(),
		version:                  0,
		userID:                   userID,
		status:                   StatusPlanned,
		workoutScore:             0,
		personalScoreCoefficient: personalScoreCoefficient,
		isDeleted:                false,
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
	isDeleted bool,
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
		isDeleted:                isDeleted,
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
func (w Workout) IsDeleted() bool               { return w.isDeleted }
func (w Workout) CreatedAt() time.Time          { return w.createdAt }
func (w Workout) UpdatedAt() time.Time          { return w.updatedAt }

// CanModifyWorkoutExercise сообщает, можно ли менять упражнения в тренировке.
// Правило «нельзя в терминальном статусе» — ADR-025 п.9. Терминальность
// определяет VO Status, здесь только бизнес-правило, которое её использует.
func (w Workout) CanModifyWorkoutExercise() bool {
	return !w.status.IsTerminal()
}

// ApplyPatch применяет патч. Включает state machine и инварианты started_at / completed_at.
func (w *Workout) ApplyPatch(patch WorkoutPatch) error {
	// 1. Валидация — до любых мутаций.
	if patch.Status != nil {
		next := *patch.Status
		if next != w.status && !w.canTransitionTo(next) {
			return fmt.Errorf("%w: %s → %s", ErrInvalidStatusTransition, w.status, next)
		}
	}

	// 2. Применение полей — независимы друг от друга.
	changed := false

	if patch.Intensity.Set {
		w.intensity = patch.Intensity.Value
		changed = true
	}

	if patch.Status != nil && *patch.Status != w.status {
		w.applyStatusTransition(*patch.Status)
		changed = true
	}

	// 3. updated_at — только если реально что-то поменяли.
	if changed {
		w.updatedAt = time.Now().UTC()
	}

	return nil
}

// applyStatusTransition применяет переход к newStatus и выставляет
// инварианты started_at / completed_at.
//
// Валидация перехода — снаружи (canTransitionTo). Метод вызывается
// только для валидного перехода, newStatus != w.status.
func (w *Workout) applyStatusTransition(newStatus Status) {
	now := time.Now().UTC()

	switch newStatus {
	case StatusInProgress:
		if w.startedAt == nil {
			w.startedAt = &now
		}
	case StatusCompleted:
		if w.completedAt == nil {
			w.completedAt = &now
		}
	}

	w.status = newStatus
}

// canTransitionTo — таблица разрешённых переходов (ADR-025 п.9).
//
//	planned     → in_progress, cancelled
//	in_progress → completed, cancelled
//	completed   → запрещено
//	cancelled   → запрещено
//
// Self-transition (same → same) разрешён как idempotent no-op:
// повторный PATCH с тем же статусом не должен возвращать ошибку.
func (w Workout) canTransitionTo(next Status) bool {
	// Self-transition — идемпотентный no-op (REST PATCH).
	if w.status == next {
		return true
	}

	if w.status.IsTerminal() {
		return false
	}

	switch w.status {
	case StatusPlanned:
		return next == StatusInProgress || next == StatusCancelled
	case StatusInProgress:
		return next == StatusCompleted || next == StatusCancelled
	default:
		return false
	}
}

// SetWorkoutScore обновляет score (вызывается из recalculateScore в usecases).
func (w *Workout) SetWorkoutScore(score int) {
	w.workoutScore = score
	w.updatedAt = time.Now().UTC()
}
