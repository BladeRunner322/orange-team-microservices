package domain

import (
	"regexp"
	"strings"
)

// Email
type Email string

func NewEmail(raw string) (Email, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrInvalidEmail
	}
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !re.MatchString(trimmed) {
		return "", ErrInvalidEmail
	}
	return Email(strings.ToLower(trimmed)), nil
}

func (e Email) String() string { return string(e) }

// FullName
type FullName string

func NewFullName(raw string) (FullName, error) {
	trimmed := strings.TrimSpace(raw)
	if len(trimmed) < 2 || len(trimmed) > 100 {
		return "", ErrInvalidFullName
	}
	return FullName(trimmed), nil
}

func (f FullName) String() string { return string(f) }

// PasswordHash (существующий)
type PasswordHash string

func NewPasswordHash(hash string) (PasswordHash, error) {
	if hash == "" {
		return "", ErrInvalidPasswordHash
	}
	if len(hash) < 10 {
		return "", ErrInvalidPasswordHash
	}
	return PasswordHash(hash), nil
}

func (p PasswordHash) String() string { return string(p) }

// Password – value object для валидации пароля (не хранится в БД)
type Password string

func NewPassword(raw string) (Password, error) {
	if len(raw) < 8 {
		return "", ErrWeakPassword
	}
	// можно добавить проверки на сложность (цифры, заглавные, спецсимволы)
	return Password(raw), nil
}

func (p Password) String() string { return string(p) }
