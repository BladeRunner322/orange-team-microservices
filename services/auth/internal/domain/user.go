package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	id           uuid.UUID
	email        Email
	passwordHash PasswordHash
	fullName     FullName
	createdAt    time.Time
	updatedAt    *time.Time
}

// NewUser создаёт нового пользователя (генерирует ID и время создания).
func NewUser(email Email, passwordHash PasswordHash, fullName FullName) User {
	return User{
		id:           uuid.New(),
		email:        email,
		passwordHash: passwordHash,
		fullName:     fullName,
		createdAt:    time.Now().UTC(),
		updatedAt:    nil,
	}
}

// RestoreUser восстанавливает пользователя из БД (для маппинга).
func RestoreUser(
	id uuid.UUID,
	email Email,
	passwordHash PasswordHash,
	fullName FullName,
	createdAt time.Time,
	updatedAt *time.Time,
) User {
	return User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		fullName:     fullName,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// Геттеры (публичные)
func (u User) ID() uuid.UUID              { return u.id }
func (u User) Email() Email               { return u.email }
func (u User) PasswordHash() PasswordHash { return u.passwordHash }
func (u User) FullName() FullName         { return u.fullName }
func (u User) CreatedAt() time.Time       { return u.createdAt }
func (u User) UpdatedAt() *time.Time      { return u.updatedAt }
