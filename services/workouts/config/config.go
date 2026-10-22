// Package config загружает конфигурацию Workouts-сервиса из переменных окружения.
package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GRPCPort string `envconfig:"GRPC_PORT" default:":50055"`
	HTTPPort string `envconfig:"HTTP_PORT" default:":8080"`

	ProfilesGRPCAddr  string        `envconfig:"PROFILES_GRPC_ADDR" required:"true"`
	ExercisesGRPCAddr string        `envconfig:"EXERCISES_GRPC_ADDR" required:"true"`
	GRPCClientTLSMode string        `envconfig:"GRPC_CLIENT_TLS_MODE" default:"insecure"`
	Timeout           time.Duration `envconfig:"WORKOUTS_TIMEOUT" default:"10s"`

	EnableReflection bool   `envconfig:"ENABLE_REFLECTION" default:"false"`
	EnableTLS        bool   `envconfig:"ENABLE_TLS" default:"false"`
	TLSCertFile      string `envconfig:"TLS_CERT_FILE" default:""`
	TLSKeyFile       string `envconfig:"TLS_KEY_FILE" default:""`
}

func Load() (Config, error) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		return Config{}, fmt.Errorf("process service config: %w", err)
	}
	return config, nil
}

func MustLoad() Config {
	config, err := Load()
	if err != nil {
		err = fmt.Errorf("get service config %w", err)
		panic(err)
	}
	return config
}
