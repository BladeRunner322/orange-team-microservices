package exercises

import (
	exercisespb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/exercises"
)

// ===== Преобразования из proto в HTTP-DTO =====

// protoToExerciseResponse преобразует proto-ответ в HTTP-DTO.
func protoToExerciseResponse(p *exercisespb.Exercise) ExerciseResponse {
	resp := ExerciseResponse{
		ID:          p.Id,
		Name:        p.Name,
		Description: p.Description,
		Difficulty:  p.Difficulty,
		Type:        p.Type,
		IsDeleted:   p.IsDeleted,
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

// protoToExerciseListResponse преобразует список proto-упражнений в список HTTP-DTO.
//
// Возвращает make([]ExerciseResponse, 0) при пустом списке — чтобы в JSON
// был [], а не null.
func protoToExerciseListResponse(p *exercisespb.GetExercisesResponse) []ExerciseResponse {
	result := make([]ExerciseResponse, 0, len(p.Exercises))
	for _, ex := range p.Exercises {
		result = append(result, protoToExerciseResponse(ex))
	}
	return result
}

// ===== Преобразования из HTTP-DTO в proto =====

// createRequestToProto преобразует HTTP-DTO создания в proto-запрос.
func createRequestToProto(req CreateExerciseRequest) *exercisespb.CreateExerciseRequest {
	return &exercisespb.CreateExerciseRequest{
		Name:        req.Name,
		Description: req.Description,
		Difficulty:  req.Difficulty,
		Type:        req.Type,
	}
}

// patchRequestToProto преобразует HTTP-DTO патча в proto-запрос.
//
// ID приходит отдельно (из chi URLParam), потому что в теле запроса его нет.
func patchRequestToProto(id string, req PatchExerciseRequest) *exercisespb.PatchExerciseRequest {
	return &exercisespb.PatchExerciseRequest{
		Id:          id,
		Name:        req.Name,
		Description: req.Description,
		Difficulty:  req.Difficulty,
	}
}
