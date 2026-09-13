package authgrpc

import (
	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/auth"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

// ===== Преобразования из protobuf в параметры use case =====

// ToDomainRegisterParams преобразует pb.RegisterRequest в параметры для use case.
func ToDomainRegisterParams(req *auth.RegisterRequest) (email, password, fullName string) {
	return req.Email, req.Password, req.FullName
}

// ToDomainLoginParams преобразует pb.LoginRequest в параметры для use case.
func ToDomainLoginParams(req *auth.LoginRequest) (email, password string) {
	return req.Email, req.Password
}

// ToDomainValidateTokenParams преобразует pb.ValidateTokenRequest в параметр для use case.
func ToDomainValidateTokenParams(req *auth.ValidateTokenRequest) string {
	return req.Token
}

// ===== Преобразования из доменных объектов в protobuf =====

// ToProtoRegisterResponse преобразует доменного пользователя в pb.RegisterResponse.
func ToProtoRegisterResponse(user domain.User) *auth.RegisterResponse {
	return &auth.RegisterResponse{
		Id:       user.ID().String(),
		Email:    user.Email().String(),
		FullName: user.FullName().String(),
	}
}

// ToProtoLoginResponse преобразует токен в pb.LoginResponse.
func ToProtoLoginResponse(accessToken, refreshToken string) *auth.LoginResponse {
	return &auth.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}
}

// ToProtoValidateTokenResponse преобразует результат валидации в pb.ValidateTokenResponse.
func ToProtoValidateTokenResponse(userID string, valid bool) *auth.ValidateTokenResponse {
	return &auth.ValidateTokenResponse{
		UserId: userID,
		Valid:  valid,
	}
}

// ToProtoRefreshTokenResponse преобразует новую пару токенов в pb.RefreshTokenResponse.
func ToProtoRefreshTokenResponse(accessToken, refreshToken string) *auth.RefreshTokenResponse {
	return &auth.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}
}
