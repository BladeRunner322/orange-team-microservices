package middleware

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/ctxkeys"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/pkg/ratelimit"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== ParseTrustedProxies =====

func TestParseTrustedProxies(t *testing.T) {
	t.Run("пустой список", func(t *testing.T) {
		nets, err := ParseTrustedProxies(nil)
		require.NoError(t, err)
		assert.Empty(t, nets)
	})

	t.Run("один валидный CIDR", func(t *testing.T) {
		nets, err := ParseTrustedProxies([]string{"10.0.0.0/8"})
		require.NoError(t, err)
		require.Len(t, nets, 1)
		assert.True(t, nets[0].Contains(net.ParseIP("10.0.0.5")))
	})

	t.Run("несколько CIDR", func(t *testing.T) {
		nets, err := ParseTrustedProxies([]string{"10.0.0.0/8", "192.168.0.0/16"})
		require.NoError(t, err)
		assert.Len(t, nets, 2)
	})

	t.Run("пустые строки и пробелы игнорируются", func(t *testing.T) {
		nets, err := ParseTrustedProxies([]string{"", "  ", "10.0.0.0/8"})
		require.NoError(t, err)
		assert.Len(t, nets, 1)
	})

	t.Run("невалидный CIDR — ошибка", func(t *testing.T) {
		_, err := ParseTrustedProxies([]string{"not-a-cidr"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not-a-cidr")
	})
}

// ===== isTrusted =====

func TestIsTrusted(t *testing.T) {
	trusted := mustParseCIDRs(t, "10.0.0.0/8", "192.168.0.0/16")

	t.Run("пустой список — false", func(t *testing.T) {
		assert.False(t, isTrusted("10.0.0.5", nil))
	})

	t.Run("IP входит в одну из сетей — true", func(t *testing.T) {
		assert.True(t, isTrusted("10.0.0.5", trusted))
		assert.True(t, isTrusted("192.168.1.100", trusted))
	})

	t.Run("IP не входит ни в одну сеть — false", func(t *testing.T) {
		assert.False(t, isTrusted("8.8.8.8", trusted))
	})

	t.Run("невалидный IP — false", func(t *testing.T) {
		assert.False(t, isTrusted("not-an-ip", trusted))
	})
}

// ===== clientIP =====

func TestClientIP(t *testing.T) {
	trusted := mustParseCIDRs(t, "10.0.0.0/8")

	t.Run("trusted пуст, XFF есть — игнорируем XFF", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "8.8.8.8:12345"
		r.Header.Set("X-Forwarded-For", "1.2.3.4")

		ip := clientIP(r, nil)

		assert.Equal(t, "8.8.8.8", ip, "XFF должен игнорироваться без trusted proxies")
	})

	t.Run("trusted пуст, XFF нет — RemoteAddr", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "8.8.8.8:12345"

		ip := clientIP(r, nil)

		assert.Equal(t, "8.8.8.8", ip)
	})

	t.Run("RemoteAddr trusted, XFF есть — берём XFF", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "10.0.0.5:12345"
		r.Header.Set("X-Forwarded-For", "1.2.3.4")

		ip := clientIP(r, trusted)

		assert.Equal(t, "1.2.3.4", ip)
	})

	t.Run("RemoteAddr trusted, XFF нет — RemoteAddr", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "10.0.0.5:12345"

		ip := clientIP(r, trusted)

		assert.Equal(t, "10.0.0.5", ip)
	})

	t.Run("RemoteAddr не trusted, XFF есть — игнорируем XFF", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "8.8.8.8:12345"
		r.Header.Set("X-Forwarded-For", "1.2.3.4")

		ip := clientIP(r, trusted)

		assert.Equal(t, "8.8.8.8", ip, "XFF от не-trusted источника должен игнорироваться")
	})

	t.Run("XFF с несколькими IP — берём первый не-trusted справа", func(t *testing.T) {
		// XFF = "клиент, предыдущий_прокси, текущий_прокси"
		// Обходим справа: 10.0.0.5 trusted → пропускаем, 5.6.7.8 не trusted → это клиент.
		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "10.0.0.5:12345"
		r.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8, 10.0.0.5")

		ip := clientIP(r, trusted)

		assert.Equal(t, "5.6.7.8", ip)
	})

	t.Run("XFF с пробелами — trim", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "10.0.0.5:12345"
		r.Header.Set("X-Forwarded-For", "  1.2.3.4  , 5.6.7.8")

		ip := clientIP(r, trusted)

		assert.Equal(t, "5.6.7.8", ip)
	})

	t.Run("RemoteAddr без порта — возвращается как есть", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "8.8.8.8"

		ip := clientIP(r, nil)

		assert.Equal(t, "8.8.8.8", ip)
	})

	t.Run("Caddy добавляет свой IP в XFF — берём реального клиента", func(t *testing.T) {
		// Клиент → Caddy (172.18.0.11). Caddy дописал свой IP справа.
		// Реальный клиент — последний не-trusted справа.
		trustedDocker := mustParseCIDRs(t, "172.18.0.0/16")

		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "172.18.0.11:54321"
		r.Header.Set("X-Forwarded-For", "203.0.113.42, 172.18.0.11")

		ip := clientIP(r, trustedDocker)

		assert.Equal(t, "203.0.113.42", ip)
	})

	t.Run("клиент подделал XFF — берём реальный IP от Caddy", func(t *testing.T) {
		// Злоумышленник послал XFF: 1.2.3.4 (подделка).
		// Caddy добавил реальный IP клиента справа: 203.0.113.42.
		// XFF на входе Gateway: "1.2.3.4, 203.0.113.42, 172.18.0.11".
		// Обходим справа: 172.18.0.11 trusted, 203.0.113.42 не trusted → реальный клиент.
		trustedDocker := mustParseCIDRs(t, "172.18.0.0/16")

		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "172.18.0.11:54321"
		r.Header.Set("X-Forwarded-For", "1.2.3.4, 203.0.113.42, 172.18.0.11")

		ip := clientIP(r, trustedDocker)

		assert.Equal(t, "203.0.113.42", ip, "подделка слева должна игнорироваться")
	})

	t.Run("все элементы XFF trusted — fall back на RemoteAddr", func(t *testing.T) {
		trustedDocker := mustParseCIDRs(t, "172.18.0.0/16")

		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "172.18.0.11:54321"
		r.Header.Set("X-Forwarded-For", "172.18.0.5, 172.18.0.11")

		ip := clientIP(r, trustedDocker)

		assert.Equal(t, "172.18.0.11", ip)
	})

	t.Run("мусор в XFF пропускается", func(t *testing.T) {
		trustedDocker := mustParseCIDRs(t, "172.18.0.0/16")

		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "172.18.0.11:54321"
		r.Header.Set("X-Forwarded-For", "not-an-ip, 203.0.113.42, 172.18.0.11")

		ip := clientIP(r, trustedDocker)

		assert.Equal(t, "203.0.113.42", ip)
	})

	t.Run("пустые элементы XFF пропускаются", func(t *testing.T) {
		trustedDocker := mustParseCIDRs(t, "172.18.0.0/16")

		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "172.18.0.11:54321"
		r.Header.Set("X-Forwarded-For", ", , 203.0.113.42, 172.18.0.11")

		ip := clientIP(r, trustedDocker)

		assert.Equal(t, "203.0.113.42", ip)
	})
}

// ===== mockLimiter =====

type mockLimiter struct {
	allowFunc func(ctx context.Context, key string, rule ratelimit.Rule) (ratelimit.Result, error)
}

func (m *mockLimiter) Allow(ctx context.Context, key string, rule ratelimit.Rule) (ratelimit.Result, error) {
	if m.allowFunc != nil {
		return m.allowFunc(ctx, key, rule)
	}
	return ratelimit.Result{Allowed: true}, nil
}

// ===== toRule =====

func TestToRule(t *testing.T) {
	in := config.RateLimitRule{Rate: 5, Burst: 10, Interval: time.Minute}

	out := toRule(in)

	assert.Equal(t, 5, out.Rate)
	assert.Equal(t, 10, out.Burst)
	assert.Equal(t, time.Minute, out.Interval)
}

// ===== pickRuleAndKey =====

func TestPickRuleAndKey(t *testing.T) {
	cfg := config.RateLimitConfig{
		Enabled:  true,
		Login:    config.RateLimitRule{Rate: 5, Burst: 5, Interval: time.Minute},
		Register: config.RateLimitRule{Rate: 3, Burst: 3, Interval: time.Minute},
		Refresh:  config.RateLimitRule{Rate: 20, Burst: 20, Interval: time.Minute},
		Default:  config.RateLimitRule{Rate: 60, Burst: 60, Interval: time.Minute},
	}

	t.Run("POST /login с email в body — ключ с IP и email", func(t *testing.T) {
		body := `{"email":"user@example.com","password":"pass"}`
		r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
		r.RemoteAddr = "1.2.3.4:1234"

		rule, key, ok := pickRuleAndKey(r, cfg, nil)

		assert.True(t, ok)
		assert.Equal(t, 5, rule.Rate)
		assert.Equal(t, "rl:login:1.2.3.4:user@example.com", key)
	})

	t.Run("POST /login с невалидным body — ключ только по IP", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("not-json"))
		r.RemoteAddr = "1.2.3.4:1234"

		rule, key, ok := pickRuleAndKey(r, cfg, nil)

		assert.True(t, ok)
		assert.Equal(t, 5, rule.Rate)
		assert.Equal(t, "rl:login:1.2.3.4", key)
	})

	t.Run("POST /register с email в body — ключ с IP и email", func(t *testing.T) {
		body := `{"email":"user@example.com"}`
		r := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
		r.RemoteAddr = "1.2.3.4:1234"

		rule, key, ok := pickRuleAndKey(r, cfg, nil)

		assert.True(t, ok)
		assert.Equal(t, 3, rule.Rate)
		assert.Equal(t, "rl:register:1.2.3.4:user@example.com", key)
	})

	t.Run("POST /refresh — ключ только по IP", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/refresh", nil)
		r.RemoteAddr = "5.6.7.8:5678"

		rule, key, ok := pickRuleAndKey(r, cfg, nil)

		assert.True(t, ok)
		assert.Equal(t, 20, rule.Rate)
		assert.Equal(t, "rl:refresh:5.6.7.8", key)
	})

	t.Run("защищённый маршрут с user_id в context — ключ по user_id", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		ctx := ctxkeys.WithUserID(r.Context(), "user-42")
		r = r.WithContext(ctx)

		rule, key, ok := pickRuleAndKey(r, cfg, nil)

		assert.True(t, ok)
		assert.Equal(t, 60, rule.Rate)
		assert.Equal(t, "rl:api:user-42", key)
	})

	t.Run("защищённый маршрут без user_id — ok=false (fail-open)", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/users/me", nil)

		_, _, ok := pickRuleAndKey(r, cfg, nil)

		assert.False(t, ok)
	})
}

// ===== extractEmail =====

func TestExtractEmail(t *testing.T) {
	t.Run("валидный JSON с email", func(t *testing.T) {
		body := `{"email":"user@example.com","password":"x"}`
		r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))

		email, err := extractEmail(r)

		require.NoError(t, err)
		assert.Equal(t, "user@example.com", email)

		// body восстановлен — можно прочитать снова
		restored, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, body, string(restored))
	})

	t.Run("JSON без email — ошибка", func(t *testing.T) {
		body := `{"password":"x"}`
		r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))

		_, err := extractEmail(r)

		assert.Error(t, err)
	})

	t.Run("невалидный JSON — ошибка", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("not-json"))

		_, err := extractEmail(r)

		assert.Error(t, err)
	})

	t.Run("body nil — ошибка", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.Body = nil

		_, err := extractEmail(r)

		assert.Error(t, err)
	})
}

// ===== RateLimitMiddleware =====

func TestRateLimitMiddleware(t *testing.T) {
	cfg := config.RateLimitConfig{
		Enabled: true,
		Login:   config.RateLimitRule{Rate: 5, Burst: 5, Interval: time.Minute},
		Default: config.RateLimitRule{Rate: 60, Burst: 60, Interval: time.Minute},
	}
	log := logger.NewTestLogger()
	nextOK := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("disabled — пропускает без вызова лимитера", func(t *testing.T) {
		called := false
		mock := &mockLimiter{
			allowFunc: func(ctx context.Context, key string, rule ratelimit.Rule) (ratelimit.Result, error) {
				called = true
				return ratelimit.Result{Allowed: true}, nil
			},
		}
		disabled := config.RateLimitConfig{Enabled: false}
		mw := RateLimitMiddleware(mock, nil, disabled, log)

		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"a@b.c"}`))
		rec := httptest.NewRecorder()
		mw(nextOK).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.False(t, called)
	})

	t.Run("разрешено — пропускает", func(t *testing.T) {
		mock := &mockLimiter{
			allowFunc: func(ctx context.Context, key string, rule ratelimit.Rule) (ratelimit.Result, error) {
				return ratelimit.Result{Allowed: true}, nil
			},
		}
		mw := RateLimitMiddleware(mock, nil, cfg, log)

		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"a@b.c"}`))
		req.RemoteAddr = "1.2.3.4:1234"
		rec := httptest.NewRecorder()
		mw(nextOK).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("превышен — 429 с Retry-After", func(t *testing.T) {
		mock := &mockLimiter{
			allowFunc: func(ctx context.Context, key string, rule ratelimit.Rule) (ratelimit.Result, error) {
				return ratelimit.Result{Allowed: false, RetryAfter: 11 * time.Second}, nil
			},
		}
		mw := RateLimitMiddleware(mock, nil, cfg, log)

		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"a@b.c"}`))
		req.RemoteAddr = "1.2.3.4:1234"
		rec := httptest.NewRecorder()
		mw(nextOK).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusTooManyRequests, rec.Code)
		assert.Equal(t, "11", rec.Header().Get("Retry-After"))
	})

	t.Run("RetryAfter < 1s — округляется до 1", func(t *testing.T) {
		mock := &mockLimiter{
			allowFunc: func(ctx context.Context, key string, rule ratelimit.Rule) (ratelimit.Result, error) {
				return ratelimit.Result{Allowed: false, RetryAfter: 100 * time.Millisecond}, nil
			},
		}
		mw := RateLimitMiddleware(mock, nil, cfg, log)

		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"a@b.c"}`))
		req.RemoteAddr = "1.2.3.4:1234"
		rec := httptest.NewRecorder()
		mw(nextOK).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusTooManyRequests, rec.Code)
		assert.Equal(t, "1", rec.Header().Get("Retry-After"))
	})

	t.Run("ошибка лимитера — fail-open, пропускает", func(t *testing.T) {
		mock := &mockLimiter{
			allowFunc: func(ctx context.Context, key string, rule ratelimit.Rule) (ratelimit.Result, error) {
				return ratelimit.Result{}, errors.New("redis down")
			},
		}
		mw := RateLimitMiddleware(mock, nil, cfg, log)

		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"a@b.c"}`))
		req.RemoteAddr = "1.2.3.4:1234"
		rec := httptest.NewRecorder()
		mw(nextOK).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("неизвестный маршрут без user_id — пропускает", func(t *testing.T) {
		called := false
		mock := &mockLimiter{
			allowFunc: func(ctx context.Context, key string, rule ratelimit.Rule) (ratelimit.Result, error) {
				called = true
				return ratelimit.Result{Allowed: true}, nil
			},
		}
		mw := RateLimitMiddleware(mock, nil, cfg, log)

		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		rec := httptest.NewRecorder()
		mw(nextOK).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.False(t, called)
	})
}

// mustParseCIDRs — хелпер для тестов: парсит CIDR или падает.
func mustParseCIDRs(t *testing.T, cidrs ...string) []*net.IPNet {
	t.Helper()
	nets, err := ParseTrustedProxies(cidrs)
	require.NoError(t, err)
	return nets
}
