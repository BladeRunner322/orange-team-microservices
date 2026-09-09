package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/config"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/bootstrap"
)

func main() {
	// 1. Логгер
	log, err := logger.NewLogger(logger.MustLoad())
	if err != nil {
		fmt.Println("failed to init logger:", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("starting API Gateway")

	// 2. Конфиг
	cfg := config.MustLoad()
	log.Info("config loaded", "http_port", cfg.HTTPPort, "auth_grpc_addr", cfg.AuthGRPCAddr)

	// 3. Сборка приложения
	app, err := bootstrap.New(cfg, log)
	if err != nil {
		log.Error("failed to bootstrap app", "error", err)
		os.Exit(1)
	}
	defer app.Close()

	// 4. Запуск сервера в горутине
	go func() {
		if err := app.Run(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// 5. Ожидание сигнала
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	log.Info("waiting for signal...")
	<-stop

	// 6. Graceful shutdown
	log.Info("shutting down...")
	if err := app.GracefulStop(); err != nil {
		log.Error("graceful stop error", "error", err)
	}
	log.Info("shutdown complete")
}
