package usecases

import (
	"context"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	"github.com/google/uuid"
)

// mockHabitRepository — реализация ports.HabitRepository для тестов.
// Хранит привычки в map по id, эмулирует правила SQL:
//   - unique-имя среди активных у одного user (partial unique index);
//   - UPDATE/MarkDeleted не работают с удалёнными (WHERE deleted_at IS NULL).
type mockHabitRepository struct {
	habits map[uuid.UUID]domain.Habit

	// err — если задана, все методы вернут её вместо нормальной логики.
	err error
}

func newMockHabitRepository() *mockHabitRepository {
	return &mockHabitRepository{habits: make(map[uuid.UUID]domain.Habit)}
}

func (m *mockHabitRepository) WithTx(ctx context.Context, fn func(ports.HabitRepository) error) error {
	if m.err != nil {
		return m.err
	}
	// В памяти транзакции нет — прокидываем тот же мок.
	return fn(m)
}

func (m *mockHabitRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (domain.Habit, error) {
	return m.GetByID(ctx, id)
}

func (m *mockHabitRepository) Create(ctx context.Context, habit domain.Habit) (domain.Habit, error) {
	if m.err != nil {
		return domain.Habit{}, m.err
	}

	for _, existing := range m.habits {
		if existing.IsDeleted() {
			continue
		}
		if existing.UserID() == habit.UserID() && existing.Name() == habit.Name() {
			return domain.Habit{}, domain.ErrHabitNameExists
		}
	}

	m.habits[habit.ID()] = habit

	return habit, nil
}

func (m *mockHabitRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Habit, error) {
	if m.err != nil {
		return domain.Habit{}, m.err
	}

	habit, ok := m.habits[id]
	if !ok {
		return domain.Habit{}, domain.ErrHabitNotFound
	}

	return habit, nil
}

func (m *mockHabitRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Habit, error) {
	if m.err != nil {
		return nil, m.err
	}

	result := make([]domain.Habit, 0)
	for _, habit := range m.habits {
		if habit.UserID() == userID && !habit.IsDeleted() {
			result = append(result, habit)
		}
	}

	return result, nil
}

func (m *mockHabitRepository) Update(ctx context.Context, habit domain.Habit) (domain.Habit, error) {
	if m.err != nil {
		return domain.Habit{}, m.err
	}

	existing, ok := m.habits[habit.ID()]
	if !ok || existing.IsDeleted() {
		return domain.Habit{}, domain.ErrHabitNotFound
	}

	for id, other := range m.habits {
		if id == habit.ID() || other.IsDeleted() {
			continue
		}
		if other.UserID() == habit.UserID() && other.Name() == habit.Name() {
			return domain.Habit{}, domain.ErrHabitNameExists
		}
	}

	m.habits[habit.ID()] = habit

	return habit, nil
}

func (m *mockHabitRepository) MarkDeleted(ctx context.Context, id uuid.UUID) error {
	if m.err != nil {
		return m.err
	}

	existing, ok := m.habits[id]
	if !ok || existing.IsDeleted() {
		return domain.ErrHabitNotFound
	}

	deleted := domain.RestoreHabit(
		existing.ID(),
		existing.UserID(),
		existing.Name(),
		existing.Description(),
		true,
		existing.CreatedAt(),
		existing.UpdatedAt(),
	)
	m.habits[id] = deleted

	return nil
}

// mockCompletionRepository — реализация ports.CompletionRepository для тестов.
//
// Хранит сами domain.Completion в map, ключ — habitID + нормализованная дата.
// В отличие от хранения только createdAt, здесь не нужно парсить ключ обратно:
// доменная сущность уже содержит habitID и domain.Date.
type mockCompletionRepository struct {
	// ключ — habitID.String() + "|" + "YYYY-MM-DD"
	completions map[string]domain.Completion

	// err — если задана, все методы вернут её.
	err error
}

func newMockCompletionRepository() *mockCompletionRepository {
	return &mockCompletionRepository{completions: make(map[string]domain.Completion)}
}

func completionKey(habitID uuid.UUID, date time.Time) string {
	return habitID.String() + "|" + dayOnly(date).Format("2006-01-02")
}

func (m *mockCompletionRepository) Complete(ctx context.Context, completion domain.Completion) error {
	if m.err != nil {
		return m.err
	}

	key := completionKey(completion.HabitID(), completion.Date().Time())
	if _, ok := m.completions[key]; ok {
		return nil // ON CONFLICT DO NOTHING
	}
	m.completions[key] = completion

	return nil
}

func (m *mockCompletionRepository) Uncomplete(ctx context.Context, habitID uuid.UUID, date time.Time) error {
	if m.err != nil {
		return m.err
	}

	delete(m.completions, completionKey(habitID, date))

	return nil
}

func (m *mockCompletionRepository) IsCompleted(ctx context.Context, habitID uuid.UUID, date time.Time) (bool, error) {
	if m.err != nil {
		return false, m.err
	}

	_, ok := m.completions[completionKey(habitID, date)]

	return ok, nil
}

func (m *mockCompletionRepository) ListByHabit(ctx context.Context, habitID uuid.UUID, from, to time.Time) ([]domain.Completion, error) {
	if m.err != nil {
		return nil, m.err
	}

	fromDay := dayOnly(from)
	toDay := dayOnly(to)

	result := make([]domain.Completion, 0)
	for _, c := range m.completions {
		if c.HabitID() != habitID {
			continue
		}
		d := c.Date().Time()
		if d.Before(fromDay) || d.After(toDay) {
			continue
		}
		result = append(result, c)
	}

	// Сортировка по дате ASC (пузырёк — массивы в тестах маленькие).
	for i := 1; i < len(result); i++ {
		for j := i; j > 0 && result[j].Date().Time().Before(result[j-1].Date().Time()); j-- {
			result[j], result[j-1] = result[j-1], result[j]
		}
	}

	return result, nil
}

func (m *mockCompletionRepository) Streak(ctx context.Context, habitID uuid.UUID, today time.Time) (int, error) {
	if m.err != nil {
		return 0, m.err
	}

	todayDay := dayOnly(today)
	dates := make([]time.Time, 0)
	for _, c := range m.completions {
		if c.HabitID() != habitID {
			continue
		}
		d := c.Date().Time()
		if !d.After(todayDay) {
			dates = append(dates, d)
		}
	}

	if len(dates) == 0 {
		return 0, nil
	}

	// Сортировка DESC (пузырёк).
	for i := 1; i < len(dates); i++ {
		for j := i; j > 0 && dates[j].After(dates[j-1]); j-- {
			dates[j], dates[j-1] = dates[j-1], dates[j]
		}
	}

	yesterdayDay := todayDay.AddDate(0, 0, -1)
	firstDay := dates[0]

	var expected time.Time
	switch {
	case firstDay.Equal(todayDay):
		expected = todayDay
	case firstDay.Equal(yesterdayDay):
		expected = yesterdayDay
	default:
		return 0, nil
	}

	streak := 0
	for _, d := range dates {
		if !d.Equal(expected) {
			break
		}
		streak++
		expected = expected.AddDate(0, 0, -1)
	}

	return streak, nil
}
