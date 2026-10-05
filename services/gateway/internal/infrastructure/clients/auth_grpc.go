package clients

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

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
// Использует общий pkg/grpc/client для соединения:
//   - TLSModeInsecure (Auth использует self-signed сертификат);
//   - Timeout из аргумента — ограничивает каждый вызов;
//   - автоматическое прокидывание user_id из context в metadata.
//
// TLS-режим приходит из config (GRPC_CLIENT_TLS_MODE).
// "insecure" — для self-signed сертификатов в dev/staging;
// "verify" — для прода; "disabled" — только для локальной разработки.
func NewAuthClient(ctx context.Context, addr string, timeout time.Duration, tlsMode string) (*AuthClient, error) {
	conn, err := grpcclient.New(ctx, grpcclient.Config{
		Target:  addr,
		TLSMode: grpcclient.TLSMode(tlsMode),
		Timeout: timeout,
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

// IsHealthy проверяет, что gRPC-соединение с Auth не в фатальном состоянии.
//
// Используется в /ready Gateway. Idle и Connecting считаются «здоровыми»,
// потому что gRPC ленив: соединение открывается при первом RPC, а не при старте.
func (c *AuthClient) IsHealthy(ctx context.Context) error {
	if c.conn == nil {
		return errors.New("auth connection is nil")
	}

	state := c.conn.GetState()
	switch state {
	case connectivity.Ready, connectivity.Idle, connectivity.Connecting:
		return nil
	default:
		return fmt.Errorf("auth connection state: %s", state)
	}
}

func (c *AuthClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
