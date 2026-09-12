package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/auth"
	"github.com/BladeRunner322/orange-team-microservices/pkg/grpc/interceptors"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/pkg/metrics"
	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/config"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/usecases"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/infrastructure/jwt"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/infrastructure/postgres_repo"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/interfaces/authgrpc"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/interfaces/http/health"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

// App — структура, объединяющая все компоненты приложения.
type App struct {
	grpcServer *grpc.Server
	listener   net.Listener
	logger     *logger.Logger
	config     config.Config
	pool       *postgres.PgxPool
	httpServer *http.Server
}

// New — сборка всех зависимостей и запуск компонентов.
func New(cfg config.Config, log *logger.Logger) (*App, error) {
	ctx := context.Background()

	// ============================================================
	// 1. ПОДКЛЮЧЕНИЕ К POSTGRESQL
	// ============================================================
	pgCfg := postgres.MustLoad()
	pool, err := postgres.NewPgxPool(ctx, pgCfg)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	log.Info("postgres connection pool created")

	// ============================================================
	// 2. РЕПОЗИТОРИЙ (реализация для auth)
	// ============================================================
	repo := postgres_repo.NewRepository(pool)

	// ============================================================
	// 3. JWT-МЕНЕДЖЕР
	// ============================================================
	tokenManager := jwt.NewManager(
		cfg.JWTSecret,
		cfg.JWTIssuer,
		cfg.JWTAudience,
		cfg.JWTExpiration,
	)
	log.Info("jwt manager initialized")

	// ============================================================
	// 4. USE CASES
	// ============================================================
	registerUC := usecases.NewRegister(repo, log)
	loginUC := usecases.NewLogin(repo, tokenManager, log)
	validateUC := usecases.NewValidateToken(tokenManager, log)
	log.Info("use cases initialized")

	// ============================================================
	// 5. МЕТРИКИ (PROMETHEUS)
	// ============================================================
	metrics.Register()
	log.Info("metrics registered")

	// ============================================================
	// 6. gRPC СЕРВЕР С ИНТЕРСЕПТОРАМИ
	// ============================================================

	// Собираем опции сервера
	var grpcOpts []grpc.ServerOption

	// Интерсепторы (были)
	grpcOpts = append(grpcOpts,
		grpc.ChainUnaryInterceptor(
			// 6.1. Метрики (prometheus)
			interceptors.MetricsInterceptor(),
			// 6.2. Восстановление после паники
			interceptors.RecoveryInterceptor(log),
			// 6.3. Логирование запросов
			interceptors.LoggingInterceptor(log),
		),
	)

	// TLS
	if cfg.EnableTLS {
		creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		grpcOpts = append(grpcOpts, grpc.Creds(creds))
		log.Info("TLS enabled for gRPC")
	} else {
		log.Warn("gRPC running without TLS (insecure mode)")
	}

	// Создаём сервер с опциями
	s := grpc.NewServer(grpcOpts...)

	// Регистрация gRPC-сервиса
	auth.RegisterAuthServiceServer(s, authgrpc.NewServer(registerUC, loginUC, validateUC))

	// Режим разработки — включаем reflection для grpcurl
	if cfg.EnableReflection {
		reflection.Register(s)
	}
	log.Info("gRPC server registered")

	// ============================================================
	// 7. gRPC ЛИСТЕНЕР
	// ============================================================
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}
	log.Info("gRPC listener created", "addr", cfg.GRPCPort)

	// ============================================================
	// 8. HTTP СЕРВЕР (HEALTHCHECK + METRICS)
	// ============================================================
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", health.Handler())
	healthMux.Handle("/metrics", promhttp.Handler())

	httpSrv := &http.Server{
		Addr:         cfg.HTTPPort,
		Handler:      healthMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Запускаем HTTP-сервер в горутине
	go func() {
		log.Info("HTTP server listening", "addr", cfg.HTTPPort, "endpoints", "/health, /ready, /metrics")
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server failed", "error", err)
		}
	}()

	// ============================================================
	// 9. СБОРКА APP
	// ============================================================
	return &App{
		grpcServer: s,
		listener:   lis,
		logger:     log,
		config:     cfg,
		pool:       pool,
		httpServer: httpSrv,
	}, nil
}

// ============================================================
// ЗАПУСК gRPC СЕРВЕРА
// ============================================================
func (a *App) Run() error {
	a.logger.Info("Auth service listening", "addr", a.config.GRPCPort)
	return a.grpcServer.Serve(a.listener)
}

// ============================================================
// GRACEFUL SHUTDOWN (ОСТАНОВКА gRPC СЕРВЕРА)
// ============================================================
// Дожидается завершения текущих запросов, затем останавливает сервер.
func (a *App) GracefulStop() {
	a.grpcServer.GracefulStop()
}

// ============================================================
// ПРИНУДИТЕЛЬНАЯ ОСТАНОВКА gRPC СЕРВЕРА
// ============================================================
// Используется при таймауте graceful shutdown.
func (a *App) Stop() {
	a.grpcServer.Stop()
}

// ============================================================
// ЗАКРЫТИЕ РЕСУРСОВ (БД, HTTP, ЛОГГЕР)
// ============================================================
// Закрывается в обратном порядке: сначала HTTP, потом БД.
func (a *App) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Закрываем HTTP-сервер (healthcheck + metrics)
	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			a.logger.Error("HTTP server shutdown error", "error", err)
		}
	}

	// 2. Закрываем пул соединений с БД
	if a.pool != nil {
		a.pool.Close()
	}

	a.logger.Info("all resources closed")
}
