package postgres_repo

import "github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"

func WorkoutExerciseModelToDomain(m WorkoutExerciseModel) (domain.WorkoutExercise, error) {
	// TODO: валидировать VO (Weight, Sets, Reps, Duration, Difficulty), RestoreWorkoutExercise
	return domain.WorkoutExercise{}, nil
}

func WorkoutExerciseDomainToModel(e domain.WorkoutExercise) WorkoutExerciseModel {
	// TODO: маппинг
	return WorkoutExerciseModel{}
}
