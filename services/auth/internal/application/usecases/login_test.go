package usecases

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

func TestLogin_Execute(t *testing.T) {
	log := logger.NewTestLogger()
	tokenManager := mockTokenManager{}

	t.Run("успешный вход", func(t *testing.T) {
		repo := newMockRepository()
		email, _ := domain.NewEmail("test@example.com")
		fullName, _ := domain.NewFullName("Test User")
		password := "password123"
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		passHash, _ := domain.NewPasswordHash(string(hash))
		user := domain.NewUser(email, passHash, fullName)
		_ = repo.Save(context.Background(), user)

		uc := NewLogin(repo, tokenManager, log)
		token, err := uc.Execute(context.Background(), "test@example.com", "password123")

		require.NoError(t, err)
		assert.Equal(t, "test-token", token)
	})

	t.Run("неверные учётные данные — пользователь не найден", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewLogin(repo, tokenManager, log)

		_, err := uc.Execute(context.Background(), "unknown@example.com", "password123")

		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("неверный пароль", func(t *testing.T) {
		repo := newMockRepository()
		email, _ := domain.NewEmail("test@example.com")
		fullName, _ := domain.NewFullName("Test User")
		hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		passHash, _ := domain.NewPasswordHash(string(hash))
		user := domain.NewUser(email, passHash, fullName)
		_ = repo.Save(context.Background(), user)

		uc := NewLogin(repo, tokenManager, log)
		_, err := uc.Execute(context.Background(), "test@example.com", "wrongpass")

		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})
}
