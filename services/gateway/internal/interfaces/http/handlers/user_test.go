package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/profiles"
	"github.com/BladeRunner322/orange-team-microservices/pkg/nullable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ===== GetUserHandler =====
func TestGetUserHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockProfilesClient{
			getMyProfileFunc: func(ctx context.Context) (*profiles.UserProfile, error) {
				return &profiles.UserProfile{
					UserId:           "user-123",
					Sex:              "male",
					WeightKg:         80.5,
					HeightCm:         180,
					BirthDate:        "1990-05-15",
					ProfileCompleted: true,
				}, nil
			},
		}
		handler := GetUserHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp UserProfileResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "user-123", resp.UserID)
		assert.Equal(t, "male", resp.Sex)
		assert.Equal(t, 80.5, resp.WeightKg)
		assert.Equal(t, int32(180), resp.HeightCm)
		assert.True(t, resp.ProfileCompleted)
	})

	t.Run("gRPC error", func(t *testing.T) {
		mock := &mockProfilesClient{
			getMyProfileFunc: func(ctx context.Context) (*profiles.UserProfile, error) {
				return nil, status.Error(codes.Unauthenticated, "unauthenticated")
			},
		}
		handler := GetUserHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "unauthenticated")
	})
}

// ===== PatchUserHandler =====
func TestPatchUserHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockProfilesClient{
			patchMyProfileFunc: func(ctx context.Context, req *profiles.PatchMyProfileRequest) (*profiles.UserProfile, error) {
				return &profiles.UserProfile{
					UserId:           "user-123",
					Sex:              "female",
					WeightKg:         60.0,
					ProfileCompleted: false,
				}, nil
			},
		}
		handler := PatchUserHandler(mock)

		body, _ := json.Marshal(map[string]any{
			"sex":       "female",
			"weight_kg": 60.0,
		})
		req := httptest.NewRequest(http.MethodPatch, "/users/me", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp UserProfileResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "female", resp.Sex)
		assert.Equal(t, 60.0, resp.WeightKg)
	})

	t.Run("invalid body", func(t *testing.T) {
		mock := &mockProfilesClient{}
		handler := PatchUserHandler(mock)

		req := httptest.NewRequest(http.MethodPatch, "/users/me", bytes.NewReader([]byte("not json")))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "invalid request body")
	})

	t.Run("gRPC error", func(t *testing.T) {
		mock := &mockProfilesClient{
			patchMyProfileFunc: func(ctx context.Context, req *profiles.PatchMyProfileRequest) (*profiles.UserProfile, error) {
				return nil, status.Error(codes.InvalidArgument, "invalid weight")
			},
		}
		handler := PatchUserHandler(mock)

		body, _ := json.Marshal(map[string]any{"weight_kg": 999.0})
		req := httptest.NewRequest(http.MethodPatch, "/users/me", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "invalid weight")
	})

	t.Run("null сбрасывает поле", func(t *testing.T) {
		var captured *profiles.PatchMyProfileRequest

		mock := &mockProfilesClient{
			patchMyProfileFunc: func(ctx context.Context, req *profiles.PatchMyProfileRequest) (*profiles.UserProfile, error) {
				captured = req
				return &profiles.UserProfile{UserId: "user-123"}, nil
			},
		}
		handler := PatchUserHandler(mock)

		body, _ := json.Marshal(map[string]any{"sex": nil})
		req := httptest.NewRequest(http.MethodPatch, "/users/me", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		require.NotNil(t, captured)
		require.NotNil(t, captured.Sex, "wrapper должен быть создан для сброса")
		assert.Nil(t, captured.Sex.Value, "Value должен быть nil — сброс в NULL")
	})

	t.Run("пропущенное поле не трогается", func(t *testing.T) {
		var captured *profiles.PatchMyProfileRequest

		mock := &mockProfilesClient{
			patchMyProfileFunc: func(ctx context.Context, req *profiles.PatchMyProfileRequest) (*profiles.UserProfile, error) {
				captured = req
				return &profiles.UserProfile{UserId: "user-123"}, nil
			},
		}
		handler := PatchUserHandler(mock)

		body, _ := json.Marshal(map[string]any{"weight_kg": 70.0})
		req := httptest.NewRequest(http.MethodPatch, "/users/me", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		require.NotNil(t, captured)
		assert.Nil(t, captured.Sex, "sex не пришёл — wrapper должен быть nil")
		assert.Nil(t, captured.HeightCm, "height_cm не пришёл — wrapper должен быть nil")
		require.NotNil(t, captured.WeightKg)
		require.NotNil(t, captured.WeightKg.Value)
		assert.Equal(t, 70.0, *captured.WeightKg.Value)
	})
}

// ===== DeleteUserHandler =====
func TestDeleteUserHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockProfilesClient{
			deleteMyProfileFunc: func(ctx context.Context) error {
				return nil
			},
		}
		handler := DeleteUserHandler(mock)

		req := httptest.NewRequest(http.MethodDelete, "/users/me", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		assert.Empty(t, rr.Body.String())
	})

	t.Run("gRPC error — not found", func(t *testing.T) {
		mock := &mockProfilesClient{
			deleteMyProfileFunc: func(ctx context.Context) error {
				return status.Error(codes.NotFound, "profile not found")
			},
		}
		handler := DeleteUserHandler(mock)

		req := httptest.NewRequest(http.MethodDelete, "/users/me", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "profile not found")
	})

	t.Run("gRPC error — unauthenticated", func(t *testing.T) {
		mock := &mockProfilesClient{
			deleteMyProfileFunc: func(ctx context.Context) error {
				return status.Error(codes.Unauthenticated, "unauthenticated")
			},
		}
		handler := DeleteUserHandler(mock)

		req := httptest.NewRequest(http.MethodDelete, "/users/me", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

// чтобы компилятор не ругался на неиспользуемый импорт nullable
var _ = nullable.Nullable[string]{}
