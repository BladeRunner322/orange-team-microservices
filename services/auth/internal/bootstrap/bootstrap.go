package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/auth"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/config"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/usecases"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/health"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/infrastructure/jwt"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/infrastructure/postgres_repo"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/interfaces/authgrpc"
)

type App struct {
	grpcServer *grpc.Server
	listener   net.Listener
	logger     *logger.Logger
	config     config.Config
	pool       *postgres.PgxPool
	httpServer *http.Server
}

func New(cfg config.Config, log *logger.Logger) (*App, error) {
	ctx := context.Background()

	// 1. ПОДКЛЮЧЕНИЕ К POSTGRESQL
	pgCfg := postgres.MustLoad()
	pool, err := postgres.NewPgxPool(ctx, pgCfg)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	log.Info("postgres connection pool created")

	// 2. РЕПОЗИТОРИЙ (реализация для auth)
	repo := postgres_repo.NewRepository(pool)

	// 3. JWT-МЕНЕДЖЕР
	tokenManager := jwt.NewManager(
		cfg.JWTSecret,
		cfg.JWTIssuer,
		cfg.JWTAudience,
		cfg.JWTExpiration,
	)
	log.Info("jwt manager initialized")

	// 4. USE CASES
	registerUC := usecases.NewRegister(repo, log)
	loginUC := usecases.NewLogin(repo, tokenManager, log)
	validateUC := usecases.NewValidateToken(tokenManager, log)
	log.Info("use cases initialized")

	// 5. gRPC СЕРВЕР С ИНТЕРСЕПТОРАМИ
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(
				recovery.WithRecoveryHandler(func(p interface{}) error {
					log.Error("panic recovered", "panic", p)
					return nil
				}),
			),
			authgrpc.LoggingInterceptor(log),
		),
	)
	auth.RegisterAuthServiceServer(s, authgrpc.NewServer(registerUC, loginUC, validateUC))

	if cfg.EnableReflection {
		reflection.Register(s)
	}
	log.Info("gRPC server registered")

	// 6. gRPC ЛИСТЕНЕР
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}
	log.Info("gRPC listener created", "addr", cfg.GRPCPort)

	// 7. HEALTHCHECK HTTP-СЕРВЕР (на отдельном порту 8080)
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", health.Handler())
	healthMux.HandleFunc("/ready", health.Handler())

	httpSrv := &http.Server{
		Addr:         ":8080",
		Handler:      healthMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("healthcheck server listening", "addr", ":8080")
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("healthcheck server failed", "error", err)
		}
	}()

	// 8. СБОРКА APP
	return &App{
		grpcServer: s,
		listener:   lis,
		logger:     log,
		config:     cfg,
		pool:       pool,
		httpServer: httpSrv,
	}, nil
}

// ЗАПУСК
func (a *App) Run() error {
	a.logger.Info("Auth service listening", "addr", a.config.GRPCPort)
	return a.grpcServer.Serve(a.listener)
}

// GRACEFUL SHUTDOWN
func (a *App) GracefulStop() {
	a.grpcServer.GracefulStop()
}

func (a *App) Stop() {
	a.grpcServer.Stop()
}

// ЗАКРЫТИЕ РЕСУРСОВ
func (a *App) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Закрываем HTTP-сервер (healthcheck)
	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			a.logger.Error("healthcheck server shutdown error", "error", err)
		}
	}

	// Закрываем пул соединений с БД
	if a.pool != nil {
		a.pool.Close()
	}
}
