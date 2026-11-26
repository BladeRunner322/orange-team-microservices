package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/leaderboard"
	grpcserver "github.com/BladeRunner322/orange-team-microservices/pkg/grpc/server"
	"github.com/BladeRunner322/orange-team-microservices/pkg/health"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/leaderboard/config"
	"github.com/BladeRunner322/orange-team-microservices/services/leaderboard/internal/interfaces/leaderboardgrpc"
	"google.golang.org/grpc"
)

// App объединяет gRPC-сервер и HTTP-сервер служебных эндпоинтов.
type App struct {
	grpcServer *grpc.Server
	listener   net.Listener
	httpServer *http.Server
	logger     *logger.Logger
}

func New(cfg config.Config, log *logger.Logger) (*App, error) {
	// TODO: Собрать клиент Workouts, хранилище топов и сценарии Leaderboard.
	// TODO: Запускать обновление топов при старте и по таймеру, останавливать воркер при завершении сервиса.

	s, err := grpcserver.New(grpcserver.Config{
		EnableTLS:        cfg.EnableTLS,
		TLSCertFile:      cfg.TLSCertFile,
		TLSKeyFile:       cfg.TLSKeyFile,
		EnableReflection: cfg.EnableReflection,
		Logger:           log,
	})
	if err != nil {
		return nil, fmt.Errorf("create grpc server: %w", err)
	}
	leaderboard.RegisterLeaderboardServiceServer(s, leaderboardgrpc.NewServer())

	listener, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		return nil, fmt.Errorf("listen for grpc: %w", err)
	}

	httpListener, err := net.Listen("tcp", cfg.HTTPPort)
	if err != nil {
		_ = listener.Close()
		return nil, fmt.Errorf("listen for http: %w", err)
	}

	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", health.HealthHandler("leaderboard"))
	// TODO: Добавить проверку готовности топов за текущие периоды.
	healthMux.HandleFunc("/ready", health.ReadyHandler(nil))
	healthMux.Handle("/metrics", health.MetricsHandler())

	httpServer := &http.Server{
		Addr:         cfg.HTTPPort,
		Handler:      healthMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("HTTP server listening", "addr", httpListener.Addr())
		if err := httpServer.Serve(httpListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server failed", "error", err)
		}
	}()

	return &App{
		grpcServer: s,
		listener:   listener,
		httpServer: httpServer,
		logger:     log,
	}, nil
}

func (a *App) Run() error {
	a.logger.Info("Leaderboard service listening", "addr", a.listener.Addr())
	if err := a.grpcServer.Serve(a.listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return err
	}
	return nil
}

func (a *App) GracefulStop() {
	a.grpcServer.GracefulStop()
}

func (a *App) Stop() {
	a.grpcServer.Stop()
}

func (a *App) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.logger.Error("HTTP server shutdown error", "error", err)
	}
}
