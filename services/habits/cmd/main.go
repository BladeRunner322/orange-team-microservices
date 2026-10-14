// Package main — точка входа Habits-сервиса.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/config"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/bootstrap"
)

func main() {
	// TODO: копия из Exercises/Profiles с изменённым именем "habits"
	log, err := logger.NewLogger(logger.MustLoad())
	if err != nil {
		fmt.Println("failed to init logger:", err)
		os.Exit(1)
	}
	defer log.Close()

	_ = log
	_ = config.MustLoad()
	_ = bootstrap.New
	_ = os.Interrupt
	_ = signal.Notify
	_ = syscall.SIGTERM
	_ = time.Second
}