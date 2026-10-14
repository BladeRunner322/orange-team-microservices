package habitsgrpc

import (
	"time"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/habits"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/usecases"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProtoHabit конвертирует доменную привычку в proto.
// streak и completedToday — вычисляемые поля, передаются отдельно.
func ToProtoHabit(h domain.Habit, streak int, completedToday bool) *habits.Habit {
	// TODO
	return nil
}

// ToProtoHabitWithStreak — из usecase-структуры в proto.
func ToProtoHabitWithStreak(hws usecases.HabitWithStreak) *habits.Habit {
	// TODO
	return nil
}

// ToProtoCompletion конвертирует Completion в proto.
func ToProtoCompletion(c domain.Completion) *habits.Completion {
	// TODO
	return nil
}

// ToDomainDate парсит "YYYY-MM-DD" в time.Time UTC (только дата, без времени).
// Пустая строка → today.
func ToDomainDate(raw string) (time.Time, error) {
	// TODO
	return time.Time{}, nil
}

// ToProtoDate конвертирует time.Time в "YYYY-MM-DD".
func ToProtoDate(t time.Time) string {
	// TODO
	return ""
}

// ToDomainHabitPatch конвертирует PatchHabitRequest в domain.HabitPatch.
// optional-поля proto → *T (nil = не трогать).
func ToDomainHabitPatch(req *habits.PatchHabitRequest) (domain.HabitPatch, error) {
	// TODO
	return domain.HabitPatch{}, nil
}

// ToDomainID парсит id из proto в uuid.UUID.
func ToDomainID(raw string) (uuid.UUID, error) {
	// TODO
	return uuid.Nil, nil
}

var _ = timestamppb.New // заглушка чтобы линтер не ругался на неиспользуемый импорт
