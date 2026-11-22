package workouts

import workoutspb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/workouts"

// protoToWorkoutResponse преобразует proto-тренировку в HTTP-DTO.
func protoToWorkoutResponse(p *workoutspb.Workout) WorkoutResponse {
	resp := WorkoutResponse{
		ID:                       p.Id,
		Version:                  p.Version,
		UserID:                   p.UserId,
		Status:                   p.Status,
		WorkoutScore:             p.WorkoutScore,
		Intensity:                p.Intensity,
		PersonalScoreCoefficient: p.PersonalScoreCoefficient,
		IsDeleted:                p.IsDeleted,
	}

	if p.StartedAt != nil {
		t := p.StartedAt.AsTime()
		resp.StartedAt = &t
	}
	if p.CompletedAt != nil {
		t := p.CompletedAt.AsTime()
		resp.CompletedAt = &t
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

// protoToWorkoutListResponse преобразует список proto-тренировок в список HTTP-DTO.
func protoToWorkoutListResponse(p *workoutspb.GetWorkoutsResponse) []WorkoutResponse {
	result := make([]WorkoutResponse, 0, len(p.Workouts))
	for _, workout := range p.Workouts {
		result = append(result, protoToWorkoutResponse(workout))
	}
	return result
}

// patchRequestToProto преобразует HTTP-DTO патча в proto-запрос.
func patchRequestToProto(id string, req PatchWorkoutRequest) *workoutspb.PatchWorkoutRequest {
	result := &workoutspb.PatchWorkoutRequest{
		Id:     id,
		Status: req.Status,
	}

	if req.Intensity.Set {
		result.Intensity = &workoutspb.NullableInt32{Value: req.Intensity.Value}
	}

	return result
}
