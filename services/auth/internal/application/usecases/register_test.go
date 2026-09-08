package usecases

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

func TestRegister_Execute(t *testing.T) {
	log := logger.NewTestLogger()

	t.Run("успешная регистрация", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewRegister(repo, log)

		user, err := uc.Execute(context.Background(), "test@example.com", "password123", "Test User")

		require.NoError(t, err)
		assert.Equal(t, "test@example.com", user.Email().String())
		assert.Equal(t, "Test User", user.FullName().String())

		saved, err := repo.FindByEmail(context.Background(), user.Email())
		require.NoError(t, err)
		assert.Equal(t, user.ID(), saved.ID())
	})

	t.Run("email уже занят", func(t *testing.T) {
		repo := newMockRepository()
		email, _ := domain.NewEmail("existing@example.com")
		fullName, _ := domain.NewFullName("Existing")
		passHash, _ := domain.NewPasswordHash("hash")
		existingUser := domain.NewUser(email, passHash, fullName)
		_ = repo.Save(context.Background(), existingUser)

		uc := NewRegister(repo, log)
		_, err := uc.Execute(context.Background(), "existing@example.com", "password123", "Test User")

		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
	})

	t.Run("невалидный email", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewRegister(repo, log)

		_, err := uc.Execute(context.Background(), "invalid", "password123", "Test User")

		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidEmail)
	})

	t.Run("слабый пароль", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewRegister(repo, log)

		_, err := uc.Execute(context.Background(), "test@example.com", "123", "Test User")

		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWeakPassword)
	})
}
