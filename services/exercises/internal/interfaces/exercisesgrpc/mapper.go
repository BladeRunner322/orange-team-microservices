package exercisesgrpc

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/exercises"
	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/domain"
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

// difficultyFromProto конвертирует *int32 из proto в *domain.Difficulty.
// nil означает «не трогать».
func difficultyFromProto(raw *int32) (*domain.Difficulty, error) {
	if raw == nil {
		return nil, nil
	}

	difficulty, err := domain.NewDifficulty(int(*raw))
	if err != nil {
		return nil, fmt.Errorf("validate difficulty: %w", err)
	}

	return &difficulty, nil
}

// patchFromProto конвертирует PatchExerciseRequest в domain.ExercisePatch.
// Поля с nil-значением не трогаются; не-nil проходят валидацию через VO.
func ToDomainPatch(req *exercises.PatchExerciseRequest) (domain.ExercisePatch, error) {
	name, err := nameFromProto(req.Name)
	if err != nil {
		return domain.ExercisePatch{}, err
	}

	description, err := descriptionFromProto(req.Description)
	if err != nil {
		return domain.ExercisePatch{}, err
	}

	difficulty, err := difficultyFromProto(req.Difficulty)
	if err != nil {
		return domain.ExercisePatch{}, err
	}

	return domain.NewExercisePatch(name, description, difficulty), nil
}

// ToDomainPatch конвертирует PatchExerciseRequest в domain.ExercisePatch.
func ToProtoExercise(exercise domain.Exercise) *exercises.Exercise {
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

// ToDomainID парсит id из proto в uuid.UUID.
func ToDomainID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse id: %w", err)
	}
	return id, nil
}
