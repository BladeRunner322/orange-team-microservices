package clients

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/exercises"
	grpcclient "github.com/BladeRunner322/orange-team-microservices/pkg/grpc/client"
)

// ExercisesClient — gRPC-клиент к Exercises Service.
type ExercisesClient struct {
	conn   *grpc.ClientConn
	client exercises.ExercisesServiceClient
}

// NewExercisesClient создаёт gRPC-клиент к Exercises Service.
//
// Использует общий pkg/grpc/client для соединения:
//   - TLSModeInsecure (Exercises использует self-signed сертификат);
//   - Timeout из аргумента — ограничивает каждый вызов;
//   - автоматическое прокидывание user_id из context в metadata.
//
// TLS-режим захардкожен внутри, потому что это деталь Exercises,
// а не Gateway — вызывающему коду не нужно об этом думать.
func NewExercisesClient(ctx context.Context, addr string, timeout time.Duration) (*ExercisesClient, error) {
	conn, err := grpcclient.New(ctx, grpcclient.Config{
		Target:  addr,
		TLSMode: grpcclient.TLSModeInsecure,
		Timeout: timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("create exercises grpc client: %w", err)
	}

	return &ExercisesClient{
		conn:   conn,
		client: exercises.NewExercisesServiceClient(conn),
	}, nil
}

// CreateExercise создаёт упражнение.
func (c *ExercisesClient) CreateExercise(ctx context.Context, req *exercises.CreateExerciseRequest) (*exercises.Exercise, error) {
	return c.client.CreateExercise(ctx, req)
}

// GetExercise возвращает упражнение по id.
func (c *ExercisesClient) GetExercise(ctx context.Context, id string) (*exercises.Exercise, error) {
	return c.client.GetExercise(ctx, &exercises.GetExerciseRequest{Id: id})
}

// GetExercises возвращает список активных упражнений.
func (c *ExercisesClient) GetExercises(ctx context.Context) (*exercises.GetExercisesResponse, error) {
	return c.client.GetExercises(ctx, &exercises.GetExercisesRequest{})
}

// PatchExercise применяет патч к упражнению.
func (c *ExercisesClient) PatchExercise(ctx context.Context, req *exercises.PatchExerciseRequest) (*exercises.Exercise, error) {
	return c.client.PatchExercise(ctx, req)
}

// DeleteExercise помечает упражнение удалённым.
func (c *ExercisesClient) DeleteExercise(ctx context.Context, id string) error {
	_, err := c.client.DeleteExercise(ctx, &exercises.DeleteExerciseRequest{Id: id})
	return err
}

// Close закрывает gRPC-соединение.
func (c *ExercisesClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
