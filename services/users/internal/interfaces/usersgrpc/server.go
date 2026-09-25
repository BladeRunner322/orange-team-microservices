// Package usersgrpc — gRPC-сервер Users-сервиса.
//
// Реализует users.UsersServiceServer (сгенерирован из api/users/users.proto):
//   - GetMyProfile    — user_id из context (UserIDServerInterceptor)
//   - PatchMyProfile  — user_id из context
//   - DeleteMyProfile — user_id из context
//   - GetProfile      — user_id из req
//
// Маппит доменные ошибки в gRPC-статусы:
//   - domain.ErrProfileNotFound  → codes.NotFound
//   - domain.ErrInvalidWeight    → codes.InvalidArgument
package usersgrpc

// TODO: type Server struct + NewServer(usecases...) + методы RPC + mapper
