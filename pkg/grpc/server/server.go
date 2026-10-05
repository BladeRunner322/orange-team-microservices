// Package server — конструктор gRPC-сервера с единой цепочкой
// интерсепторов, опциональным TLS и reflection.
//
// Используется всеми gRPC-сервисами проекта (Auth, Profiles,
// Exercises, Workouts, ...), чтобы не дублировать обвязку.
package server

import (
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/grpc/interceptors"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

// Config — параметры gRPC-сервера.
type Config struct {
	EnableTLS        bool
	TLSCertFile      string
	TLSKeyFile       string
	EnableReflection bool

	// WithUserID включает UserIDServerInterceptor — извлекает
	// x-user-id и x-user-role из gRPC metadata в context.
	// У Auth — false (Auth сам источник user_id).
	// У Profiles, Exercises, Workouts и т.д. — true.
	WithUserID bool

	Logger *logger.Logger
}

// New создаёт gRPC-сервер с цепочкой интерсепторов:
//
//	Metrics → Recovery → [UserID] → Logging
//
// UserID добавляется в цепочку, если cfg.WithUserID == true.
// Если cfg.EnableTLS — подгружает сертификаты.
// Если cfg.EnableReflection — регистрирует reflection.
func New(cfg Config) (*grpc.Server, error) {
	if cfg.Logger == nil {
		return nil, fmt.Errorf("grpc server: logger is required")
	}

	interceptorsList := []grpc.UnaryServerInterceptor{
		interceptors.MetricsInterceptor(),
		interceptors.RecoveryInterceptor(cfg.Logger),
	}

	if cfg.WithUserID {
		interceptorsList = append(interceptorsList, interceptors.UserIDServerInterceptor())
	}

	interceptorsList = append(interceptorsList, interceptors.LoggingInterceptor(cfg.Logger))

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(interceptorsList...),
	}

	if cfg.EnableTLS {
		creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.Creds(creds))
		cfg.Logger.Info("TLS enabled for gRPC")
	} else {
		cfg.Logger.Warn("gRPC running without TLS (insecure mode)")
	}

	s := grpc.NewServer(opts...)

	if cfg.EnableReflection {
		reflection.Register(s)
	}

	return s, nil
}
