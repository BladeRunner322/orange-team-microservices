// Package client предоставляет общий конструктор gRPC-клиентов
// с единой настройкой TLS, таймаутом на вызов, Circuit Breaker,
// Retry и автоматическим прокидыванием user_id из context в metadata.
package client

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/BladeRunner322/orange-team-microservices/pkg/grpc/interceptors"
)

// New создаёт gRPC-соединение с указанным конфигом.
//
// Все клиенты в проекте должны создаваться через этот конструктор —
// это гарантирует:
//   - единый TLS-режим;
//   - автоматическое прокидывание user_id в metadata;
//   - таймаут на каждый вызов (если cfg.Timeout == 0 — применяется defaultTimeout);
//   - Circuit Breaker (если cfg.CBEnabled);
//   - Retry для методов из whitelist (если cfg.RetryEnabled);
//   - клиентские метрики для исходящих gRPC-вызовов.
//
// Цепочка интерсепторов (снаружи внутрь):
//
//	CB → Retry → ClientMetrics → Timeout → UserID → RequestID → invoker
//
// Порядок зафиксирован в ADR-020:
//   - CB снаружи Retry — видит финальный результат retry-цикла;
//   - Timeout внутри Retry — таймаут на попытку, не на цикл;
//   - ClientMetrics внутри Retry — считает каждую попытку отдельно;
//   - UserID и RequestID — финализирующие, кладут данные в metadata.
func New(ctx context.Context, cfg Config) (*grpc.ClientConn, error) {
	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate grpc client config: %w", err)
	}

	chain := make([]grpc.UnaryClientInterceptor, 0, 6)

	// 1. Circuit Breaker — снаружи всей цепочки. Видит финальный
	// результат retry и обновляет своё состояние по нему.
	if cfg.CBEnabled {
		cb := interceptors.NewCircuitBreaker(
			cfg.Target,
			cfg.CBMaxRequests,
			cfg.CBInterval,
			cfg.CBTimeout,
			cfg.CBMinRequests,
			cfg.CBErrorRate,
		)
		chain = append(chain, interceptors.CircuitBreakerInterceptor(cb))
	}

	// 2. Retry — только для методов из whitelist.
	if cfg.RetryEnabled {
		chain = append(chain, interceptors.RetryInterceptor(
			cfg.Target,
			cfg.RetryMethods,
			cfg.RetryMaxAttempts,
			cfg.RetryBaseDelay,
			cfg.RetryMaxDelay,
		))
	}

	// 3. ClientMetrics — внутри Retry, чтобы видеть каждую попытку.
	chain = append(chain, interceptors.ClientMetricsInterceptor(cfg.Target))

	// 4. Timeout — per-attempt, внутри Retry.
	chain = append(chain, interceptors.TimeoutInterceptor(cfg.Timeout))

	// 5. UserID и RequestID — финализирующие, кладут данные в metadata.
	chain = append(chain,
		interceptors.UserIDClientInterceptor(),
		interceptors.RequestIDClientInterceptor(),
	)

	opts := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(chain...),
	}

	switch cfg.TLSMode {
	case TLSModeDisabled:
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	case TLSModeInsecure, TLSModeVerify:
		opts = append(opts, grpc.WithTransportCredentials(
			credentials.NewTLS(cfg.tlsConfig()),
		))
	}

	conn, err := grpc.NewClient(cfg.Target, opts...)
	if err != nil {
		return nil, fmt.Errorf("create grpc client for %q: %w", cfg.Target, err)
	}

	return conn, nil
}