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

	// ============================================================
	//  Circuit Breaker (ADR-020). Один набор параметров
	//  применяется ко всем downstream-клиентам (Profiles, Exercises).
	// ============================================================

	// CBEnabled включает Circuit Breaker на всех gRPC-клиентах Workouts.
	CBEnabled bool `envconfig:"WORKOUTS_CB_ENABLED" default:"true"`

	// CBMaxRequests — сколько запросов пропустить в half-open.
	CBMaxRequests uint32 `envconfig:"WORKOUTS_CB_MAX_REQUESTS" default:"1"`

	// CBInterval — окно подсчёта ошибок.
	CBInterval time.Duration `envconfig:"WORKOUTS_CB_INTERVAL" default:"60s"`

	// CBTimeout — сколько breaker висит в OPEN до half-open.
	CBTimeout time.Duration `envconfig:"WORKOUTS_CB_TIMEOUT" default:"30s"`

	// CBMinRequests — минимум запросов в окне для расчёта ErrorRate.
	CBMinRequests uint32 `envconfig:"WORKOUTS_CB_MIN_REQUESTS" default:"10"`

	// CBErrorRate — доля ошибок (0..1), при превышении — OPEN.
	CBErrorRate float64 `envconfig:"WORKOUTS_CB_ERROR_RATE" default:"0.5"`

	// ============================================================
	//  Retry (ADR-020).
	// ============================================================

	// RetryEnabled включает retry для методов из whitelist.
	RetryEnabled bool `envconfig:"WORKOUTS_RETRY_ENABLED" default:"true"`

	// RetryMaxAttempts — максимум попыток включая первую.
	RetryMaxAttempts uint32 `envconfig:"WORKOUTS_RETRY_MAX_ATTEMPTS" default:"3"`

	// RetryBaseDelay — начальная задержка exponential backoff.
	RetryBaseDelay time.Duration `envconfig:"WORKOUTS_RETRY_BASE_DELAY" default:"100ms"`

	// RetryMaxDelay — максимальная задержка между попытками.
	RetryMaxDelay time.Duration `envconfig:"WORKOUTS_RETRY_MAX_DELAY" default:"1s"`

	// RetryIdempotentMethods — whitelist read-only методов через запятую.
	// Формат: "/package.Service/Method". Применяется к обоим клиентам
	// (Profiles и Exercises); метод, которого нет в списке, не ретраится.
	RetryIdempotentMethods []string `envconfig:"WORKOUTS_RETRY_IDEMPOTENT_METHODS"`

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
