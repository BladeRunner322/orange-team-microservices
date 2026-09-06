package bootstrap

import (
	"fmt"
	"net"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/auth"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/config"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/usecases"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/infrastructure/jwt"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/infrastructure/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/interfaces/authgrpc"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	grpcServer *grpc.Server
	listener   net.Listener
	logger     *logger.Logger
	config     config.Config
}

func New(cfg config.Config, log *logger.Logger) (*App, error) {
	repo := postgres.NewInMemoryRepository()

	tokenManager := jwt.NewManager(
		cfg.JWTSecret,
		cfg.JWTIssuer,
		cfg.JWTAudience,
		cfg.JWTExpiration,
	)

	registerUC := usecases.NewRegister(repo, log)
	loginUC := usecases.NewLogin(repo, tokenManager, log)
	validateUC := usecases.NewValidateToken(tokenManager, log)

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

	// Включаем reflection только если явно разрешено
	if cfg.EnableReflection {
		reflection.Register(s)
	}

	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	return &App{
		grpcServer: s,
		listener:   lis,
		logger:     log,
		config:     cfg,
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
