package usecases

import (
	"context"
	"errors"

	"github.com/BladeRunner322/orange-team-microservices/services/leaderboard/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/leaderboard/internal/domain"
)

type GetLeaderboard struct {
	store ports.Repository
}

func NewGetLeaderboard(store ports.Repository) *GetLeaderboard {
	return &GetLeaderboard{store: store}
}

// Execute возвращает готовый топ за текущий период указанного вида.
func (uc *GetLeaderboard) Execute(ctx context.Context, kind domain.PeriodKind) (domain.Leaderboard, error) {
	// TODO: Вернуть готовый топ за текущий период или ошибку, если снимка ещё нет.
	return domain.Leaderboard{}, errors.ErrUnsupported
}
