package postgres_repo

import "github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"

// CompletionModelToDomain конвертирует модель БД в доменную сущность.
func CompletionModelToDomain(m CompletionModel) (domain.Completion, error) {
	date, err := domain.NewDate(m.Date)
	if err != nil {
		return domain.Completion{}, err
	}

	return domain.RestoreCompletion(
		m.HabitID,
		date,
		m.CreatedAt,
	), nil
}

// CompletionDomainToModel конвертирует доменную сущность в модель БД.
func CompletionDomainToModel(completion domain.Completion) CompletionModel {
	return CompletionModel{
		HabitID:   completion.HabitID(),
		Date:      completion.Date().Time(),
		CreatedAt: completion.CreatedAt(),
	}
}
