package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
)

// mockRepository — реализация ports.Repository для тестов.
// Хранит профили в map, позволяет эмулировать ошибки.
type mockRepository struct {
	profiles map[uuid.UUID]domain.Profile
	err      error // если задана — все методы вернут её
}

func newMockRepository() *mockRepository {
	return &mockRepository{profiles: make(map[uuid.UUID]domain.Profile)}
}

func (m *mockRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (domain.Profile, error) {
	if m.err != nil {
		return domain.Profile{}, m.err
	}

	profile, ok := m.profiles[userID]
	if !ok {
		return domain.Profile{}, domain.ErrProfileNotFound
	}

	return profile, nil
}

func (m *mockRepository) Upsert(ctx context.Context, profile domain.Profile) error {
	if m.err != nil {
		return m.err
	}

	// идемпотентно: если уже есть — не трогаем
	if _, ok := m.profiles[profile.UserID()]; !ok {
		m.profiles[profile.UserID()] = profile
	}

	return nil
}

func (m *mockRepository) Update(ctx context.Context, profile domain.Profile) error {
	if m.err != nil {
		return m.err
	}

	if _, ok := m.profiles[profile.UserID()]; !ok {
		return domain.ErrProfileNotFound
	}

	m.profiles[profile.UserID()] = profile

	return nil
}

func (m *mockRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	if m.err != nil {
		return m.err
	}

	if _, ok := m.profiles[userID]; !ok {
		return domain.ErrProfileNotFound
	}

	delete(m.profiles, userID)

	return nil
}
