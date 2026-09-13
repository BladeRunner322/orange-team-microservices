package domain

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
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

// RefreshToken — value object для refresh-токена.
type RefreshToken string

const (
	// refreshTokenLength — длина случайной строки (в байтах до base64).
	refreshTokenLength = 32
)

// NewRefreshToken валидирует существующий refresh-токен (пришедший от клиента).
func NewRefreshToken(raw string) (RefreshToken, error) {
	if raw == "" {
		return "", ErrInvalidRefreshToken
	}
	// базовая проверка длины (base64 от 32 байт ≈ 44 символа)
	if len(raw) < 40 {
		return "", ErrInvalidRefreshToken
	}
	return RefreshToken(raw), nil
}

// GenerateRefreshToken создаёт новый случайный refresh-токен.
func GenerateRefreshToken() (RefreshToken, error) {
	b := make([]byte, refreshTokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return RefreshToken(base64.RawURLEncoding.EncodeToString(b)), nil
}

func (t RefreshToken) String() string {
	return string(t)
}
