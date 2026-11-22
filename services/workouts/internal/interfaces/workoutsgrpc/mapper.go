package workoutsgrpc

import (
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/workouts"
	"github.com/BladeRunner322/orange-team-microservices/pkg/nullable"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ===== Хелперы для отдельных полей =====

// domainStatusFromProto конвертирует *string из proto в *domain.Status.
func domainStatusFromProto(raw *string) (*domain.Status, error) {
	if raw == nil {
		return nil, nil
	}

	status, err := domain.NewStatus(*raw)
	if err != nil {
		return nil, fmt.Errorf("validate status: %w", err)
	}

	return &status, nil
}

// domainIntensityFromProto конвертирует NullableInt32 в nullable.Nullable[domain.Intensity].
func domainIntensityFromProto(wrapper *workouts.NullableInt32) (nullable.Nullable[domain.Intensity], error) {
	if wrapper == nil {
		return nullable.Nullable[domain.Intensity]{}, nil
	}

	if wrapper.Value == nil {
		return nullable.Nullable[domain.Intensity]{Set: true}, nil
	}

	intensity, err := domain.NewIntensity(int(*wrapper.Value))
	if err != nil {
		return nullable.Nullable[domain.Intensity]{}, fmt.Errorf("validate intensity: %w", err)
	}

	return nullable.Nullable[domain.Intensity]{Set: true, Value: &intensity}, nil
}

// domainWeightFromProto конвертирует *int32 из proto в *domain.Weight.
func domainWeightFromProto(raw *int32) (*domain.Weight, error) {
	if raw == nil {
		return nil, nil
	}

	weight, err := domain.NewWeight(int(*raw))
	if err != nil {
		return nil, fmt.Errorf("validate weight: %w", err)
	}

	return &weight, nil
}

// domainSetsFromProto конвертирует *int32 из proto в *domain.Sets.
func domainSetsFromProto(raw *int32) (*domain.Sets, error) {
	if raw == nil {
		return nil, nil
	}

	sets, err := domain.NewSets(int(*raw))
	if err != nil {
		return nil, fmt.Errorf("validate sets: %w", err)
	}

	return &sets, nil
}

// domainRepsFromProto конвертирует *int32 из proto в *domain.Reps.
func domainRepsFromProto(raw *int32) (*domain.Reps, error) {
	if raw == nil {
		return nil, nil
	}

	reps, err := domain.NewReps(int(*raw))
	if err != nil {
		return nil, fmt.Errorf("validate reps: %w", err)
	}

	return &reps, nil
}

// domainDurationFromProto конвертирует *int32 из proto в *domain.Duration.
func domainDurationFromProto(raw *int32) (*domain.Duration, error) {
	if raw == nil {
		return nil, nil
	}

	duration, err := domain.NewDuration(int(*raw))
	if err != nil {
		return nil, fmt.Errorf("validate duration: %w", err)
	}

	return &duration, nil
}

// domainBoolFromProto конвертирует *bool из proto в *bool.
// Возвращаем копию, чтобы не алиасить proto-объект.
func domainBoolFromProto(raw *bool) *bool {
	if raw == nil {
		return nil
	}

	b := *raw
	return &b
}

// domainUUIDFromString парсит строку (из proto) в uuid.UUID (domain).
func domainUUIDFromString(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse id: %w", err)
	}
	return id, nil
}

// ===== Преобразования из protobuf в параметры use case =====

// domainWorkoutPatchFromProto преобразует pb.PatchWorkoutRequest в domain.WorkoutPatch.
func domainWorkoutPatchFromProto(req *workouts.PatchWorkoutRequest) (domain.WorkoutPatch, error) {
	status, err := domainStatusFromProto(req.Status)
	if err != nil {
		return domain.WorkoutPatch{}, err
	}

	intensity, err := domainIntensityFromProto(req.Intensity)
	if err != nil {
		return domain.WorkoutPatch{}, err
	}

	return domain.NewWorkoutPatch(status, intensity), nil
}

// domainWorkoutExercisePatchFromProto преобразует pb.PatchWorkoutExerciseRequest в domain.WorkoutExercisePatch.
func domainWorkoutExercisePatchFromProto(req *workouts.PatchWorkoutExerciseRequest) (domain.WorkoutExercisePatch, error) {
	weight, err := domainWeightFromProto(req.Weight)
	if err != nil {
		return domain.WorkoutExercisePatch{}, err
	}

	sets, err := domainSetsFromProto(req.Sets)
	if err != nil {
		return domain.WorkoutExercisePatch{}, err
	}

	reps, err := domainRepsFromProto(req.Reps)
	if err != nil {
		return domain.WorkoutExercisePatch{}, err
	}

	duration, err := domainDurationFromProto(req.Duration)
	if err != nil {
		return domain.WorkoutExercisePatch{}, err
	}

	completed := domainBoolFromProto(req.Completed)

	return domain.NewWorkoutExercisePatch(weight, sets, reps, duration, completed), nil
}

// ===== Преобразования из доменных объектов в protobuf =====

// protoWorkoutFromDomain преобразует domain.Workout в pb.Workout.
func protoWorkoutFromDomain(w domain.Workout) *workouts.Workout {
	var startedAt, completedAt, updatedAt *timestamppb.Timestamp

	if t := w.StartedAt(); t != nil {
		startedAt = timestamppb.New(*t)
	}
	if t := w.CompletedAt(); t != nil {
		completedAt = timestamppb.New(*t)
	}
	if t := w.UpdatedAt(); t != nil {
		updatedAt = timestamppb.New(*t)
	}

	var intensity *int32
	if i := w.Intensity(); i != nil {
		val := int32(i.Int())
		intensity = &val
	}

	return &workouts.Workout{
		Id:                       w.ID().String(),
		Version:                  int32(w.Version()),
		UserId:                   w.UserID().String(),
		Status:                   w.Status().String(),
		StartedAt:                startedAt,
		CompletedAt:              completedAt,
		WorkoutScore:             int32(w.WorkoutScore()),
		Intensity:                intensity,
		PersonalScoreCoefficient: int32(w.PersonalScoreCoefficient()),
		CreatedAt:                timestamppb.New(w.CreatedAt()),
		UpdatedAt:                updatedAt,
		IsDeleted:                w.IsDeleted(),
	}
}

// protoWorkoutsFromDomain преобразует список тренировок в список proto-объектов.
func protoWorkoutsFromDomain(ws []domain.Workout) []*workouts.Workout {
	result := make([]*workouts.Workout, 0, len(ws))
	for _, w := range ws {
		result = append(result, protoWorkoutFromDomain(w))
	}
	return result
}

// protoWorkoutExerciseFromDomain преобразует domain.WorkoutExercise в pb.WorkoutExercise.
func protoWorkoutExerciseFromDomain(e domain.WorkoutExercise) *workouts.WorkoutExercise {
	var weight, sets, reps, duration *int32

	if w := e.Weight(); w != nil {
		val := int32(w.Int())
		weight = &val
	}
	if s := e.Sets(); s != nil {
		val := int32(s.Int())
		sets = &val
	}
	if r := e.Reps(); r != nil {
		val := int32(r.Int())
		reps = &val
	}
	if d := e.Duration(); d != nil {
		val := int32(d.Int())
		duration = &val
	}

	var updatedAt *timestamppb.Timestamp
	if t := e.UpdatedAt(); t != nil {
		updatedAt = timestamppb.New(*t)
	}

	return &workouts.WorkoutExercise{
		Id:           e.ID().String(),
		Version:      int32(e.Version()),
		WorkoutId:    e.WorkoutID().String(),
		ExerciseId:   e.ExerciseID().String(),
		Difficulty:   int32(e.Difficulty().Int()),
		Weight:       weight,
		Sets:         sets,
		Reps:         reps,
		Duration:     duration,
		Completed:    e.Completed(),
		ExerciseLoad: int32(e.ExerciseLoad()),
		CreatedAt:    timestamppb.New(e.CreatedAt()),
		UpdatedAt:    updatedAt,
	}
}

// protoWorkoutExercisesFromDomain преобразует список упражнений в список proto-объектов.
func protoWorkoutExercisesFromDomain(es []domain.WorkoutExercise) []*workouts.WorkoutExercise {
	result := make([]*workouts.WorkoutExercise, 0, len(es))
	for _, e := range es {
		result = append(result, protoWorkoutExerciseFromDomain(e))
	}
	return result
}
