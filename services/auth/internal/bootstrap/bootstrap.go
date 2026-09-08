package bootstrap

import (
	"context"
	"fmt"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/auth"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/config"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/usecases"
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
}

func New(cfg config.Config, log *logger.Logger) (*App, error) {
	ctx := context.Background()

	// 1. Загружаем конфиг для PostgreSQL (паникует при ошибке)
	pgCfg := postgres.MustLoad()
	pool, err := postgres.NewPgxPool(ctx, pgCfg)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	// 2. Репозиторий для auth
	repo := postgres_repo.NewRepository(pool)

	// 3. JWT менеджер
	tokenManager := jwt.NewManager(
		cfg.JWTSecret,
		cfg.JWTIssuer,
		cfg.JWTAudience,
		cfg.JWTExpiration,
	)

	// 4. Use cases
	registerUC := usecases.NewRegister(repo, log)
	loginUC := usecases.NewLogin(repo, tokenManager, log)
	validateUC := usecases.NewValidateToken(tokenManager, log)

	// 5. gRPC сервер
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

	// 6. Слушаем порт
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	return &App{
		grpcServer: s,
		listener:   lis,
		logger:     log,
		config:     cfg,
		pool:       pool,
	}, nil
}

func (a *App) Run() error {
	a.logger.Info("Auth service listening", "addr", a.config.GRPCPort)
	return a.grpcServer.Serve(a.listener)
}

func (a *App) GracefulStop() {
	a.grpcServer.GracefulStop()
}

func (a *App) Stop() {
	a.grpcServer.Stop()
}

// Close закрывает пул соединений
func (a *App) Close() {
	if a.pool != nil {
		a.pool.Close()
		a.logger.Info("PostgreSQL connection pool closed")
	}
}
