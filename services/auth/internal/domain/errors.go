package domain

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidEmail        = errors.New("invalid email")
	ErrInvalidFullName     = errors.New("invalid full name")
	ErrInvalidPasswordHash = errors.New("invalid password hash")
	ErrWeakPassword        = errors.New("password must be at least 8 characters long")
	ErrInvalidToken        = errors.New("invalid token")
)
