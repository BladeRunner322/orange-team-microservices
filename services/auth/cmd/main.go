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
	log, err := logger.NewLogger(logger.MustLoad())
	if err != nil {
		fmt.Println("failed to init application logger:", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("starting auth service")

	cfg := config.MustLoad()
	log.Info("config loaded", "port", cfg.GRPCPort)

	app, err := bootstrap.New(cfg, log)
	if err != nil {
		log.Error("failed to bootstrap app", "error", err)
		os.Exit(1)
	}

	go func() {
		if err := app.Run(); err != nil {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	log.Info("waiting for signal...")
	<-stop

	log.Info("signal received, shutting down...")

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
