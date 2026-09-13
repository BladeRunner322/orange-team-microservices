package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GRPCPort string `envconfig:"GRPC_PORT" default:":50051"`
	HTTPPort string `envconfig:"HTTP_PORT" default:":8080"`

	JWTSecret   string `envconfig:"JWT_SECRET" required:"true"`
	JWTIssuer   string `envconfig:"JWT_ISSUER" default:"auth-service"`
	JWTAudience string `envconfig:"JWT_AUDIENCE" default:"orange-team"`

	AccessTokenTTL  time.Duration `envconfig:"ACCESS_TOKEN_TTL" default:"15m"`
	RefreshTokenTTL time.Duration `envconfig:"REFRESH_TOKEN_TTL" default:"720h"`

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
		err = fmt.Errorf("get Service config %w", err)
		panic(err)
	}
	return config
}
