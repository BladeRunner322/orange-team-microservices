package clients

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/profiles"
	grpcclient "github.com/BladeRunner322/orange-team-microservices/pkg/grpc/client"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/google/uuid"
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
// из config.Config Workouts (Target + TLSMode + Timeout).
//
// Использует общий pkg/grpc/client для соединения:
//   - единый TLS-режим (GRPC_CLIENT_TLS_MODE);
//   - таймаут на каждый вызов;
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

// GetProfile возвращает профиль пользователя.
// Используется в CreateWorkout для расчёта personal_score_coefficient.
func (c *ProfilesClient) GetProfile(ctx context.Context, userID uuid.UUID) (ports.UserProfile, error) {
	resp, err := c.client.GetProfile(ctx, &profiles.GetProfileRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return ports.UserProfile{}, err
	}

	parsedID, err := uuid.Parse(resp.UserId)
	if err != nil {
		return ports.UserProfile{}, fmt.Errorf("parse user_id from profiles response: %w", err)
	}

	return ports.UserProfile{
		UserID:      parsedID,
		Sex:         optionalString(resp.Sex),
		WeightGrams: optionalWeightGrams(resp.WeightKg),
		HeightCM:    optionalInt(resp.HeightCm),
	}, nil
}

// optionalString возвращает nil для пустой строки (proto не различает
// «не задано» и «пусто», поэтому нулевое значение трактуем как nil).
func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// optionalWeightGrams конвертирует proto double (кг) в *int (граммы).
// 0 → nil (профиль не заполнен, ADR-001).
func optionalWeightGrams(kg float64) *int {
	if kg == 0 {
		return nil
	}
	grams := int(math.Round(kg * 1000))
	return &grams
}

// optionalInt конвертирует proto int32 в *int. 0 → nil.
func optionalInt(v int32) *int {
	if v == 0 {
		return nil
	}
	i := int(v)
	return &i
}

// IsHealthy проверяет, что gRPC-соединение с Profiles не в фатальном состоянии.
//
// Используется в /ready Workouts. Idle и Connecting считаются «здоровыми»,
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
