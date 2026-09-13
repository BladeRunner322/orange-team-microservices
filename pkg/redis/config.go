package redis

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr     string `envconfig:"ADDR" required:"true"`
	Password string `envconfig:"PASSWORD" default:""`
	DB       int    `envconfig:"DB" default:"0"`
}

// Load загружает конфиг из переменных окружения с префиксом REDIS_.
func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("REDIS", &cfg); err != nil {
		return Config{}, fmt.Errorf("process redis config: %w", err)
	}
	return cfg, nil
}

// MustLoad загружает конфиг или паникует.
func MustLoad() Config {
	cfg, err := Load()
	if err != nil {
		err := fmt.Errorf("load redis config: %w", err)
		panic(err)
	}
	return cfg
}
