package ports

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/profiles"
)

// ProfilesClientInterface — контракт gRPC-клиента к Profiles Service.
type ProfilesClientInterface interface {
	GetMyProfile(ctx context.Context) (*profiles.UserProfile, error)
	PatchMyProfile(ctx context.Context, req *profiles.PatchMyProfileRequest) (*profiles.UserProfile, error)
	DeleteMyProfile(ctx context.Context) error
	Close()
}
