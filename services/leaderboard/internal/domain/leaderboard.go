package domain

import (
	"time"

	"github.com/google/uuid"
)

// Entry — строка рейтинга с местом пользователя и суммой очков.
type Entry struct {
	Rank   int32
	UserID uuid.UUID
	Score  int64
}

// Leaderboard — снимок рейтинга за календарный период.
type Leaderboard struct {
	Period    Period
	UpdatedAt time.Time
	Entries   []Entry
}
