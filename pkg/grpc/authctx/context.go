// Package authctx хранит информацию об аутентифицированном пользователе
// в context.Context и определяет ключи gRPC metadata для её передачи
// между сервисами.
package authctx

import "context"

// Metadata-ключи, используемые для передачи данных пользователя
// между сервисами через gRPC.
const (
	MetadataUserID = "x-user-id"
	MetadataRole   = "x-user-role"
)

type (
	userIDKey struct{}
	roleKey   struct{}
)

// WithUserID кладёт user_id в контекст.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// UserIDFromContext достаёт user_id из контекста.
// Возвращает ok=false, если user_id не установлен.
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	return userID, ok
}

// WithRole кладёт роль пользователя в контекст.
func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey{}, role)
}

// RoleFromContext достаёт роль пользователя из контекста.
// Возвращает ok=false, если роль не установлена.
func RoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleKey{}).(string)
	return role, ok
}
