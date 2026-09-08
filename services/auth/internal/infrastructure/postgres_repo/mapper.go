package postgres_repo

import (
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

// UserModelToDomain преобразует модель БД в доменную сущность.
func UserModelToDomain(m UserModel) (domain.User, error) {
	email, err := domain.NewEmail(m.Email)
	if err != nil {
		return domain.User{}, err
	}
	passHash, err := domain.NewPasswordHash(m.PasswordHash)
	if err != nil {
		return domain.User{}, err
	}
	fullName, err := domain.NewFullName(m.FullName)
	if err != nil {
		return domain.User{}, err
	}
	// Используем специальный конструктор для восстановления
	return domain.RestoreUser(m.ID, email, passHash, fullName, m.CreatedAt, m.UpdatedAt), nil
}

// DomainToUserModel преобразует доменную сущность в модель БД.
func DomainToUserModel(user domain.User) UserModel {
	return UserModel{
		ID:           user.ID(),
		Email:        user.Email().String(),
		PasswordHash: user.PasswordHash().String(),
		FullName:     user.FullName().String(),
		CreatedAt:    user.CreatedAt(),
		UpdatedAt:    user.UpdatedAt(),
	}
}
