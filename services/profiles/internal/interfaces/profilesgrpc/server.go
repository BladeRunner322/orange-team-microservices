package profilesgrpc

import (
	"context"
	"errors"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/profiles"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/application/usecases"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server реализует profiles.ProfilesServiceServer.
// Зависимости — usecases для четырёх RPC.
type Server struct {
	profiles.UnimplementedProfilesServiceServer
	getMyProfileUC    *usecases.GetMyProfile
	getProfileUC      *usecases.GetProfile
	patchMyProfileUC  *usecases.PatchMyProfile
	deleteMyProfileUC *usecases.DeleteMyProfile
}

func NewServer(
	getMyProfileUC *usecases.GetMyProfile,
	getProfileUC *usecases.GetProfile,
	patchMyProfileUC *usecases.PatchMyProfile,
	deleteMyProfileUC *usecases.DeleteMyProfile,
) *Server {
	return &Server{
		getMyProfileUC:    getMyProfileUC,
		getProfileUC:      getProfileUC,
		patchMyProfileUC:  patchMyProfileUC,
		deleteMyProfileUC: deleteMyProfileUC,
	}
}

// GetMyProfile возвращает профиль текущего пользователя (user_id из context).
// Если профиля нет — создаёт пустой (lazy-create).
func (s *Server) GetMyProfile(
	ctx context.Context,
	req *profiles.GetMyProfileRequest,
) (*profiles.UserProfile, error) {
	profile, err := s.getMyProfileUC.Execute(ctx)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUnauthenticated):
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")

		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return ToProtoProfile(profile), nil
}

// GetProfile возвращает профиль указанного пользователя (user_id из req).
// Внутренний метод: вызывается другими сервисами через gRPC.
func (s *Server) GetProfile(
	ctx context.Context,
	req *profiles.GetProfileRequest,
) (*profiles.UserProfile, error) {
	userID, err := ToDomainUserID(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	profile, err := s.getProfileUC.Execute(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return ToProtoProfile(profile), nil
}

// PatchMyProfile применяет патч к профилю текущего пользователя.
// Если профиля нет — возвращает NotFound.
func (s *Server) PatchMyProfile(
	ctx context.Context,
	req *profiles.PatchMyProfileRequest,
) (*profiles.UserProfile, error) {
	patch, err := ToDomainProfilePatch(req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidSex):
			return nil, status.Error(codes.InvalidArgument, "invalid sex")

		case errors.Is(err, domain.ErrInvalidWeight):
			return nil, status.Error(codes.InvalidArgument, "invalid weight")

		case errors.Is(err, domain.ErrInvalidBirthDate):
			return nil, status.Error(codes.InvalidArgument, "invalid birth_date")

		case errors.Is(err, domain.ErrInvalidHeight):
			return nil, status.Error(codes.InvalidArgument, "invalid height")

		default:
			return nil, status.Error(codes.InvalidArgument, "invalid patch data")
		}
	}

	profile, err := s.patchMyProfileUC.Execute(ctx, patch)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUnauthenticated):
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")

		case errors.Is(err, domain.ErrProfileNotFound):
			return nil, status.Error(codes.NotFound, "profile not found")

		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return ToProtoProfile(profile), nil
}

// DeleteMyProfile удаляет профиль текущего пользователя.
// Если профиля нет — возвращает NotFound.
func (s *Server) DeleteMyProfile(
	ctx context.Context,
	req *profiles.DeleteMyProfileRequest,
) (*profiles.DeleteMyProfileResponse, error) {
	err := s.deleteMyProfileUC.Execute(ctx)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUnauthenticated):
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")

		case errors.Is(err, domain.ErrProfileNotFound):
			return nil, status.Error(codes.NotFound, "profile not found")

		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return &profiles.DeleteMyProfileResponse{}, nil
}
