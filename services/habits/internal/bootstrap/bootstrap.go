package bootstrap

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/habits"
	grpcserver "github.com/BladeRunner322/orange-team-microservices/pkg/grpc/server"
	"github.com/BladeRunner322/orange-team-microservices/pkg/health"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/config"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/usecases"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/infrastructure/postgres_repo"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/interfaces/habitsgrpc"
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

// New — сборка зависимостей по образцу Exercises/Profiles:
//  1. Postgres pool
//  2. HabitRepository + CompletionRepository
//  3. Usecases (8 штук)
//  4. gRPC-сервер с интерсепторами (WithUserID: true)
//  5. HTTP-сервер /health /ready /metrics
func New(cfg config.Config, log *logger.Logger) (*App, error) {
	// TODO
	return nil, nil
}

// Run запускает gRPC-сервер.
func (a *App) Run() error {
	// TODO
	return nil
}

// GracefulStop останавливает gRPC-сервер.
func (a *App) GracefulStop() {
	// TODO
}

// Stop принудительно останавливает gRPC-сервер.
func (a *App) Stop() {
	// TODO
}

// Close закрывает БД, HTTP, логи.
func (a *App) Close() {
	// TODO
}

var _ = context.Background
var _ = http.NewServeMux
var _ = time.Second
var _ = habits.RegisterHabitsServiceServer
var _ = grpcserver.New
var _ = health.HealthHandler
var _ = usecases.NewCreateHabit
var _ = postgres_repo.NewHabitRepository
var _ = habitsgrpc.NewServer
