package habits

import (
	habitspb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/habits"
)

// protoToHabitResponse преобразует proto-ответ в HTTP-DTO.
func protoToHabitResponse(p *habitspb.Habit) HabitResponse {
	resp := HabitResponse{
		ID:             p.Id,
		UserID:         p.UserId,
		Name:           p.Name,
		Description:    p.Description,
		Streak:         int(p.Streak),
		CompletedToday: p.CompletedToday,
		IsDeleted:      p.IsDeleted,
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

// protoToHabitListResponse преобразует список proto-привычек в список HTTP-DTO.
func protoToHabitListResponse(p *habitspb.GetHabitsResponse) []HabitResponse {
	result := make([]HabitResponse, 0, len(p.Habits))
	for _, h := range p.Habits {
		result = append(result, protoToHabitResponse(h))
	}
	return result
}

// protoToCompletionListResponse преобразует список proto-отметок в список HTTP-DTO.
func protoToCompletionListResponse(p *habitspb.GetCompletionsResponse) []CompletionResponse {
	result := make([]CompletionResponse, 0, len(p.Completions))
	for _, c := range p.Completions {
		result = append(result, CompletionResponse{
			HabitID:   c.HabitId,
			Date:      c.Date,
			CreatedAt: c.CreatedAt.AsTime(),
		})
	}
	return result
}

// createRequestToProto преобразует HTTP-DTO создания в proto-запрос.
func createRequestToProto(req CreateHabitRequest) *habitspb.CreateHabitRequest {
	return &habitspb.CreateHabitRequest{
		Name:        req.Name,
		Description: req.Description,
	}
}

// patchRequestToProto преобразует HTTP-DTO патча в proto-запрос.
func patchRequestToProto(id string, req PatchHabitRequest) *habitspb.PatchHabitRequest {
	return &habitspb.PatchHabitRequest{
		Id:          id,
		Name:        req.Name,
		Description: req.Description,
	}
}

// completeRequestToProto преобразует HTTP-DTO отметки в proto-запрос.
func completeRequestToProto(id string, req CompleteHabitRequest) *habitspb.CompleteHabitRequest {
	return &habitspb.CompleteHabitRequest{
		Id:   id,
		Date: req.Date,
	}
}

// uncompleteRequestToProto преобразует HTTP-DTO снятия отметки в proto-запрос.
func uncompleteRequestToProto(id string, req UncompleteHabitRequest) *habitspb.UncompleteHabitRequest {
	return &habitspb.UncompleteHabitRequest{
		Id:   id,
		Date: req.Date,
	}
}
