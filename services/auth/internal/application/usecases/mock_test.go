package usecases

import (
	"context"
	"errors"

	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
	"github.com/google/uuid"
)

// mockRepository реализует ports.Repository
type mockRepository struct {
	users map[string]domain.User // key = email
	err   error                  // для имитации ошибок
}

func newMockRepository() *mockRepository {
	return &mockRepository{users: make(map[string]domain.User)}
}

func (m *mockRepository) Save(ctx context.Context, user domain.User) error {
	if m.err != nil {
		return m.err
	}
	m.users[user.Email().String()] = user
	return nil
}

func (m *mockRepository) FindByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	if m.err != nil {
		return domain.User{}, m.err
	}
	user, ok := m.users[email.String()]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *mockRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	if m.err != nil {
		return domain.User{}, m.err
	}
	for _, u := range m.users {
		if u.ID() == id {
			return u, nil
		}
	}
	return domain.User{}, domain.ErrUserNotFound
}

// mockTokenManager реализует ports.TokenManager
type mockTokenManager struct {
	generateErr error
	validateErr error
}

func (m mockTokenManager) Generate(ctx context.Context, userID string) (string, error) {
	if m.generateErr != nil {
		return "", m.generateErr
	}
	return "test-token", nil
}

func (m mockTokenManager) Validate(ctx context.Context, token string) (string, error) {
	if m.validateErr != nil {
		return "", m.validateErr
	}
	if token == "" {
		return "", errors.New("empty token")
	}
	return "user-id", nil
}
