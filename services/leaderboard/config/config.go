// Package config загружает конфигурацию Leaderboard из переменных окружения.
package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GRPCPort string `envconfig:"GRPC_PORT" default:":50056"`
	HTTPPort string `envconfig:"HTTP_PORT" default:":8080"`

	EnableReflection bool   `envconfig:"ENABLE_REFLECTION" default:"false"`
	EnableTLS        bool   `envconfig:"ENABLE_TLS" default:"false"`
	TLSCertFile      string `envconfig:"TLS_CERT_FILE" default:""`
	TLSKeyFile       string `envconfig:"TLS_KEY_FILE" default:""`
}

func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("process service config: %w", err)
	}
	return cfg, nil
}

func MustLoad() Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Errorf("get service config: %w", err))
	}
	return cfg
}
