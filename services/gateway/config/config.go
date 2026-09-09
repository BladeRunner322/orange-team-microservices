package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	HTTPPort     string        `envconfig:"GATEWAY_HTTP_PORT" default:":8080"`
	AuthGRPCAddr string        `envconfig:"AUTH_GRPC_ADDR" required:"true"`
	Timeout      time.Duration `envconfig:"GATEWAY_TIMEOUT" default:"10s"`
	EnableTLS    bool          `envconfig:"ENABLE_TLS" default:"false"`
	TLSCertFile  string        `envconfig:"TLS_CERT_FILE" default:""`
	TLSKeyFile   string        `envconfig:"TLS_KEY_FILE" default:""`
}

func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("process gateway config: %w", err)
	}
	return cfg, nil
}

func MustLoad() Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Errorf("load gateway config: %w", err))
	}
	return cfg
}
