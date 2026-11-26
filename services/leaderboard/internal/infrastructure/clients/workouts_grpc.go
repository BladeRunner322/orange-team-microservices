package clients

import (
	"context"
	"errors"

	"github.com/BladeRunner322/orange-team-microservices/pkg/grpc/client"
	"github.com/BladeRunner322/orange-team-microservices/services/leaderboard/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/leaderboard/internal/domain"
)

var _ ports.WorkoutsClientInterface = (*WorkoutsClient)(nil)

// WorkoutsClient реализует порт клиента Workouts через gRPC.
type WorkoutsClient struct {
	// TODO: Хранить gRPC-соединение и сгенерированный клиент Workouts.
}

func NewWorkoutsClient(ctx context.Context, cfg client.Config) (*WorkoutsClient, error) {
	// TODO: Подключиться через общий gRPC-клиент с таймаутом, Circuit Breaker и Retry.
	return nil, errors.ErrUnsupported
}

// GetTop запрашивает первые строки рейтинга за заданный период.
func (c *WorkoutsClient) GetTop(ctx context.Context, period domain.Period, limit int32) ([]ports.UserScore, error) {
	// TODO: Вызвать внутренний gRPC-метод Workouts и преобразовать ответ в список очков.
	return nil, errors.ErrUnsupported
}
