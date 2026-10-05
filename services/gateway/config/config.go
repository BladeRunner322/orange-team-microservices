package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// RateLimitRule — параметры Token Bucket для одного эндпоинта.
type RateLimitRule struct {
	Rate     int
	Burst    int
	Interval time.Duration
}

// RateLimitConfig — настройки rate limiting для всех эндпоинтов.
type RateLimitConfig struct {
	Enabled  bool
	Login    RateLimitRule
	Register RateLimitRule
	Refresh  RateLimitRule
	Default  RateLimitRule
}

type Config struct {
	HTTPPort          string        `envconfig:"GATEWAY_HTTP_PORT" default:":8081"`
	AuthGRPCAddr      string        `envconfig:"AUTH_GRPC_ADDR" required:"true"`
	ProfilesGRPCAddr  string        `envconfig:"PROFILES_GRPC_ADDR" required:"true"`
	ExercisesGRPCAddr string        `envconfig:"EXERCISES_GRPC_ADDR" required:"true"`
	Timeout           time.Duration `envconfig:"GATEWAY_TIMEOUT" default:"10s"`

	// GRPCClientTLSMode — режим TLS для исходящих gRPC-соединений
	// к Auth, Profiles, Exercises.
	// Значения: "disabled" | "insecure" | "verify".
	GRPCClientTLSMode string `envconfig:"GRPC_CLIENT_TLS_MODE" default:"insecure"`

	// TrustedProxies — список CIDR доверенных прокси.
	// Если RemoteAddr входит в этот список — доверяем заголовку
	// X-Forwarded-For. Если пусто — XFF игнорируется полностью.
	// envconfig парсит значение как slice через запятую:
	//   TRUSTED_PROXIES=10.0.0.0/8,172.16.0.0/12
	TrustedProxies []string `envconfig:"TRUSTED_PROXIES"`

	// Redis для rate limiting
	RedisAddr     string `envconfig:"REDIS_ADDR" required:"true"`
	RedisPassword string `envconfig:"REDIS_PASSWORD" default:""`
	RedisDB       int    `envconfig:"REDIS_DB" default:"0"`

	// Rate limiting — плоские поля для envconfig
	RateLimitEnabled bool `envconfig:"RATE_LIMIT_ENABLED" default:"true"`

	RateLimitLoginRate     int           `envconfig:"RATE_LIMIT_LOGIN_RATE" default:"5"`
	RateLimitLoginBurst    int           `envconfig:"RATE_LIMIT_LOGIN_BURST" default:"5"`
	RateLimitLoginInterval time.Duration `envconfig:"RATE_LIMIT_LOGIN_INTERVAL" default:"1m"`

	RateLimitRegisterRate     int           `envconfig:"RATE_LIMIT_REGISTER_RATE" default:"3"`
	RateLimitRegisterBurst    int           `envconfig:"RATE_LIMIT_REGISTER_BURST" default:"3"`
	RateLimitRegisterInterval time.Duration `envconfig:"RATE_LIMIT_REGISTER_INTERVAL" default:"1m"`

	RateLimitRefreshRate     int           `envconfig:"RATE_LIMIT_REFRESH_RATE" default:"20"`
	RateLimitRefreshBurst    int           `envconfig:"RATE_LIMIT_REFRESH_BURST" default:"20"`
	RateLimitRefreshInterval time.Duration `envconfig:"RATE_LIMIT_REFRESH_INTERVAL" default:"1m"`

	RateLimitDefaultRate     int           `envconfig:"RATE_LIMIT_DEFAULT_RATE" default:"60"`
	RateLimitDefaultBurst    int           `envconfig:"RATE_LIMIT_DEFAULT_BURST" default:"60"`
	RateLimitDefaultInterval time.Duration `envconfig:"RATE_LIMIT_DEFAULT_INTERVAL" default:"1m"`

	// Собранная структура — заполняется в Load(), не через envconfig.
	RateLimit RateLimitConfig `envconfig:"-"`
}

func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("process gateway config: %w", err)
	}

	// Собираем RateLimitConfig из плоских полей.
	cfg.RateLimit = RateLimitConfig{
		Enabled: cfg.RateLimitEnabled,
		Login: RateLimitRule{
			Rate:     cfg.RateLimitLoginRate,
			Burst:    cfg.RateLimitLoginBurst,
			Interval: cfg.RateLimitLoginInterval,
		},
		Register: RateLimitRule{
			Rate:     cfg.RateLimitRegisterRate,
			Burst:    cfg.RateLimitRegisterBurst,
			Interval: cfg.RateLimitRegisterInterval,
		},
		Refresh: RateLimitRule{
			Rate:     cfg.RateLimitRefreshRate,
			Burst:    cfg.RateLimitRefreshBurst,
			Interval: cfg.RateLimitRefreshInterval,
		},
		Default: RateLimitRule{
			Rate:     cfg.RateLimitDefaultRate,
			Burst:    cfg.RateLimitDefaultBurst,
			Interval: cfg.RateLimitDefaultInterval,
		},
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
