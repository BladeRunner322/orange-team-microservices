package postgres_repo

import (
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
)

func WorkoutModelToDomain(m WorkoutModel) (domain.Workout, error) {
	// TODO: валидировать VO (Status, Intensity), вернуть RestoreWorkout
	//       с isDeleted := (m.DeletedAt != nil)
	return domain.Workout{}, nil
}

func WorkoutDomainToModel(w domain.Workout) WorkoutModel {
	// TODO: маппинг domain → model. DeletedAt всегда nil —
	//       deleted_at управляется только через repo.MarkDeleted (SQL UPDATE),
	//       как в Exercises/Habits (см. HabitDomainToModel).
	return WorkoutModel{}
}
