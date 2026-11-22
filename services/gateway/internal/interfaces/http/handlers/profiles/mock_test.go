package profiles

import (
	"context"

	profilespb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/profiles"
	workoutspb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/workouts"
)

// mockProfilesClient — реализация ports.ProfilesClientInterface для тестов.
type mockProfilesClient struct {
	getMyProfileFunc    func(ctx context.Context) (*profilespb.UserProfile, error)
	patchMyProfileFunc  func(ctx context.Context, patch *profilespb.PatchMyProfileRequest) (*profilespb.UserProfile, error)
	deleteMyProfileFunc func(ctx context.Context) error
}

func (m *mockProfilesClient) GetMyProfile(ctx context.Context) (*profilespb.UserProfile, error) {
	if m.getMyProfileFunc != nil {
		return m.getMyProfileFunc(ctx)
	}
	return &profilespb.UserProfile{
		UserId:           "test-user-id",
		ProfileCompleted: false,
	}, nil
}

func (m *mockProfilesClient) PatchMyProfile(ctx context.Context, patch *profilespb.PatchMyProfileRequest) (*profilespb.UserProfile, error) {
	if m.patchMyProfileFunc != nil {
		return m.patchMyProfileFunc(ctx, patch)
	}
	return &profilespb.UserProfile{
		UserId:           "test-user-id",
		ProfileCompleted: false,
	}, nil
}

func (m *mockProfilesClient) DeleteMyProfile(ctx context.Context) error {
	if m.deleteMyProfileFunc != nil {
		return m.deleteMyProfileFunc(ctx)
	}
	return nil
}

func (m *mockProfilesClient) Close() {}

// mockWorkoutScoreClient подставляет результат запроса очков для GET /users/me.
type mockWorkoutScoreClient struct {
	getUserScoreFunc func(ctx context.Context) (*workoutspb.GetUserScoreResponse, error)
}

func (m *mockWorkoutScoreClient) GetUserScore(ctx context.Context) (*workoutspb.GetUserScoreResponse, error) {
	return m.getUserScoreFunc(ctx)
}
