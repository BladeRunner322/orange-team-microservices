package authgrpc

import (
	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/auth"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

// ===== Преобразования из protobuf в параметры use case =====

// registerParamsFromProto преобразует pb.RegisterRequest в параметры для use case.
func registerParamsFromProto(req *auth.RegisterRequest) (email, password, fullName string) {
	return req.Email, req.Password, req.FullName
}

// loginParamsFromProto преобразует pb.LoginRequest в параметры для use case.
func loginParamsFromProto(req *auth.LoginRequest) (email, password string) {
	return req.Email, req.Password
}

// tokenFromProto преобразует pb.ValidateTokenRequest в параметр для use case.
func tokenFromProto(req *auth.ValidateTokenRequest) string {
	return req.Token
}

// ===== Преобразования из доменных объектов в protobuf =====

// protoRegisterResponseFromDomain преобразует доменного пользователя в pb.RegisterResponse.
func protoRegisterResponseFromDomain(user domain.User) *auth.RegisterResponse {
	return &auth.RegisterResponse{
		Id:       user.ID().String(),
		Email:    user.Email().String(),
		FullName: user.FullName().String(),
	}
}

// protoLoginResponseFromTokens преобразует токен в pb.LoginResponse.
func protoLoginResponseFromTokens(accessToken, refreshToken string) *auth.LoginResponse {
	return &auth.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}
}

// protoValidateTokenResponseFromUserInfo преобразует результат валидации в pb.ValidateTokenResponse.
func protoValidateTokenResponseFromUserInfo(info ports.UserInfo, valid bool) *auth.ValidateTokenResponse {
	return &auth.ValidateTokenResponse{
		UserId: info.UserID,
		Valid:  valid,
		Role:   info.Role,
	}
}

// protoRefreshTokenResponseFromTokens преобразует новую пару токенов в pb.RefreshTokenResponse.
func protoRefreshTokenResponseFromTokens(accessToken, refreshToken string) *auth.RefreshTokenResponse {
	return &auth.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}
}
