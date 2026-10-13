// Package pgerrors — сентинелы и утилиты для маппинга ошибок PostgreSQL
// (pgx/pgconn) в единые ошибки уровня проекта.
//
// Пакет нейтрален к транспорту: используется репозиториями любых сервисов,
// работающих с PostgreSQL через pkg/postgres.
package pgerrors

import "errors"

var (
	ErrNoRows             = errors.New("no rows")
	ErrViolatesForeignKey = errors.New("violates foreign key")
	ErrViolatesUnique     = errors.New("violates unique constraint")
	ErrUnknown            = errors.New("unknown")
)