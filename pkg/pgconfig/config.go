// Package pgconfig — конфигурация подключения к PostgreSQL.
//
// Отделён от pkg/postgres, чтобы конфиг тестировался unit-тестами,
// а pkg/postgres оставался integration-only (pgx-специфика: пул, транзакции, адаптеры).
package pgconfig

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config — параметры подключения к PostgreSQL.
type Config struct {
	Host     string        `envconfig:"HOST"`
	Port     string        `envconfig:"PORT" default:"5432"`
	User     string        `envconfig:"USER"`
	Password string        `envconfig:"PASSWORD"`
	Database string        `envconfig:"DB"`
	Timeout  time.Duration `envconfig:"TIMEOUT"`
}

// Load загружает конфиг из переменных окружения (префикс POSTGRES_).
func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("POSTGRES", &cfg); err != nil {
		return Config{}, fmt.Errorf("process env config: %w", err)
	}

	if cfg.Host == "" {
		return Config{}, fmt.Errorf("POSTGRES_HOST is required")
	}
	if cfg.User == "" {
		return Config{}, fmt.Errorf("POSTGRES_USER is required")
	}
	if cfg.Password == "" {
		return Config{}, fmt.Errorf("POSTGRES_PASSWORD is required")
	}
	if cfg.Database == "" {
		return Config{}, fmt.Errorf("POSTGRES_DB is required")
	}
	if cfg.Timeout == 0 {
		return Config{}, fmt.Errorf("POSTGRES_TIMEOUT is required")
	}

	return cfg, nil
}

// MustLoad загружает конфиг или паникует.
func MustLoad() Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Errorf("load postgres config: %w", err))
	}
	
	return cfg
}