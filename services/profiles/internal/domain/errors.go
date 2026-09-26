// Package domain — доменные ошибки Profiles-сервиса.
package domain

import "errors"

var (
	ErrProfileNotFound  = errors.New("profile not found")
	ErrInvalidSex       = errors.New("invalid sex")
	ErrInvalidWeight    = errors.New("invalid weight")
	ErrInvalidHeight    = errors.New("invalid height")
	ErrInvalidBirthDate = errors.New("invalid birth date")
)
