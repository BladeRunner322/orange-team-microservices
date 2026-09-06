package usecases

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

type RegisterUseCase struct {
	repo ports.Repository
}

func NewRegisterUseCase(repo ports.Repository) *RegisterUseCase {
	return &RegisterUseCase{repo: repo}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, emailStr, passwordStr, fullNameStr string) (domain.User, error) {
	// Валидация email, fullName, password
	email, err := domain.NewEmail(emailStr)
	if err != nil {
		return domain.User{}, err
	}
	fullName, err := domain.NewFullName(fullNameStr)
	if err != nil {
		return domain.User{}, err
	}
	password, err := domain.NewPassword(passwordStr)
	if err != nil {
		return domain.User{}, err
	}

	if _, err := uc.repo.FindByEmail(ctx, email); err == nil {
		return domain.User{}, domain.ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password.String()), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}
	passwordHash, err := domain.NewPasswordHash(string(hash))
	if err != nil {
		return domain.User{}, err
	}

	user := domain.NewUser(email, passwordHash, fullName)
	if err := uc.repo.Save(ctx, user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}
