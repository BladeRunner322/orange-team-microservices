// Package ctxkeys хранит ключи контекста, которые пропагируются
// между сервисами через HTTP-заголовки и gRPC metadata.
//
// Пакет нейтрален к транспорту: значения устанавливаются в HTTP-слое
// (Gateway middleware), передаются через gRPC metadata (interceptors)
// и читаются в любом слое через функции FromContext.
//
// Имена констант совпадают с именами HTTP-заголовков и gRPC metadata —
// так один источник истины для написания.
package ctxkeys

import "context"

const (
	MetadataUserID    = "x-user-id"
	MetadataRole      = "x-user-role"
	MetadataRequestID = "x-request-id"
)

type (
	userIDKey    struct{}
	roleKey      struct{}
	requestIDKey struct{}
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

// WithRequestID кладёт request_id в контекст.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// RequestIDFromContext достаёт request_id из контекста.
// Возвращает ok=false, если request_id не установлен.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDKey{}).(string)
	return requestID, ok
}
