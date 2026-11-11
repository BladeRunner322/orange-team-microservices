package postgres_repo

import (
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
)

// WorkoutExerciseModelToDomain конвертирует модель БД в доменную сущность.
// Валидирует VO.
func WorkoutExerciseModelToDomain(m WorkoutExerciseModel) (domain.WorkoutExercise, error) {
	difficulty, err := domain.NewDifficulty(m.Difficulty)
	if err != nil {
		return domain.WorkoutExercise{}, fmt.Errorf("parse difficulty: %w", err)
	}

	var weight *domain.Weight
	if m.Weight != nil {
		w, err := domain.NewWeight(*m.Weight)
		if err != nil {
			return domain.WorkoutExercise{}, fmt.Errorf("parse weight: %w", err)
		}
		weight = &w
	}

	var sets *domain.Sets
	if m.Sets != nil {
		s, err := domain.NewSets(*m.Sets)
		if err != nil {
			return domain.WorkoutExercise{}, fmt.Errorf("parse sets: %w", err)
		}
		sets = &s
	}

	var reps *domain.Reps
	if m.Reps != nil {
		r, err := domain.NewReps(*m.Reps)
		if err != nil {
			return domain.WorkoutExercise{}, fmt.Errorf("parse reps: %w", err)
		}
		reps = &r
	}

	var duration *domain.Duration
	if m.Duration != nil {
		d, err := domain.NewDuration(*m.Duration)
		if err != nil {
			return domain.WorkoutExercise{}, fmt.Errorf("parse duration: %w", err)
		}
		duration = &d
	}

	return domain.RestoreWorkoutExercise(
		m.ID,
		m.Version,
		m.WorkoutID,
		m.ExerciseID,
		difficulty,
		weight,
		sets,
		reps,
		duration,
		m.Completed,
		m.ExerciseLoad,
		m.CreatedAt,
		m.UpdatedAt,
	), nil
}

// WorkoutExerciseDomainToModel конвертирует доменную сущность в модель БД.
func WorkoutExerciseDomainToModel(e domain.WorkoutExercise) WorkoutExerciseModel {
	var weight, sets, reps, duration *int

	if w := e.Weight(); w != nil {
		val := w.Int()
		weight = &val
	}
	if s := e.Sets(); s != nil {
		val := s.Int()
		sets = &val
	}
	if r := e.Reps(); r != nil {
		val := r.Int()
		reps = &val
	}
	if d := e.Duration(); d != nil {
		val := d.Int()
		duration = &val
	}

	return WorkoutExerciseModel{
		ID:           e.ID(),
		Version:      e.Version(),
		WorkoutID:    e.WorkoutID(),
		ExerciseID:   e.ExerciseID(),
		Difficulty:   e.Difficulty().Int(),
		Weight:       weight,
		Sets:         sets,
		Reps:         reps,
		Duration:     duration,
		Completed:    e.Completed(),
		ExerciseLoad: e.ExerciseLoad(),
		CreatedAt:    e.CreatedAt(),
		UpdatedAt:    e.UpdatedAt(),
	}
}
