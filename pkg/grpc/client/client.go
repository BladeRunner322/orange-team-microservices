// Package client предоставляет общий конструктор gRPC-клиентов
// с единой настройкой TLS и автоматическим прокидыванием user_id
// из context в metadata.
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
// это гарантирует единый TLS-режим и автоматическое прокидывание
// user_id в metadata для downstream-сервисов.
func New(ctx context.Context, cfg Config) (*grpc.ClientConn, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate grpc client config: %w", err)
	}

	opts := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(
			interceptors.UserIDClientInterceptor(),
		),
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
