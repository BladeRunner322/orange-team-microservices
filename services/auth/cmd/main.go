package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/auth"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/config"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/usecases"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/infrastructure/jwt"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/infrastructure/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/interfaces/authgrpc"
)

func main() {
	cfg := config.MustLoad()

	repo := postgres.NewInMemoryRepository()

	tokenManager := jwt.NewManager(
		cfg.JWTSecret,
		cfg.JWTIssuer,
		cfg.JWTAudience,
		cfg.JWTExpiration,
	)

	registerUC := usecases.NewRegisterUseCase(repo)
	loginUC := usecases.NewLoginUseCase(repo, tokenManager)
	validateUC := usecases.NewValidateTokenUseCase(tokenManager)

	// gRPC сервер с recovery интерсептором
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(
				recovery.WithRecoveryHandler(func(p interface{}) error {
					log.Printf("panic recovered: %v", p)
					return nil // или вернуть ошибку
				}),
			),
		),
	)
	auth.RegisterAuthServiceServer(s, authgrpc.NewAuthServer(registerUC, loginUC, validateUC))

	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Graceful Shutdown
	go func() {
		log.Printf("Auth service listening on %s", cfg.GRPCPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Ожидание сигнала
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down gracefully...")
	s.GracefulStop()
	log.Println("Server stopped")
}
