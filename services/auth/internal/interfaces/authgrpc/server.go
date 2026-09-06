package authgrpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/auth"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/usecases"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

type AuthServer struct {
	auth.UnimplementedAuthServiceServer
	registerUC      *usecases.RegisterUseCase
	loginUC         *usecases.LoginUseCase
	validateTokenUC *usecases.ValidateTokenUseCase
}

func NewAuthServer(
	registerUC *usecases.RegisterUseCase,
	loginUC *usecases.LoginUseCase,
	validateTokenUC *usecases.ValidateTokenUseCase,
) *AuthServer {
	return &AuthServer{
		registerUC:      registerUC,
		loginUC:         loginUC,
		validateTokenUC: validateTokenUC,
	}
}

func (s *AuthServer) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	user, err := s.registerUC.Execute(ctx, req.Email, req.Password, req.FullName)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		case errors.Is(err, domain.ErrInvalidEmail):
			return nil, status.Error(codes.InvalidArgument, "invalid email")
		case errors.Is(err, domain.ErrInvalidFullName):
			return nil, status.Error(codes.InvalidArgument, "invalid full name")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}
	return &auth.RegisterResponse{
		Id:       user.ID.String(),
		Email:    user.Email.String(),
		FullName: user.FullName.String(),
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	token, err := s.loginUC.Execute(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &auth.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	}, nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	userID, err := s.validateTokenUC.Execute(ctx, req.Token)
	if err != nil {
		return &auth.ValidateTokenResponse{Valid: false}, nil
	}
	return &auth.ValidateTokenResponse{
		UserId: userID,
		Valid:  true,
	}, nil
}
