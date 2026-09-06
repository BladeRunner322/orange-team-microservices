package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GRPCPort      string        `envconfig:"GRPC_PORT" default:":50051"`
	JWTSecret     string        `envconfig:"JWT_SECRET" required:"true"`
	JWTIssuer     string        `envconfig:"JWT_ISSUER" default:"auth-service"`
	JWTAudience   string        `envconfig:"JWT_AUDIENCE" default:"orange-team"`
	JWTExpiration time.Duration `envconfig:"JWT_EXPIRATION" default:"24h"`
}

func Load() (Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	return cfg, err
}

func MustLoad() Config {
	cfg, err := Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	return cfg
}
