package clients

import (
	"context"
	"errors"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/exercises"
	grpcclient "github.com/BladeRunner322/orange-team-microservices/pkg/grpc/client"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// ExercisesClient — gRPC-клиент к Exercises Service.
type ExercisesClient struct {
	conn   *grpc.ClientConn
	client exercises.ExercisesServiceClient
}

// NewExercisesClient создаёт gRPC-клиент к Exercises Service.
//
// Принимает готовый grpcclient.Config — bootstrap заполняет его
// из config.Config Workouts (Target + TLSMode + Timeout).
//
// Использует общий pkg/grpc/client для соединения:
//   - единый TLS-режим (GRPC_CLIENT_TLS_MODE);
//   - таймаут на каждый вызов;
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

// GetExercise возвращает данные упражнения по id.
// Используется в CreateWorkoutExercise и PatchWorkoutExercise:
//   - type — для валидации полей (weight vs duration);
//   - difficulty — для снапшота в workout_exercises (ADR-008, ADR-025 п.1).
func (c *ExercisesClient) GetExercise(ctx context.Context, id uuid.UUID) (ports.ExerciseInfo, error) {
	resp, err := c.client.GetExercise(ctx, &exercises.GetExerciseRequest{
		Id: id.String(),
	})
	if err != nil {
		return ports.ExerciseInfo{}, err
	}

	parsedID, err := uuid.Parse(resp.Id)
	if err != nil {
		return ports.ExerciseInfo{}, fmt.Errorf("parse exercise_id from exercises response: %w", err)
	}

	return ports.ExerciseInfo{
		ID:         parsedID,
		Type:       resp.Type,
		Difficulty: int(resp.Difficulty),
	}, nil
}

// IsHealthy проверяет, что gRPC-соединение с Exercises не в фатальном состоянии.
//
// Используется в /ready Workouts. Idle и Connecting считаются «здоровыми»,
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
