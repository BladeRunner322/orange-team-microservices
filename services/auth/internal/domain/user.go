package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        Email
	PasswordHash PasswordHash
	FullName     FullName
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

func NewUser(email Email, passwordHash PasswordHash, fullName FullName) User {
	return User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    nil,
	}
}
