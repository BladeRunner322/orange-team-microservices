package postgres

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string        `envconfig:"HOST" required:"true"`
	Port     string        `envconfig:"PORT" default:"5432"`
	User     string        `envconfig:"USER" required:"true"`
	Password string        `envconfig:"PASSWORD" required:"true"`
	Database string        `envconfig:"DB" required:"true"`
	Timeout  time.Duration `envconfig:"TIMEOUT" required:"true"`
}

// Load загружает конфиг из переменных окружения (префикс POSTGRES_).
func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("POSTGRES", &cfg); err != nil {
		return Config{}, fmt.Errorf("process env config: %w", err)
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
