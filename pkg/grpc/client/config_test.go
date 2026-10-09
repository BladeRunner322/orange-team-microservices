package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyDefaults_Timeout(t *testing.T) {
	t.Run("нулевой timeout заменяется на defaultTimeout", func(t *testing.T) {
		cfg := Config{Target: "x", TLSMode: TLSModeInsecure}
		cfg.applyDefaults()
		assert.Equal(t, defaultTimeout, cfg.Timeout)
	})

	t.Run("заданный timeout не перезаписывается", func(t *testing.T) {
		cfg := Config{Target: "x", TLSMode: TLSModeInsecure, Timeout: 3 * time.Second}
		cfg.applyDefaults()
		assert.Equal(t, 3*time.Second, cfg.Timeout)
	})
}

func TestApplyDefaults_CBDisabled(t *testing.T) {
	t.Run("CB выключен — поля CB не трогаются", func(t *testing.T) {
		cfg := Config{Target: "x", TLSMode: TLSModeInsecure, CBEnabled: false}
		cfg.applyDefaults()
		assert.Zero(t, cfg.CBMaxRequests)
		assert.Zero(t, cfg.CBInterval)
		assert.Zero(t, cfg.CBTimeout)
		assert.Zero(t, cfg.CBMinRequests)
		assert.Zero(t, cfg.CBErrorRate)
	})
}

func TestApplyDefaults_CBEnabled(t *testing.T) {
	t.Run("CB включён — все поля заполнены дефолтами", func(t *testing.T) {
		cfg := Config{Target: "x", TLSMode: TLSModeInsecure, CBEnabled: true}
		cfg.applyDefaults()

		assert.Equal(t, defaultCBMaxRequests, cfg.CBMaxRequests)
		assert.Equal(t, defaultCBInterval, cfg.CBInterval)
		assert.Equal(t, defaultCBTimeout, cfg.CBTimeout)
		assert.Equal(t, defaultCBMinRequests, cfg.CBMinRequests)
		assert.Equal(t, defaultCBErrorRate, cfg.CBErrorRate)
	})

	t.Run("заданные поля не перезаписываются", func(t *testing.T) {
		cfg := Config{
			Target:        "x",
			TLSMode:       TLSModeInsecure,
			CBEnabled:     true,
			CBMaxRequests: 5,
			CBErrorRate:   0.8,
		}
		cfg.applyDefaults()

		assert.Equal(t, uint32(5), cfg.CBMaxRequests)
		assert.Equal(t, 0.8, cfg.CBErrorRate)
		// Остальные — дефолт.
		assert.Equal(t, defaultCBTimeout, cfg.CBTimeout)
	})
}

func TestApplyDefaults_Retry(t *testing.T) {
	t.Run("Retry выключен — поля не трогаются", func(t *testing.T) {
		cfg := Config{Target: "x", TLSMode: TLSModeInsecure, RetryEnabled: false}
		cfg.applyDefaults()
		assert.Zero(t, cfg.RetryMaxAttempts)
		assert.Zero(t, cfg.RetryBaseDelay)
		assert.Zero(t, cfg.RetryMaxDelay)
	})

	t.Run("Retry включён — дефолты заполнены", func(t *testing.T) {
		cfg := Config{Target: "x", TLSMode: TLSModeInsecure, RetryEnabled: true}
		cfg.applyDefaults()

		assert.Equal(t, defaultRetryMaxAttempts, cfg.RetryMaxAttempts)
		assert.Equal(t, defaultRetryBaseDelay, cfg.RetryBaseDelay)
		assert.Equal(t, defaultRetryMaxDelay, cfg.RetryMaxDelay)
	})
}

func TestValidate_Base(t *testing.T) {
	t.Run("пустой Target — ошибка", func(t *testing.T) {
		cfg := Config{TLSMode: TLSModeInsecure}
		assert.Error(t, cfg.validate())
	})

	t.Run("неизвестный TLSMode — ошибка", func(t *testing.T) {
		cfg := Config{Target: "x", TLSMode: "bogus"}
		assert.Error(t, cfg.validate())
	})

	t.Run("отрицательный Timeout — ошибка", func(t *testing.T) {
		cfg := Config{Target: "x", TLSMode: TLSModeInsecure, Timeout: -1 * time.Second}
		assert.Error(t, cfg.validate())
	})

	t.Run("валидный минимальный конфиг", func(t *testing.T) {
		cfg := Config{Target: "x", TLSMode: TLSModeInsecure, Timeout: time.Second}
		assert.NoError(t, cfg.validate())
	})
}

func TestValidate_CircuitBreaker(t *testing.T) {
	base := func() Config {
		return Config{
			Target:        "x",
			TLSMode:       TLSModeInsecure,
			Timeout:       time.Second,
			CBEnabled:     true,
			CBMaxRequests: 1,
			CBInterval:    30 * time.Second,
			CBTimeout:     10 * time.Second,
			CBMinRequests: 5,
			CBErrorRate:   0.5,
		}
	}

	t.Run("валидный CB", func(t *testing.T) {
		cfg := base()
		assert.NoError(t, cfg.validate())
	})

	t.Run("CBMaxRequests = 0 — ошибка", func(t *testing.T) {
		cfg := base()
		cfg.CBMaxRequests = 0
		assert.Error(t, cfg.validate())
	})

	t.Run("CBInterval = 0 — ошибка", func(t *testing.T) {
		cfg := base()
		cfg.CBInterval = 0
		assert.Error(t, cfg.validate())
	})

	t.Run("CBTimeout = 0 — ошибка", func(t *testing.T) {
		cfg := base()
		cfg.CBTimeout = 0
		assert.Error(t, cfg.validate())
	})

	t.Run("CBMinRequests = 0 — ошибка", func(t *testing.T) {
		cfg := base()
		cfg.CBMinRequests = 0
		assert.Error(t, cfg.validate())
	})

	t.Run("CBErrorRate = 0 — ошибка", func(t *testing.T) {
		cfg := base()
		cfg.CBErrorRate = 0
		assert.Error(t, cfg.validate())
	})

	t.Run("CBErrorRate > 1 — ошибка", func(t *testing.T) {
		cfg := base()
		cfg.CBErrorRate = 1.5
		assert.Error(t, cfg.validate())
	})

	t.Run("CBErrorRate = 1 — валидно (100% ошибок)", func(t *testing.T) {
		cfg := base()
		cfg.CBErrorRate = 1.0
		assert.NoError(t, cfg.validate())
	})
}

func TestValidate_Retry(t *testing.T) {
	base := func() Config {
		return Config{
			Target:           "x",
			TLSMode:          TLSModeInsecure,
			Timeout:          time.Second,
			RetryEnabled:     true,
			RetryMaxAttempts: 3,
			RetryBaseDelay:   100 * time.Millisecond,
			RetryMaxDelay:    1 * time.Second,
			RetryMethods:     []string{"/test.Service/Method"},
		}
	}

	t.Run("валидный retry", func(t *testing.T) {
		cfg := base()
		assert.NoError(t, cfg.validate())
	})

	t.Run("RetryMaxAttempts < 2 — ошибка", func(t *testing.T) {
		cfg := base()
		cfg.RetryMaxAttempts = 1
		assert.Error(t, cfg.validate())
	})

	t.Run("RetryBaseDelay = 0 — ошибка", func(t *testing.T) {
		cfg := base()
		cfg.RetryBaseDelay = 0
		assert.Error(t, cfg.validate())
	})

	t.Run("RetryMaxDelay = 0 — ошибка", func(t *testing.T) {
		cfg := base()
		cfg.RetryMaxDelay = 0
		assert.Error(t, cfg.validate())
	})

	t.Run("RetryMaxDelay < RetryBaseDelay — ошибка", func(t *testing.T) {
		cfg := base()
		cfg.RetryBaseDelay = 1 * time.Second
		cfg.RetryMaxDelay = 100 * time.Millisecond
		assert.Error(t, cfg.validate())
	})

	t.Run("пустой whitelist — ошибка", func(t *testing.T) {
		cfg := base()
		cfg.RetryMethods = nil
		assert.Error(t, cfg.validate())
	})

	t.Run("пустой slice whitelist — ошибка", func(t *testing.T) {
		cfg := base()
		cfg.RetryMethods = []string{}
		assert.Error(t, cfg.validate())
	})
}

func TestApplyDefaults_ThenValidate(t *testing.T) {
	t.Run("полный конфиг с CB+Retry проходит валидацию после дефолтов", func(t *testing.T) {
		cfg := Config{
			Target:       "auth-service:50051",
			TLSMode:      TLSModeInsecure,
			CBEnabled:    true,
			RetryEnabled: true,
			RetryMethods: []string{"/auth.AuthService/ValidateToken"},
		}
		cfg.applyDefaults()

		require.NoError(t, cfg.validate())

		// Проверяем, что дефолты применились.
		assert.Equal(t, defaultTimeout, cfg.Timeout)
		assert.Equal(t, defaultCBMaxRequests, cfg.CBMaxRequests)
		assert.Equal(t, defaultCBErrorRate, cfg.CBErrorRate)
		assert.Equal(t, defaultRetryMaxAttempts, cfg.RetryMaxAttempts)
	})

	t.Run("конфиг без CB+Retry проходит валидацию", func(t *testing.T) {
		cfg := Config{
			Target:  "auth-service:50051",
			TLSMode: TLSModeInsecure,
		}
		cfg.applyDefaults()

		require.NoError(t, cfg.validate())
		assert.Equal(t, defaultTimeout, cfg.Timeout)
	})
}