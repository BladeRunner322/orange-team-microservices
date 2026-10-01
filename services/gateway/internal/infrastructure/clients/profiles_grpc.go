package clients

import (
	"context"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/profiles"
	grpcclient "github.com/BladeRunner322/orange-team-microservices/pkg/grpc/client"
	"google.golang.org/grpc"
)

// ProfilesClient — gRPC-клиент к Profiles Service.
type ProfilesClient struct {
	conn   *grpc.ClientConn
	client profiles.ProfilesServiceClient
}

// NewProfilesClient создаёт gRPC-клиент к Profiles Service.
func NewProfilesClient(ctx context.Context, addr string) (*ProfilesClient, error) {
	conn, err := grpcclient.New(ctx, grpcclient.Config{
		Target:  addr,
		TLSMode: grpcclient.TLSModeInsecure,
	})

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

// Close закрывает gRPC-соединение.
func (c *ProfilesClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
