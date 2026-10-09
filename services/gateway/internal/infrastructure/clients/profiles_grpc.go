package clients

import (
	"context"
	"errors"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/profiles"
	grpcclient "github.com/BladeRunner322/orange-team-microservices/pkg/grpc/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// ProfilesClient — gRPC-клиент к Profiles Service.
type ProfilesClient struct {
	conn   *grpc.ClientConn
	client profiles.ProfilesServiceClient
}

// NewProfilesClient создаёт gRPC-клиент к Profiles Service.
//
// Принимает готовый grpcclient.Config — bootstrap заполняет его
// из config.Config Gateway (Target + TLSMode + Timeout + CB + Retry).
//
// Использует общий pkg/grpc/client для соединения:
//   - единый TLS-режим (GRPC_CLIENT_TLS_MODE);
//   - таймаут на каждый вызов;
//   - Circuit Breaker (ADR-020);
//   - Retry для read-only методов (whitelist в config);
//   - автоматическое прокидывание user_id из context в metadata.
func NewProfilesClient(ctx context.Context, cfg grpcclient.Config) (*ProfilesClient, error) {
	conn, err := grpcclient.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create profiles grpc client: %w", err)
	}

	return &ProfilesClient{
		conn:   conn,
		client: profiles.NewProfilesServiceClient(conn),
	}, nil
}

// GetMyProfile возвращает профиль текущего пользователя.
func (c *ProfilesClient) GetMyProfile(ctx context.Context) (*profiles.UserProfile, error) {
	req := &profiles.GetMyProfileRequest{}
	return c.client.GetMyProfile(ctx, req)
}

// PatchMyProfile применяет патч к профилю текущего пользователя.
func (c *ProfilesClient) PatchMyProfile(ctx context.Context, req *profiles.PatchMyProfileRequest) (*profiles.UserProfile, error) {
	return c.client.PatchMyProfile(ctx, req)
}

// DeleteMyProfile удаляет профиль текущего пользователя.
func (c *ProfilesClient) DeleteMyProfile(ctx context.Context) error {
	req := &profiles.DeleteMyProfileRequest{}
	_, err := c.client.DeleteMyProfile(ctx, req)

	return err
}

// IsHealthy проверяет, что gRPC-соединение с Profiles не в фатальном состоянии.
//
// Используется в /ready Gateway. Idle и Connecting считаются «здоровыми»,
// потому что gRPC ленив: соединение открывается при первом RPC, а не при старте.
func (c *ProfilesClient) IsHealthy(ctx context.Context) error {
	if c.conn == nil {
		return errors.New("profiles connection is nil")
	}

	state := c.conn.GetState()
	switch state {
	case connectivity.Ready, connectivity.Idle, connectivity.Connecting:
		return nil
	default:
		return fmt.Errorf("profiles connection state: %s", state)
	}
}

// Close закрывает gRPC-соединение.
func (c *ProfilesClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
