package client

import (
	"crypto/tls"
	"fmt"
	"time"
)

const (
	// defaultTimeout — таймаут на каждый gRPC-вызов, если не задан в Config.
	defaultTimeout = 5 * time.Second

	// Дефолты Circuit Breaker (ADR-020). Применяются в applyDefaults(),
	// если CBEnabled=true и соответствующее поле оставлено нулевым.
	defaultCBMaxRequests uint32        = 1
	defaultCBInterval    time.Duration = 60 * time.Second
	defaultCBTimeout     time.Duration = 30 * time.Second
	defaultCBMinRequests uint32        = 10
	defaultCBErrorRate   float64       = 0.5

	// Дефолты Retry (ADR-020). Применяются в applyDefaults(),
	// если RetryEnabled=true и соответствующее поле оставлено нулевым.
	defaultRetryMaxAttempts uint32        = 3
	defaultRetryBaseDelay   time.Duration = 100 * time.Millisecond
	defaultRetryMaxDelay    time.Duration = 1 * time.Second
)

// TLSMode определяет, как клиент устанавливает соединение с сервером.
type TLSMode string

const (
	// TLSModeDisabled — без шифрования (только для локальной разработки).
	TLSModeDisabled TLSMode = "disabled"

	// TLSModeInsecure — TLS без проверки сертификата сервера
	// (для self-signed сертификатов в dev/staging).
	TLSModeInsecure TLSMode = "insecure"

	// TLSModeVerify — TLS с проверкой сертификата (для прода).
	TLSModeVerify TLSMode = "verify"
)

// Config — параметры подключения к gRPC-серверу.
type Config struct {
	// Target — адрес сервера, например "auth-service:50051".
	Target string

	// TLSMode — режим TLS.
	TLSMode TLSMode

	// Timeout — таймаут на каждый gRPC-вызов (per-call).
	// Если 0 — применяется 5 секунд по умолчанию.
	Timeout time.Duration

	// ============================================================
	//  Circuit Breaker (ADR-020).
	//  Если CBEnabled=false — остальные поля игнорируются.
	// ============================================================

	// CBEnabled включает Circuit Breaker на клиенте.
	// При открытом breaker вызовы падают мгновенно (fail-fast).
	CBEnabled bool

	// CBMaxRequests — сколько запросов пропустить в HALF-OPEN
	// подряд, прежде чем вернуться в CLOSED. Если 0 — defaultCBMaxRequests.
	CBMaxRequests uint32

	// CBInterval — окно подсчёта ошибок для CLOSED → OPEN.
	// Если 0 — defaultCBInterval.
	CBInterval time.Duration

	// CBTimeout — сколько breaker висит в OPEN до перехода в HALF-OPEN.
	// Если 0 — defaultCBTimeout.
	CBTimeout time.Duration

	// CBMinRequests — минимум запросов в окне CBInterval, необходимый
	// для расчёта ErrorRate. Защищает от открытия breaker на 2-3 ошибках.
	// Если 0 — defaultCBMinRequests.
	CBMinRequests uint32

	// CBErrorRate — доля ошибок (0..1), при превышении которой
	// breaker переходит CLOSED → OPEN. Если 0 — defaultCBErrorRate.
	CBErrorRate float64

	// ============================================================
	//  Retry (ADR-020).
	//  Если RetryEnabled=false — остальные поля игнорируются.
	// ============================================================

	// RetryEnabled включает retry на клиенте.
	RetryEnabled bool

	// RetryMaxAttempts — максимум попыток (включая первую).
	// Если 0 — defaultRetryMaxAttempts.
	RetryMaxAttempts uint32

	// RetryBaseDelay — начальная задержка exponential backoff.
	// Если 0 — defaultRetryBaseDelay.
	RetryBaseDelay time.Duration

	// RetryMaxDelay — максимальная задержка между попытками.
	// Если 0 — defaultRetryMaxDelay.
	RetryMaxDelay time.Duration

	// RetryMethods — whitelist полных имён методов, для которых
	// разрешён retry (например, "/auth.AuthService/ValidateToken").
	// Методы вне списка не ретраятся, даже если RetryEnabled=true.
	// См. ADR-020, п.3.
	RetryMethods []string
}

// applyDefaults применяет значения по умолчанию к нулевым полям.
//
// Для bool-полей дефолтов нет: включение CB/Retry — явное решение
// вызывающего. Для числовых — дефолт применяется только если
// соответствующая фича включена.
//
// Вызывается из client.New() до validate().
func (c *Config) applyDefaults() {
	if c.Timeout == 0 {
		c.Timeout = defaultTimeout
	}

	if c.CBEnabled {
		if c.CBMaxRequests == 0 {
			c.CBMaxRequests = defaultCBMaxRequests
		}
		if c.CBInterval == 0 {
			c.CBInterval = defaultCBInterval
		}
		if c.CBTimeout == 0 {
			c.CBTimeout = defaultCBTimeout
		}
		if c.CBMinRequests == 0 {
			c.CBMinRequests = defaultCBMinRequests
		}
		if c.CBErrorRate == 0 {
			c.CBErrorRate = defaultCBErrorRate
		}
	}

	if c.RetryEnabled {
		if c.RetryMaxAttempts == 0 {
			c.RetryMaxAttempts = defaultRetryMaxAttempts
		}
		if c.RetryBaseDelay == 0 {
			c.RetryBaseDelay = defaultRetryBaseDelay
		}
		if c.RetryMaxDelay == 0 {
			c.RetryMaxDelay = defaultRetryMaxDelay
		}
	}
}

func (c Config) validate() error {
	if c.Target == "" {
		return fmt.Errorf("grpc client target is empty")
	}

	if c.Timeout < 0 {
		return fmt.Errorf("grpc client timeout must be non-negative, got %s", c.Timeout)
	}

	switch c.TLSMode {
	case TLSModeDisabled, TLSModeInsecure, TLSModeVerify:
		// ok
	default:
		return fmt.Errorf("unknown tls mode %q", c.TLSMode)
	}

	// Валидация применяется только если соответствующая фича включена.
	// При CBEnabled=false поля CB игнорируются — это позволяет
	// передавать Config{} без CB-параметров.
	if c.CBEnabled {
		if err := c.validateCircuitBreaker(); err != nil {
			return err
		}
	}

	if c.RetryEnabled {
		if err := c.validateRetry(); err != nil {
			return err
		}
	}

	return nil
}

func (c Config) validateCircuitBreaker() error {
	if c.CBMaxRequests == 0 {
		return fmt.Errorf("circuit breaker max requests must be > 0")
	}
	if c.CBInterval <= 0 {
		return fmt.Errorf("circuit breaker interval must be > 0")
	}
	if c.CBTimeout <= 0 {
		return fmt.Errorf("circuit breaker timeout must be > 0")
	}
	if c.CBMinRequests == 0 {
		return fmt.Errorf("circuit breaker min requests must be > 0")
	}
	if c.CBErrorRate <= 0 || c.CBErrorRate > 1 {
		return fmt.Errorf("circuit breaker error rate must be in (0, 1], got %v", c.CBErrorRate)
	}
	return nil
}

func (c Config) validateRetry() error {
	if c.RetryMaxAttempts < 2 {
		return fmt.Errorf("retry max attempts must be >= 2 (1 initial + retries), got %d", c.RetryMaxAttempts)
	}
	if c.RetryBaseDelay <= 0 {
		return fmt.Errorf("retry base delay must be > 0")
	}
	if c.RetryMaxDelay <= 0 {
		return fmt.Errorf("retry max delay must be > 0")
	}
	if c.RetryMaxDelay < c.RetryBaseDelay {
		return fmt.Errorf("retry max delay (%s) must be >= base delay (%s)", c.RetryMaxDelay, c.RetryBaseDelay)
	}
	if len(c.RetryMethods) == 0 {
		return fmt.Errorf("retry enabled but no methods in whitelist")
	}
	return nil
}

// tlsConfig возвращает *tls.Config для данного режима.
// Для TLSModeDisabled возвращает nil.
func (c Config) tlsConfig() *tls.Config {
	switch c.TLSMode {
	case TLSModeVerify:
		return &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	case TLSModeInsecure:
		return &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true, //nolint:gosec // осознанно для self-signed
		}
	default:
		return nil
	}
}