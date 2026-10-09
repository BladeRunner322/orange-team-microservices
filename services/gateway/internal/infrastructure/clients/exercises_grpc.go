package clients

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

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
// Принимает готовый grpcclient.Config — bootstrap заполняет его
// из config.Config Gateway (Target + TLSMode + Timeout + CB + Retry).
//
// Использует общий pkg/grpc/client для соединения:
//   - единый TLS-режим (GRPC_CLIENT_TLS_MODE);
//   - таймаут на каждый вызов;
//   - Circuit Breaker (ADR-020);
//   - Retry для read-only методов (whitelist в config);
//   - автоматическое прокидывание user_id из context в metadata.
func NewExercisesClient(ctx context.Context, cfg grpcclient.Config) (*ExercisesClient, error) {
	conn, err := grpcclient.New(ctx, cfg)
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

// IsHealthy проверяет, что gRPC-соединение с Exercises не в фатальном состоянии.
//
// Используется в /ready Gateway. Idle и Connecting считаются «здоровыми»,
// потому что gRPC ленив: соединение открывается при первом RPC, а не при старте.
func (c *ExercisesClient) IsHealthy(ctx context.Context) error {
	if c.conn == nil {
		return errors.New("exercises connection is nil")
	}

	state := c.conn.GetState()
	switch state {
	case connectivity.Ready, connectivity.Idle, connectivity.Connecting:
		return nil
	default:
		return fmt.Errorf("exercises connection state: %s", state)
	}
}

// Close закрывает gRPC-соединение.
func (c *ExercisesClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
