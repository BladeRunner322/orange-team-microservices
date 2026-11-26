package memory_repo

import (
	"sync"

	"github.com/BladeRunner322/orange-team-microservices/services/leaderboard/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/leaderboard/internal/domain"
)

var _ ports.Repository = (*Repository)(nil)

// Repository хранит последние рассчитанные топы в памяти.
type Repository struct {
	mu   sync.RWMutex
	tops map[domain.PeriodKind]domain.Leaderboard
}

func NewRepository() *Repository {
	return &Repository{tops: make(map[domain.PeriodKind]domain.Leaderboard)}
}

// Get возвращает сохранённый топ и признак его наличия.
func (s *Repository) Get(kind domain.PeriodKind) (domain.Leaderboard, bool) {
	// TODO: Прочитать снимок под блокировкой и вернуть копию строк рейтинга.
	return domain.Leaderboard{}, false
}

// Save заменяет сохранённый топ для указанного вида периода.
func (s *Repository) Save(kind domain.PeriodKind, top domain.Leaderboard) {
	// TODO: Сохранить копию снимка под блокировкой.
}
