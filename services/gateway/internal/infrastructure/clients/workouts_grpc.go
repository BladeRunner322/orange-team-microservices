package clients

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/workouts"
	grpcclient "github.com/BladeRunner322/orange-team-microservices/pkg/grpc/client"
)

// WorkoutsClient — gRPC-клиент к сервису Workouts.
type WorkoutsClient struct {
	conn   *grpc.ClientConn
	client workouts.WorkoutsServiceClient
}

// NewWorkoutsClient создаёт gRPC-клиент к сервису Workouts.
//
// Принимает готовый grpcclient.Config — bootstrap заполняет его
// из конфигурации Gateway (адрес, TLS, таймаут, Circuit Breaker и Retry).
//
// Использует общий pkg/grpc/client для соединения:
//   - единый TLS-режим (GRPC_CLIENT_TLS_MODE);
//   - таймаут на каждый вызов;
//   - Circuit Breaker (ADR-020);
//   - Retry для методов чтения из списка в конфигурации;
//   - автоматическая передача user_id из контекста в gRPC metadata.
func NewWorkoutsClient(ctx context.Context, cfg grpcclient.Config) (*WorkoutsClient, error) {
	conn, err := grpcclient.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create workouts grpc client: %w", err)
	}

	return &WorkoutsClient{
		conn:   conn,
		client: workouts.NewWorkoutsServiceClient(conn),
	}, nil
}

// CreateWorkout создаёт тренировку для текущего пользователя.
func (c *WorkoutsClient) CreateWorkout(ctx context.Context) (*workouts.Workout, error) {
	return c.client.CreateWorkout(ctx, &workouts.CreateWorkoutRequest{})
}

// GetWorkout возвращает тренировку по идентификатору.
func (c *WorkoutsClient) GetWorkout(ctx context.Context, id string) (*workouts.Workout, error) {
	return c.client.GetWorkout(ctx, &workouts.GetWorkoutRequest{Id: id})
}

// GetWorkouts возвращает список тренировок текущего пользователя.
func (c *WorkoutsClient) GetWorkouts(ctx context.Context) (*workouts.GetWorkoutsResponse, error) {
	return c.client.GetWorkouts(ctx, &workouts.GetWorkoutsRequest{})
}

// PatchWorkout обновляет переданные поля тренировки.
func (c *WorkoutsClient) PatchWorkout(ctx context.Context, req *workouts.PatchWorkoutRequest) (*workouts.Workout, error) {
	return c.client.PatchWorkout(ctx, req)
}

// DeleteWorkout помечает тренировку удалённой.
func (c *WorkoutsClient) DeleteWorkout(ctx context.Context, id string) error {
	_, err := c.client.DeleteWorkout(ctx, &workouts.DeleteWorkoutRequest{Id: id})
	return err
}

// CreateWorkoutExercise добавляет упражнение в тренировку.
func (c *WorkoutsClient) CreateWorkoutExercise(ctx context.Context, req *workouts.CreateWorkoutExerciseRequest) (*workouts.WorkoutExercise, error) {
	return c.client.CreateWorkoutExercise(ctx, req)
}

// GetWorkoutExercises возвращает список упражнений указанной тренировки.
func (c *WorkoutsClient) GetWorkoutExercises(ctx context.Context, workoutID string) (*workouts.GetWorkoutExercisesResponse, error) {
	return c.client.GetWorkoutExercises(ctx, &workouts.GetWorkoutExercisesRequest{WorkoutId: workoutID})
}

// PatchWorkoutExercise обновляет переданные поля упражнения в тренировке.
func (c *WorkoutsClient) PatchWorkoutExercise(ctx context.Context, req *workouts.PatchWorkoutExerciseRequest) (*workouts.WorkoutExercise, error) {
	return c.client.PatchWorkoutExercise(ctx, req)
}

// DeleteWorkoutExercise удаляет упражнение из тренировки.
func (c *WorkoutsClient) DeleteWorkoutExercise(ctx context.Context, req *workouts.DeleteWorkoutExerciseRequest) error {
	_, err := c.client.DeleteWorkoutExercise(ctx, req)
	return err
}

// GetUserScore возвращает суммарный балл за тренировки текущего пользователя.
func (c *WorkoutsClient) GetUserScore(ctx context.Context) (*workouts.GetUserScoreResponse, error) {
	return c.client.GetUserScore(ctx, &workouts.GetUserScoreRequest{})
}

// Close закрывает gRPC-соединение.
func (c *WorkoutsClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
