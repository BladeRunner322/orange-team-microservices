package ports

import "context"

// UserInfo — данные пользователя, извлекаемые из токена.
type UserInfo struct {
	UserID string
	Role   string
}

type TokenManager interface {
	Generate(ctx context.Context, userID string, role string) (string, error)
	Validate(ctx context.Context, token string) (UserInfo, error)
}
