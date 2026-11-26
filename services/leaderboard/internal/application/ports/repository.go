package ports

import "github.com/BladeRunner322/orange-team-microservices/services/leaderboard/internal/domain"

// Repository — контракт хранения готовых топов.
type Repository interface {
	Get(kind domain.PeriodKind) (domain.Leaderboard, bool)
	Save(kind domain.PeriodKind, top domain.Leaderboard)
}
