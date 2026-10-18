package clients

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/grpc/client"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type ExercisesClient struct {
	conn *grpc.ClientConn
	// TODO: client exercises.ExercisesServiceClient
}

func NewExercisesClient(ctx context.Context, cfg client.Config) (*ExercisesClient, error) {
	// TODO: grpcclient.New(ctx, cfg) + init exercises client
	return &ExercisesClient{}, nil
}

func (c *ExercisesClient) GetExercise(ctx context.Context, id uuid.UUID) (ports.ExerciseInfo, error) {
	// TODO: c.client.GetExercise(ctx, &exercises.GetExerciseRequest{Id: id.String()})
	//       → маппинг в ports.ExerciseInfo
	return ports.ExerciseInfo{}, nil
}

func (c *ExercisesClient) IsHealthy(ctx context.Context) error {
	// TODO
	return nil
}

func (c *ExercisesClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
