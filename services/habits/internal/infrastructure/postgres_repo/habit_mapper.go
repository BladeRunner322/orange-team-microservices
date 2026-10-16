package postgres_repo

import "github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"

// HabitModelToDomain конвертирует модель БД в доменную сущность.
func HabitModelToDomain(m HabitModel) (domain.Habit, error) {
	name, err := domain.NewName(m.Name)
	if err != nil {
		return domain.Habit{}, err
	}

	description, err := domain.NewDescription(m.Description)
	if err != nil {
		return domain.Habit{}, err
	}

	isDeleted := m.DeletedAt != nil

	return domain.RestoreHabit(
		m.ID,
		m.UserID,
		name,
		description,
		isDeleted,
		m.CreatedAt,
		m.UpdatedAt,
	), nil
}

// HabitDomainToModel конвертирует доменную сущность в модель БД.
func HabitDomainToModel(habit domain.Habit) HabitModel {
	return HabitModel{
		ID:          habit.ID(),
		UserID:      habit.UserID(),
		Name:        habit.Name().String(),
		Description: habit.Description().String(),
		DeletedAt:   nil,
		CreatedAt:   habit.CreatedAt(),
		UpdatedAt:   habit.UpdatedAt(),
	}
}
