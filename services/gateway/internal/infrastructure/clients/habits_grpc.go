package clients

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/habits"
	grpcclient "github.com/BladeRunner322/orange-team-microservices/pkg/grpc/client"
)

// HabitsClient — gRPC-клиент к Habits Service.
type HabitsClient struct {
	conn   *grpc.ClientConn
	client habits.HabitsServiceClient
}

// NewHabitsClient создаёт gRPC-клиент к Habits Service.
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
func NewHabitsClient(ctx context.Context, cfg grpcclient.Config) (*HabitsClient, error) {
	conn, err := grpcclient.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create habits grpc client: %w", err)
	}

	return &HabitsClient{
		conn:   conn,
		client: habits.NewHabitsServiceClient(conn),
	}, nil
}

// CreateHabit создаёт привычку.
func (c *HabitsClient) CreateHabit(ctx context.Context, req *habits.CreateHabitRequest) (*habits.Habit, error) {
	return c.client.CreateHabit(ctx, req)
}

// GetHabit возвращает привычку по id.
func (c *HabitsClient) GetHabit(ctx context.Context, id string) (*habits.Habit, error) {
	return c.client.GetHabit(ctx, &habits.GetHabitRequest{Id: id})
}

// GetHabits возвращает список привычек пользователя.
func (c *HabitsClient) GetHabits(ctx context.Context) (*habits.GetHabitsResponse, error) {
	return c.client.GetHabits(ctx, &habits.GetHabitsRequest{})
}

// PatchHabit применяет патч к привычке.
func (c *HabitsClient) PatchHabit(ctx context.Context, req *habits.PatchHabitRequest) (*habits.Habit, error) {
	return c.client.PatchHabit(ctx, req)
}

// DeleteHabit помечает привычку удалённой.
func (c *HabitsClient) DeleteHabit(ctx context.Context, id string) error {
	_, err := c.client.DeleteHabit(ctx, &habits.DeleteHabitRequest{Id: id})
	return err
}

// CompleteHabit отмечает выполнение привычки на дату.
func (c *HabitsClient) CompleteHabit(ctx context.Context, req *habits.CompleteHabitRequest) error {
	_, err := c.client.CompleteHabit(ctx, req)
	return err
}

// UncompleteHabit снимает отметку выполнения.
func (c *HabitsClient) UncompleteHabit(ctx context.Context, req *habits.UncompleteHabitRequest) error {
	_, err := c.client.UncompleteHabit(ctx, req)
	return err
}

// GetCompletions возвращает историю выполнений за период.
func (c *HabitsClient) GetCompletions(ctx context.Context, req *habits.GetCompletionsRequest) (*habits.GetCompletionsResponse, error) {
	return c.client.GetCompletions(ctx, req)
}

// IsHealthy проверяет, что gRPC-соединение с Habits не в фатальном состоянии.
//
// Используется в /ready Gateway. Idle и Connecting считаются «здоровыми»,
// потому что gRPC ленив: соединение открывается при первом RPC, а не при старте.
func (c *HabitsClient) IsHealthy(ctx context.Context) error {
	if c.conn == nil {
		return errors.New("habits connection is nil")
	}

	state := c.conn.GetState()
	switch state {
	case connectivity.Ready, connectivity.Idle, connectivity.Connecting:
		return nil
	default:
		return fmt.Errorf("habits connection state: %s", state)
	}
}

// Close закрывает gRPC-соединение.
func (c *HabitsClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
