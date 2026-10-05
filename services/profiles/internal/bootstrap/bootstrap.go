package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/profiles"
	grpcserver "github.com/BladeRunner322/orange-team-microservices/pkg/grpc/server"
	"github.com/BladeRunner322/orange-team-microservices/pkg/health"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/config"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/application/usecases"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/infrastructure/postgres_repo"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/interfaces/profilesgrpc"
	"google.golang.org/grpc"
)

// App объединяет все компоненты приложения.
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
	// 2. РЕПОЗИТОРИЙ (реализация для profiles)
	// ============================================================
	repo := postgres_repo.NewRepository(pool)

	// ============================================================
	// 3. USE CASES
	// ============================================================
	getMyProfileUC := usecases.NewGetMyProfile(repo, log)
	getProfileUC := usecases.NewGetProfile(repo, log)
	patchMyProfileUC := usecases.NewPatchMyProfile(repo, log)
	deleteMyProfileUC := usecases.NewDeleteMyProfile(repo, log)
	log.Info("use cases initialized")

	// ============================================================
	// 4. gRPC СЕРВЕР С ИНТЕРСЕПТОРАМИ
	// ============================================================
	s, err := grpcserver.New(grpcserver.Config{
		EnableTLS:        cfg.EnableTLS,
		TLSCertFile:      cfg.TLSCertFile,
		TLSKeyFile:       cfg.TLSKeyFile,
		EnableReflection: cfg.EnableReflection,
		WithUserID:       true, // Profiles читает user_id из metadata
		Logger:           log,
	})
	if err != nil {
		return nil, fmt.Errorf("create grpc server: %w", err)
	}

	// Регистрация gRPC-сервиса
	profiles.RegisterProfilesServiceServer(s, profilesgrpc.NewServer(getMyProfileUC, getProfileUC, patchMyProfileUC, deleteMyProfileUC))
	log.Info("gRPC server registered")

	// ============================================================
	// 5. gRPC ЛИСТЕНЕР
	// ============================================================
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}
	log.Info("gRPC listener created", "addr", cfg.GRPCPort)

	// ============================================================
	// 6. HTTP СЕРВЕР (HEALTHCHECK + METRICS + READINESS)
	// ============================================================
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", health.HealthHandler("profiles"))
	healthMux.HandleFunc("/ready", health.ReadyHandler(map[string]health.Check{
		"postgres": health.PostgresCheck(pool),
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
	// 7. СБОРКА APP
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
	a.logger.Info("Profiles service listening", "addr", a.config.GRPCPort)
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

	if a.pool != nil {
		a.pool.Close()
	}

	a.logger.Info("all resources closed")
}
