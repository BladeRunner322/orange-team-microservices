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

// statusFromProto конвертирует *string из proto в *domain.Status.
// nil означает «не трогать».
func statusFromProto(raw *string) (*domain.Status, error) {
	if raw == nil {
		return nil, nil
	}

	// TODO: domain.NewStatus(*raw) + wrap err → *Status
	return nil, nil
}

// intensityFromProto конвертирует NullableInt32 в Nullable[domain.Intensity].
// Три состояния: nil (не трогать) / пустой wrapper (сброс в NULL) / значение.
func intensityFromProto(wrapper *workouts.NullableInt32) (nullable.Nullable[domain.Intensity], error) {
	if wrapper == nil {
		return nullable.Nullable[domain.Intensity]{}, nil
	}

	if wrapper.Value == nil {
		return nullable.Nullable[domain.Intensity]{Set: true}, nil
	}

	// TODO: domain.NewIntensity(int(*wrapper.Value)) + wrap err → Set: true, Value: &v
	return nullable.Nullable[domain.Intensity]{Set: true}, nil
}

// weightFromProto конвертирует *int32 из proto в *domain.Weight.
// nil означает «не трогать».
func weightFromProto(raw *int32) (*domain.Weight, error) {
	if raw == nil {
		return nil, nil
	}

	// TODO: domain.NewWeight(int(*raw)) + wrap err
	return nil, nil
}

// setsFromProto конвертирует *int32 из proto в *domain.Sets.
// nil означает «не трогать».
func setsFromProto(raw *int32) (*domain.Sets, error) {
	if raw == nil {
		return nil, nil
	}

	// TODO: domain.NewSets(int(*raw)) + wrap err
	return nil, nil
}

// repsFromProto конвертирует *int32 из proto в *domain.Reps.
// nil означает «не трогать».
func repsFromProto(raw *int32) (*domain.Reps, error) {
	if raw == nil {
		return nil, nil
	}

	// TODO: domain.NewReps(int(*raw)) + wrap err
	return nil, nil
}

// durationFromProto конвертирует *int32 из proto в *domain.Duration.
// nil означает «не трогать».
func durationFromProto(raw *int32) (*domain.Duration, error) {
	if raw == nil {
		return nil, nil
	}

	// TODO: domain.NewDuration(int(*raw)) + wrap err
	return nil, nil
}

// completedFromProto конвертирует *bool из proto в *bool.
// nil означает «не трогать».
func completedFromProto(raw *bool) *bool {
	if raw == nil {
		return nil
	}

	// TODO: вернуть копию *raw (или сам указатель — решить на этапе реализации)
	return nil
}

// ===== Преобразования из protobuf в параметры use case =====

// ToDomainWorkoutPatch преобразует pb.PatchWorkoutRequest в domain.WorkoutPatch.
func ToDomainWorkoutPatch(req *workouts.PatchWorkoutRequest) (domain.WorkoutPatch, error) {
	status, err := statusFromProto(req.Status)
	if err != nil {
		return domain.WorkoutPatch{}, err
	}

	intensity, err := intensityFromProto(req.Intensity)
	if err != nil {
		return domain.WorkoutPatch{}, err
	}

	return domain.NewWorkoutPatch(status, intensity), nil
}

// ToDomainWorkoutExercisePatch преобразует pb.PatchWorkoutExerciseRequest
// в domain.WorkoutExercisePatch.
func ToDomainWorkoutExercisePatch(req *workouts.PatchWorkoutExerciseRequest) (domain.WorkoutExercisePatch, error) {
	weight, err := weightFromProto(req.Weight)
	if err != nil {
		return domain.WorkoutExercisePatch{}, err
	}

	sets, err := setsFromProto(req.Sets)
	if err != nil {
		return domain.WorkoutExercisePatch{}, err
	}

	reps, err := repsFromProto(req.Reps)
	if err != nil {
		return domain.WorkoutExercisePatch{}, err
	}

	duration, err := durationFromProto(req.Duration)
	if err != nil {
		return domain.WorkoutExercisePatch{}, err
	}

	completed := completedFromProto(req.Completed)

	return domain.NewWorkoutExercisePatch(weight, sets, reps, duration, completed), nil
}

// ToDomainID парсит id из proto в uuid.UUID.
func ToDomainID(raw string) (uuid.UUID, error) {
	// TODO: uuid.Parse + wrap err через fmt.Errorf("parse id: %w", err)
	return uuid.Nil, nil
}

// ===== Преобразования из доменных объектов в protobuf =====

// ToProtoWorkout преобразует domain.Workout в pb.Workout.
func ToProtoWorkout(w domain.Workout) *workouts.Workout {
	// TODO:
	//   - id, version, user_id, status — маппинг полей
	//   - started_at / completed_at / created_at / updated_at — *timestamppb.Timestamp
	//     (nil для optional-полей started_at / completed_at, если домен вернул nil)
	//   - intensity — *int32 (nil если Intensity() == nil)
	//   - is_deleted — w.IsDeleted()
	return nil
}

// ToProtoWorkouts преобразует список тренировок в список proto-объектов.
func ToProtoWorkouts(ws []domain.Workout) []*workouts.Workout {
	// TODO: make([]*workouts.Workout, 0, len(ws)) + цикл через ToProtoWorkout
	return nil
}

// ToProtoWorkoutExercise преобразует domain.WorkoutExercise в pb.WorkoutExercise.
func ToProtoWorkoutExercise(e domain.WorkoutExercise) *workouts.WorkoutExercise {
	// TODO:
	//   - id, version, workout_id, exercise_id, difficulty, completed, exercise_load
	//   - weight / sets / reps / duration — *int32 (nil для optional-полей)
	//   - created_at / updated_at — *timestamppb.Timestamp
	return nil
}

// ToProtoWorkoutExercises преобразует список упражнений в список proto-объектов.
func ToProtoWorkoutExercises(es []domain.WorkoutExercise) []*workouts.WorkoutExercise {
	// TODO: make([]*workouts.WorkoutExercise, 0, len(es)) + цикл через ToProtoWorkoutExercise
	return nil
}

var _ = fmt.Sprintf       // placeholder — удалить, когда появятся реальные вызовы fmt
var _ = timestamppb.Now  // placeholder — удалить, когда появится реальный маппинг time.Time