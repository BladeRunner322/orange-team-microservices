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

type Server struct {
	auth.UnimplementedAuthServiceServer
	registerUC      *usecases.Register
	loginUC         *usecases.Login
	validateTokenUC *usecases.Validate
	refreshTokenUC  *usecases.RefreshToken
	logoutUC        *usecases.Logout
}

func NewServer(
	registerUC *usecases.Register,
	loginUC *usecases.Login,
	validateTokenUC *usecases.Validate,
	refreshTokenUC *usecases.RefreshToken,
	logoutUC *usecases.Logout,
) *Server {
	return &Server{
		registerUC:      registerUC,
		loginUC:         loginUC,
		validateTokenUC: validateTokenUC,
		refreshTokenUC:  refreshTokenUC,
		logoutUC:        logoutUC,
	}
}

func (s *Server) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	email, password, fullName := ToDomainRegisterParams(req)
	user, err := s.registerUC.Execute(ctx, email, password, fullName)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		case errors.Is(err, domain.ErrInvalidEmail):
			return nil, status.Error(codes.InvalidArgument, "invalid email")
		case errors.Is(err, domain.ErrInvalidFullName):
			return nil, status.Error(codes.InvalidArgument, "invalid full name")
		case errors.Is(err, domain.ErrWeakPassword):
			return nil, status.Error(codes.InvalidArgument, "password must be at least 8 characters long")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}
	return ToProtoRegisterResponse(user), nil
}

func (s *Server) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	email, password := ToDomainLoginParams(req)
	result, err := s.loginUC.Execute(ctx, email, password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return ToProtoLoginResponse(result.AccessToken, result.RefreshToken), nil
}

func (s *Server) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	token := ToDomainValidateTokenParams(req)
	userID, err := s.validateTokenUC.Execute(ctx, token)
	if err != nil {
		return &auth.ValidateTokenResponse{Valid: false}, nil
	}
	return ToProtoValidateTokenResponse(userID, true), nil
}

func (s *Server) RefreshToken(
	ctx context.Context,
	req *auth.RefreshTokenRequest,
) (*auth.RefreshTokenResponse, error) {
	result, err := s.refreshTokenUC.Execute(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) {
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return ToProtoRefreshTokenResponse(result.AccessToken, result.RefreshToken), nil
}

func (s *Server) Logout(
	ctx context.Context,
	req *auth.LogoutRequest,
) (*auth.LogoutResponse, error) {
	err := s.logoutUC.Execute(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) {
			return nil, status.Error(codes.InvalidArgument, "invalid refresh token")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &auth.LogoutResponse{}, nil
}
