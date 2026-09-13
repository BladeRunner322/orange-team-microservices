package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/pkg/ratelimit"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/config"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/httputil"
)

// RateLimitMiddleware ограничивает частоту запросов.
//
// Работает по-разному в зависимости от пути:
//   - /login, /register → ключ = IP + email из body
//   - /refresh         → ключ = IP
//   - прочие           → ключ = user_id из контекста (если есть)
//
// Если правило или ключ определить не удалось — запрос пропускается
// без проверки (fail-open), чтобы не блокировать рабочий трафик.
func RateLimitMiddleware(
	limiter *ratelimit.Limiter,
	cfg config.RateLimitConfig,
	log *logger.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			rule, key, ok := pickRuleAndKey(r, cfg)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			result, err := limiter.Allow(r.Context(), key, rule)
			if err != nil {
				// fail-open: если Redis недоступен, не блокируем трафик
				log.Error("rate limit check failed", "error", err, "key", key)
				next.ServeHTTP(w, r)
				return
			}

			if !result.Allowed {
				retryAfter := int(result.RetryAfter.Seconds())
				if retryAfter < 1 {
					retryAfter = 1
				}
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				log.Warn("rate limit exceeded",
					"key", key,
					"path", r.URL.Path,
					"retry_after_sec", retryAfter,
				)
				httputil.SendError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// pickRuleAndKey выбирает правило и ключ для запроса.
// Возвращает ok=false, если правило применить нельзя — тогда запрос пропускается.
func pickRuleAndKey(r *http.Request, cfg config.RateLimitConfig) (ratelimit.Rule, string, bool) {
	ip := clientIP(r)

	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/login":
		email, err := extractEmail(r)
		if err == nil && email != "" {
			return toRule(cfg.Login), fmt.Sprintf("rl:login:%s:%s", ip, email), true
		}
		// если body не прочитали — падаем на IP
		return toRule(cfg.Login), fmt.Sprintf("rl:login:%s", ip), true

	case r.Method == http.MethodPost && r.URL.Path == "/register":
		email, err := extractEmail(r)
		if err == nil && email != "" {
			return toRule(cfg.Register), fmt.Sprintf("rl:register:%s:%s", ip, email), true
		}
		return toRule(cfg.Register), fmt.Sprintf("rl:register:%s", ip), true

	case r.Method == http.MethodPost && r.URL.Path == "/refresh":
		return toRule(cfg.Refresh), fmt.Sprintf("rl:refresh:%s", ip), true

	default:
		// защищённые маршруты — нужен user_id из контекста
		userID, ok := GetUserID(r.Context())
		if !ok {
			return ratelimit.Rule{}, "", false
		}
		return toRule(cfg.Default), fmt.Sprintf("rl:api:%s", userID), true
	}
}

// toRule конвертирует config.RateLimitRule в ratelimit.Rule.
func toRule(r config.RateLimitRule) ratelimit.Rule {
	return ratelimit.Rule{
		Rate:     r.Rate,
		Burst:    r.Burst,
		Interval: r.Interval,
	}
}

// clientIP извлекает IP клиента с учётом X-Forwarded-For (за прокси).
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// extractEmail читает body, вытаскивает email и восстанавливает body.
//
// Важно: body восстанавливается, чтобы handler мог его прочитать заново.
func extractEmail(r *http.Request) (string, error) {
	if r.Body == nil {
		return "", fmt.Errorf("empty body")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	// Возвращаем body обратно — handler прочитает.
	r.Body = io.NopCloser(bytes.NewReader(body))

	var req struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return "", fmt.Errorf("unmarshal body: %w", err)
	}
	if req.Email == "" {
		return "", fmt.Errorf("empty email")
	}

	return req.Email, nil
}
