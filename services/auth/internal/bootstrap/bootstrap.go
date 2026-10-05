package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/auth"
	grpcserver "github.com/BladeRunner322/orange-team-microservices/pkg/grpc/server"
	"github.com/BladeRunner322/orange-team-microservices/pkg/health"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/pkg/redis"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/config"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/usecases"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/infrastructure/jwt"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/infrastructure/postgres_repo"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/infrastructure/redis_repo"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/interfaces/authgrpc"
	"google.golang.org/grpc"
)

// App — структура, объединяющая все компоненты приложения.
type App struct {
	grpcServer  *grpc.Server
	listener    net.Listener
	logger      *logger.Logger
	config      config.Config
	pool        *postgres.PgxPool
	redisClient *redis.Client
	httpServer  *http.Server
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
	// 1.1. ПОДКЛЮЧЕНИЕ К REDIS (refresh tokens)
	// ============================================================
	redisCfg := redis.MustLoad()
	redisClient, err := redis.NewClient(ctx, redisCfg)
	if err != nil {
		return nil, fmt.Errorf("connect to redis: %w", err)
	}
	log.Info("redis client created", "addr", redisCfg.Addr)

	// ============================================================
	// 2. РЕПОЗИТОРИЙ (реализация для auth)
	// ============================================================
	repo := postgres_repo.NewRepository(pool)
	refreshRepo := redis_repo.NewRepository(redisClient)

	// ============================================================
	// 3. JWT-МЕНЕДЖЕР
	// ============================================================
	tokenManager := jwt.NewManager(
		cfg.JWTSecret,
		cfg.JWTIssuer,
		cfg.JWTAudience,
		cfg.AccessTokenTTL,
	)
	log.Info("jwt manager initialized")

	// ============================================================
	// 4. USE CASES
	// ============================================================
	registerUC := usecases.NewRegister(repo, log)
	loginUC := usecases.NewLogin(repo, tokenManager, refreshRepo, cfg.RefreshTokenTTL, log)
	validateUC := usecases.NewValidateToken(tokenManager, log)
	refreshUC := usecases.NewRefreshToken(refreshRepo, repo, tokenManager, cfg.RefreshTokenTTL, log)
	logoutUC := usecases.NewLogout(refreshRepo, log)
	log.Info("use cases initialized")

	// ============================================================
	// 5. МЕТРИКИ (PROMETHEUS)
	// ============================================================
	// gRPC-метрики (grpc_requests_total, grpc_request_duration_ms,
	// grpc_requests_in_flight) регистрируются в promauto при импорте
	// pkg/metrics. Пакет подтягивается транзитивно через pkg/grpc/interceptors
	// (см. interceptors.MetricsInterceptor).
	log.Info("metrics enabled")

	// ============================================================
	// 6. gRPC СЕРВЕР С ИНТЕРСЕПТОРАМИ
	// =
	s, err := grpcserver.New(grpcserver.Config{
		EnableTLS:        cfg.EnableTLS,
		TLSCertFile:      cfg.TLSCertFile,
		TLSKeyFile:       cfg.TLSKeyFile,
		EnableReflection: cfg.EnableReflection,
		WithUserID:       false, // Auth — источник user_id, не читает из metadata
		Logger:           log,
	})
	if err != nil {
		return nil, fmt.Errorf("create grpc server: %w", err)
	}

	// Регистрация gRPC-сервиса
	auth.RegisterAuthServiceServer(s, authgrpc.NewServer(registerUC, loginUC, validateUC, refreshUC, logoutUC))
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
	// 8. HTTP СЕРВЕР (HEALTHCHECK + METRICS + READINESS)
	// ============================================================
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", health.HealthHandler("auth"))
	healthMux.HandleFunc("/ready", health.ReadyHandler(map[string]health.Check{
		"postgres": health.PostgresCheck(pool),
		"redis":    health.RedisCheck(redisClient),
	}))
	healthMux.Handle("/metrics", health.MetricsHandler())

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
		grpcServer:  s,
		listener:    lis,
		logger:      log,
		config:      cfg,
		pool:        pool,
		redisClient: redisClient,
		httpServer:  httpSrv,
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

	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			a.logger.Error("HTTP server shutdown error", "error", err)
		}
	}

	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			a.logger.Error("redis close error", "error", err)
		}
	}

	if a.pool != nil {
		a.pool.Close()
	}

	a.logger.Info("all resources closed")
}
