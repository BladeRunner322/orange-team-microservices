package postgres_repo

import (
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
)

// WorkoutModelToDomain конвертирует модель БД в доменную сущность.
// Валидирует VO — если в БД лежит мусор, вернёт ошибку.
func WorkoutModelToDomain(m WorkoutModel) (domain.Workout, error) {
	status, err := domain.NewStatus(m.Status)
	if err != nil {
		return domain.Workout{}, fmt.Errorf("parse status: %w", err)
	}

	var intensity *domain.Intensity
	if m.Intensity != nil {
		i, err := domain.NewIntensity(*m.Intensity)
		if err != nil {
			return domain.Workout{}, fmt.Errorf("parse intensity: %w", err)
		}
		intensity = &i
	}

	isDeleted := m.DeletedAt != nil

	return domain.RestoreWorkout(
		m.ID,
		m.Version,
		m.UserID,
		status,
		m.StartedAt,
		m.CompletedAt,
		m.WorkoutScore,
		intensity,
		m.PersonalScoreCoefficient,
		isDeleted,
		m.CreatedAt,
		m.UpdatedAt,
	), nil
}

// WorkoutDomainToModel конвертирует доменную сущность в модель БД.
// DeletedAt всегда nil — этим полем управляет только MarkDeleted.
func WorkoutDomainToModel(w domain.Workout) WorkoutModel {
	var intensity *int
	if i := w.Intensity(); i != nil {
		val := i.Int()
		intensity = &val
	}

	return WorkoutModel{
		ID:                       w.ID(),
		Version:                  w.Version(),
		UserID:                   w.UserID(),
		Status:                   w.Status().String(),
		StartedAt:                w.StartedAt(),
		CompletedAt:              w.CompletedAt(),
		WorkoutScore:             w.WorkoutScore(),
		Intensity:                intensity,
		PersonalScoreCoefficient: w.PersonalScoreCoefficient(),
		DeletedAt:                nil,
		CreatedAt:                w.CreatedAt(),
		UpdatedAt:                w.UpdatedAt(),
	}
}
