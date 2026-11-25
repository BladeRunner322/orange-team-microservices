package exercisesgrpc

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/exercises"
	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// domainNameFromProto конвертирует *string из proto в *domain.Name.
// nil означает «не трогать».
func domainNameFromProto(raw *string) (*domain.Name, error) {
	if raw == nil {
		return nil, nil
	}

	name, err := domain.NewName(*raw)
	if err != nil {
		return nil, fmt.Errorf("validate name: %w", err)
	}

	return &name, nil
}

// domainDescriptionFromProto конвертирует *string из proto в *domain.Description.
// nil означает «не трогать».
func domainDescriptionFromProto(raw *string) (*domain.Description, error) {
	if raw == nil {
		return nil, nil
	}

	description, err := domain.NewDescription(*raw)
	if err != nil {
		return nil, fmt.Errorf("validate description: %w", err)
	}

	return &description, nil
}

// domainDifficultyFromProto конвертирует *int32 из proto в *domain.Difficulty.
// nil означает «не трогать».
func domainDifficultyFromProto(raw *int32) (*domain.Difficulty, error) {
	if raw == nil {
		return nil, nil
	}

	difficulty, err := domain.NewDifficulty(int(*raw))
	if err != nil {
		return nil, fmt.Errorf("validate difficulty: %w", err)
	}

	return &difficulty, nil
}

// domainExercisePatchFromProto конвертирует PatchExerciseRequest в domain.ExercisePatch.
// Поля с nil-значением не трогаются; не-nil проходят валидацию через VO.
func domainExercisePatchFromProto(req *exercises.PatchExerciseRequest) (domain.ExercisePatch, error) {
	name, err := domainNameFromProto(req.Name)
	if err != nil {
		return domain.ExercisePatch{}, err
	}

	description, err := domainDescriptionFromProto(req.Description)
	if err != nil {
		return domain.ExercisePatch{}, err
	}

	difficulty, err := domainDifficultyFromProto(req.Difficulty)
	if err != nil {
		return domain.ExercisePatch{}, err
	}

	return domain.NewExercisePatch(name, description, difficulty), nil
}

// protoExerciseFromDomain конвертирует domain.Exercise в *exercises.Exercise (proto-ответ).
func protoExerciseFromDomain(exercise domain.Exercise) *exercises.Exercise {
	var updatedAt *timestamppb.Timestamp
	if t := exercise.UpdatedAt(); t != nil {
		updatedAt = timestamppb.New(*t)
	}

	return &exercises.Exercise{
		Id:          exercise.ID().String(),
		Name:        exercise.Name().String(),
		Description: exercise.Description().String(),
		Difficulty:  int32(exercise.Difficulty().Int()),
		Type:        exercise.ExerciseType().String(),
		CreatedAt:   timestamppb.New(exercise.CreatedAt()),
		UpdatedAt:   updatedAt,
		IsDeleted:   exercise.IsDeleted(),
	}
}

// domainUUIDFromString парсит id из proto в uuid.UUID.
func domainUUIDFromString(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse id: %w", err)
	}
	return id, nil
}
