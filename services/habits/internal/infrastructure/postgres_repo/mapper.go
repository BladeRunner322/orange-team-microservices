package postgres_repo

import "github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"

// HabitModelToDomain конвертирует модель БД в доменную сущность.
// Валидирует поля через VO — если в БД мусор, вернёт ошибку.
func HabitModelToDomain(m HabitModel) (domain.Habit, error) {
	// TODO: NewName, NewDescription, RestoreHabit
	return domain.Habit{}, nil
}

// HabitDomainToModel конвертирует доменную сущность в модель БД.
func HabitDomainToModel(h domain.Habit) HabitModel {
	// TODO
	return HabitModel{}
}

// CompletionModelToDomain конвертирует модель БД в доменную сущность.
func CompletionModelToDomain(m CompletionModel) (domain.Completion, error) {
	// TODO
	return domain.Completion{}, nil
}