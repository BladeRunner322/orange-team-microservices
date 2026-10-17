package habitsgrpc

import (
	"fmt"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/habits"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/usecases"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// nameFromProto конвертирует *string из proto в *domain.Name.
// nil означает «не трогать».
func nameFromProto(raw *string) (*domain.Name, error) {
	if raw == nil {
		return nil, nil
	}

	name, err := domain.NewName(*raw)
	if err != nil {
		return nil, fmt.Errorf("validate name: %w", err)
	}

	return &name, nil
}

// descriptionFromProto конвертирует *string из proto в *domain.Description.
// nil означает «не трогать».
func descriptionFromProto(raw *string) (*domain.Description, error) {
	if raw == nil {
		return nil, nil
	}

	description, err := domain.NewDescription(*raw)
	if err != nil {
		return nil, fmt.Errorf("validate description: %w", err)
	}

	return &description, nil
}

// ToDomainHabitPatch конвертирует PatchHabitRequest в domain.HabitPatch.
// optional-поля proto → *T (nil = не трогать).
func ToDomainHabitPatch(req *habits.PatchHabitRequest) (domain.HabitPatch, error) {
	name, err := nameFromProto(req.Name)
	if err != nil {
		return domain.HabitPatch{}, err
	}

	description, err := descriptionFromProto(req.Description)
	if err != nil {
		return domain.HabitPatch{}, err
	}

	return domain.NewHabitPatch(name, description), nil
}

// ToDomainID парсит id из proto в uuid.UUID.
func ToDomainID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse id: %w", err)
	}

	return id, nil
}

// ToDomainDate парсит "YYYY-MM-DD" в time.Time UTC (только дата, без времени).
// Пустая строка → today (UTC, начало дня).
//
// Используется в CompleteHabit / UncompleteHabit, где пустая дата означает
// «отметить сегодня» (ADR-023 п.4).
func ToDomainDate(raw string) (time.Time, error) {
	if raw == "" {
		now := time.Now().UTC()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
	}

	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date: %w", domain.ErrInvalidDate)
	}

	return t, nil
}

// ToDomainDateOrZero парсит "YYYY-MM-DD" в time.Time UTC.
// Пустая строка → zero time (означает «не задано»).
//
// Используется в GetCompletions для from/to: если обе пустые, usecase
// применит дефолт «последние 30 дней» (ADR-023 п.8).
func ToDomainDateOrZero(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}

	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date: %w", domain.ErrInvalidDate)
	}

	return t, nil
}

// ToProtoDate конвертирует time.Time в "YYYY-MM-DD".
func ToProtoDate(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

// ToProtoHabit конвертирует доменную привычку в proto.
// streak и completedToday — вычисляемые поля, передаются отдельно.
func ToProtoHabit(h domain.Habit, streak int, completedToday bool) *habits.Habit {
	var updatedAt *timestamppb.Timestamp
	if t := h.UpdatedAt(); t != nil {
		updatedAt = timestamppb.New(*t)
	}

	return &habits.Habit{
		Id:             h.ID().String(),
		UserId:         h.UserID().String(),
		Name:           h.Name().String(),
		Description:    h.Description().String(),
		Streak:         int32(streak),
		CompletedToday: completedToday,
		IsDeleted:      h.IsDeleted(),
		CreatedAt:      timestamppb.New(h.CreatedAt()),
		UpdatedAt:      updatedAt,
	}
}

// ToProtoHabitWithStreak — из usecase-структуры в proto.
func ToProtoHabitWithStreak(hws usecases.HabitWithStreak) *habits.Habit {
	return ToProtoHabit(hws.Habit, hws.Streak, hws.CompletedToday)
}

// ToProtoCompletion конвертирует Completion в proto.
func ToProtoCompletion(c domain.Completion) *habits.Completion {
	return &habits.Completion{
		HabitId:   c.HabitID().String(),
		Date:      ToProtoDate(c.Date().Time()),
		CreatedAt: timestamppb.New(c.CreatedAt()),
	}
}