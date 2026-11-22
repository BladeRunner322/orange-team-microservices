package workouts

import workoutspb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/workouts"

// protoToWorkoutExerciseResponse преобразует proto-упражнение в HTTP-DTO.
func protoToWorkoutExerciseResponse(p *workoutspb.WorkoutExercise) WorkoutExerciseResponse {
	resp := WorkoutExerciseResponse{
		ID:           p.Id,
		Version:      p.Version,
		WorkoutID:    p.WorkoutId,
		ExerciseID:   p.ExerciseId,
		Difficulty:   p.Difficulty,
		Weight:       p.Weight,
		Sets:         p.Sets,
		Reps:         p.Reps,
		Duration:     p.Duration,
		Completed:    p.Completed,
		ExerciseLoad: p.ExerciseLoad,
	}

	if p.CreatedAt != nil {
		resp.CreatedAt = p.CreatedAt.AsTime()
	}
	if p.UpdatedAt != nil {
		t := p.UpdatedAt.AsTime()
		resp.UpdatedAt = &t
	}

	return resp
}

// protoToWorkoutExerciseListResponse преобразует список proto-упражнений в список HTTP-DTO.
func protoToWorkoutExerciseListResponse(p *workoutspb.GetWorkoutExercisesResponse) []WorkoutExerciseResponse {
	result := make([]WorkoutExerciseResponse, 0, len(p.WorkoutExercises))
	for _, exercise := range p.WorkoutExercises {
		result = append(result, protoToWorkoutExerciseResponse(exercise))
	}
	return result
}

// createWorkoutExerciseRequestToProto преобразует HTTP-DTO создания в proto-запрос.
func createWorkoutExerciseRequestToProto(workoutID string, req CreateWorkoutExerciseRequest) *workoutspb.CreateWorkoutExerciseRequest {
	return &workoutspb.CreateWorkoutExerciseRequest{
		WorkoutId:  workoutID,
		ExerciseId: req.ExerciseID,
		Weight:     req.Weight,
		Sets:       req.Sets,
		Reps:       req.Reps,
		Duration:   req.Duration,
		Completed:  req.Completed,
	}
}

// patchWorkoutExerciseRequestToProto преобразует HTTP-DTO патча в proto-запрос.
func patchWorkoutExerciseRequestToProto(workoutID, workoutExerciseID string, req PatchWorkoutExerciseRequest) *workoutspb.PatchWorkoutExerciseRequest {
	return &workoutspb.PatchWorkoutExerciseRequest{
		WorkoutId:         workoutID,
		WorkoutExerciseId: workoutExerciseID,
		Weight:            req.Weight,
		Sets:              req.Sets,
		Reps:              req.Reps,
		Duration:          req.Duration,
		Completed:         req.Completed,
	}
}
