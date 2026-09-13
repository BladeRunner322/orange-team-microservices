package clients

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/auth"
	grpcclient "github.com/BladeRunner322/orange-team-microservices/pkg/grpc/client"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/application/ports"
)

type AuthClient struct {
	conn   *grpc.ClientConn
	client auth.AuthServiceClient
}

// NewAuthClient создаёт gRPC-клиент к Auth Service.
//
// Использует общий pkg/grpc/client — единый TLS-режим (TLSModeInsecure,
// так как Auth использует self-signed сертификат) и автоматическое
// прокидывание user_id в metadata (для будущих вызовов).
func NewAuthClient(ctx context.Context, addr string) (*AuthClient, error) {
	conn, err := grpcclient.New(ctx, grpcclient.Config{
		Target:  addr,
		TLSMode: grpcclient.TLSModeInsecure,
	})
	if err != nil {
		return nil, fmt.Errorf("create auth grpc client: %w", err)
	}

	return &AuthClient{
		conn:   conn,
		client: auth.NewAuthServiceClient(conn),
	}, nil
}

func (c *AuthClient) ValidateToken(ctx context.Context, token string) (ports.UserInfo, error) {
	resp, err := c.client.ValidateToken(ctx, &auth.ValidateTokenRequest{Token: token})
	if err != nil {
		return ports.UserInfo{}, err
	}
	if !resp.Valid {
		return ports.UserInfo{}, nil
	}
	return ports.UserInfo{
		UserID: resp.UserId,
		Role:   resp.Role,
	}, nil
}

func (c *AuthClient) Register(ctx context.Context, email, password, fullName string) (*auth.RegisterResponse, error) {
	req := &auth.RegisterRequest{
		Email:    email,
		Password: password,
		FullName: fullName,
	}
	return c.client.Register(ctx, req)
}

func (c *AuthClient) Login(ctx context.Context, email, password string) (*auth.LoginResponse, error) {
	req := &auth.LoginRequest{
		Email:    email,
		Password: password,
	}
	return c.client.Login(ctx, req)
}

func (c *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (*auth.RefreshTokenResponse, error) {
	req := &auth.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}
	return c.client.RefreshToken(ctx, req)
}

func (c *AuthClient) Logout(ctx context.Context, refreshToken string) error {
	req := &auth.LogoutRequest{
		RefreshToken: refreshToken,
	}
	_, err := c.client.Logout(ctx, req)
	return err
}

func (c *AuthClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
