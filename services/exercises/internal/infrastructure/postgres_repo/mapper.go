package postgres_repo

import "github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/domain"

// ModelToDomain конвертирует модель БД в доменную сущность.
// Валидирует поля через VO — если в БД лежит мусор, вернёт ошибку.
func ModelToDomain(m ExerciseModel) (domain.Exercise, error) {
	name, err := domain.NewName(m.Name)
	if err != nil {
		return domain.Exercise{}, err
	}

	description, err := domain.NewDescription(m.Description)
	if err != nil {
		return domain.Exercise{}, err
	}

	difficulty, err := domain.NewDifficulty(m.Difficulty)
	if err != nil {
		return domain.Exercise{}, err
	}

	exerciseType, err := domain.NewExerciseType(m.Type)
	if err != nil {
		return domain.Exercise{}, err
	}

	isDeleted := m.DeletedAt != nil

	return domain.RestoreExercise(
		m.ID,
		name,
		description,
		difficulty,
		exerciseType,
		isDeleted,
		m.CreatedAt,
		m.UpdatedAt,
	), nil
}

// DomainToModel конвертирует доменную сущность в модель БД.
// DeletedAt всегда nil — этим полем управляют отдельные операции
// (Create оставляет NULL, MarkDeleted ставит NOW()).
func DomainToModel(exercise domain.Exercise) ExerciseModel {
	return ExerciseModel{
		ID:          exercise.ID(),
		Name:        exercise.Name().String(),
		Description: exercise.Description().String(),
		Difficulty:  exercise.Difficulty().Int(),
		Type:        exercise.ExerciseType().String(),
		DeletedAt:   nil,
		CreatedAt:   exercise.CreatedAt(),
		UpdatedAt:   exercise.UpdatedAt(),
	}
}
