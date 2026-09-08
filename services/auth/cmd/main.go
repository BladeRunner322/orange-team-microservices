package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/config"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/bootstrap"
)

func main() {
	// 1. Инициализация логгера
	log, err := logger.NewLogger(logger.MustLoad())
	if err != nil {
		fmt.Println("failed to init application logger:", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("starting auth service")

	// 2. Загрузка конфигурации сервиса (порт, JWT, рефлексия)
	cfg := config.MustLoad()
	log.Info("config loaded", "port", cfg.GRPCPort)

	// 3. Сборка приложения: БД → репозиторий → use cases → gRPC-сервер
	app, err := bootstrap.New(cfg, log)
	if err != nil {
		log.Error("failed to bootstrap app", "error", err)
		os.Exit(1)
	}
	defer app.Close() // закрываем пул соединений с БД

	// 4. Запуск gRPC-сервера в горутине
	go func() {
		if err := app.Run(); err != nil {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// 5. Ожидание сигнала завершения (Ctrl+C / SIGTERM)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	log.Info("waiting for signal...")
	<-stop

	log.Info("signal received, shutting down...")

	// 6. Graceful shutdown с таймаутом 5 секунд
	done := make(chan struct{})
	go func() {
		app.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
		log.Info("server stopped gracefully")
	case <-time.After(5 * time.Second):
		log.Warn("graceful stop timeout, forcing stop")
		app.Stop()
	}
}
