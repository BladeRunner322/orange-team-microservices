package usecases

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

type Register struct {
	repo   ports.Repository
	logger *logger.Logger
}

func NewRegister(repo ports.Repository, log *logger.Logger) *Register {
	return &Register{repo: repo, logger: log}
}

func (uc *Register) Execute(ctx context.Context, emailStr, password, fullNameStr string) (domain.User, error) {
	log := uc.logger.With("email", emailStr, "operation", "Register")

	email, err := domain.NewEmail(emailStr)
	if err != nil {
		log.Warn("invalid email", "error", err)
		return domain.User{}, err
	}
	fullName, err := domain.NewFullName(fullNameStr)
	if err != nil {
		log.Warn("invalid full name", "error", err)
		return domain.User{}, err
	}
	if _, err := domain.NewPassword(password); err != nil {
		log.Warn("weak password", "error", err)
		return domain.User{}, err
	}

	if _, err := uc.repo.FindByEmail(ctx, email); err == nil {
		log.Warn("email already exists")
		return domain.User{}, domain.ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", "error", err)
		return domain.User{}, err
	}
	passwordHash, err := domain.NewPasswordHash(string(hash))
	if err != nil {
		log.Error("invalid password hash generated", "error", err)
		return domain.User{}, err
	}

	user := domain.NewUser(email, passwordHash, fullName)
	if err := uc.repo.Save(ctx, user); err != nil {
		log.Error("failed to save user", "error", err)
		return domain.User{}, err
	}

	log.Info("user registered successfully", "user_id", user.ID().String())
	return user, nil
}
