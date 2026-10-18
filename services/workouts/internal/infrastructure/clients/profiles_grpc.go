package clients

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/grpc/client"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type ProfilesClient struct {
	conn *grpc.ClientConn
	// TODO: client profiles.ProfilesServiceClient (после генерации из proto)
}

func NewProfilesClient(ctx context.Context, cfg client.Config) (*ProfilesClient, error) {
	// TODO: grpcclient.New(ctx, cfg) + init profiles client
	return &ProfilesClient{}, nil
}

func (c *ProfilesClient) GetProfile(ctx context.Context, userID uuid.UUID) (ports.UserProfile, error) {
	// TODO: c.client.GetProfile(ctx, &profiles.GetProfileRequest{UserId: userID.String()})
	//       → маппинг в ports.UserProfile
	return ports.UserProfile{}, nil
}

func (c *ProfilesClient) IsHealthy(ctx context.Context) error {
	// TODO: проверка состояния conn
	return nil
}

func (c *ProfilesClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
