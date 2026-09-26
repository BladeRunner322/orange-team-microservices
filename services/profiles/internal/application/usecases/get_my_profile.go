// Package usecases — бизнес-логика Profiles-сервиса.
//
// GetMyProfile:
//   - user_id из authctx (metadata x-user-id)
//   - lazy-create: если профиля нет — создать пустой (INSERT ... ON CONFLICT DO NOTHING)
//   - вернуть profile с флагом profile_completed
//
// См. ADR-001.
package usecases

// TODO: type GetMyProfile struct + NewGetMyProfile() + Execute()
