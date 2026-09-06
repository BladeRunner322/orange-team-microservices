package usecases

import (
	"context"
	"testing"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

func TestRegister_Execute(t *testing.T) {
	log := logger.NewTestLogger() // нужно создать, см. ниже

	t.Run("успешная регистрация", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewRegister(repo, log)

		user, err := uc.Execute(context.Background(), "test@example.com", "password123", "Test User")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.Email.String() != "test@example.com" {
			t.Errorf("expected email test@example.com, got %s", user.Email.String())
		}
		// проверяем, что пользователь сохранён
		saved, _ := repo.FindByEmail(context.Background(), user.Email)
		if saved.ID != user.ID {
			t.Error("user was not saved correctly")
		}
	})

	t.Run("email уже занят", func(t *testing.T) {
		repo := newMockRepository()
		// предварительно сохраняем пользователя
		email, _ := domain.NewEmail("existing@example.com")
		fullName, _ := domain.NewFullName("Existing")
		passHash, _ := domain.NewPasswordHash("hash")
		existingUser := domain.NewUser(email, passHash, fullName)
		_ = repo.Save(context.Background(), existingUser)

		uc := NewRegister(repo, log)
		_, err := uc.Execute(context.Background(), "existing@example.com", "password123", "Test User")
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err != domain.ErrEmailAlreadyExists {
			t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
		}
	})

	t.Run("невалидный email", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewRegister(repo, log)
		_, err := uc.Execute(context.Background(), "invalid", "password123", "Test User")
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err != domain.ErrInvalidEmail {
			t.Errorf("expected ErrInvalidEmail, got %v", err)
		}
	})

	t.Run("слабый пароль", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewRegister(repo, log)
		_, err := uc.Execute(context.Background(), "test@example.com", "123", "Test User")
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err != domain.ErrWeakPassword {
			t.Errorf("expected ErrWeakPassword, got %v", err)
		}
	})
}
