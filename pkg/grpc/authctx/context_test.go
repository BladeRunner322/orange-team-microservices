package authctx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserID(t *testing.T) {
	t.Run("установка и чтение", func(t *testing.T) {
		ctx := WithUserID(context.Background(), "user-123")

		got, ok := UserIDFromContext(ctx)

		assert.True(t, ok)
		assert.Equal(t, "user-123", got)
	})

	t.Run("пустой контекст", func(t *testing.T) {
		_, ok := UserIDFromContext(context.Background())
		assert.False(t, ok)
	})

	t.Run("пустая строка — тоже валидное значение", func(t *testing.T) {
		ctx := WithUserID(context.Background(), "")

		got, ok := UserIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "", got)
	})
}

func TestRole(t *testing.T) {
	t.Run("установка и чтение", func(t *testing.T) {
		ctx := WithRole(context.Background(), "admin")

		got, ok := RoleFromContext(ctx)

		assert.True(t, ok)
		assert.Equal(t, "admin", got)
	})

	t.Run("пустой контекст", func(t *testing.T) {
		_, ok := RoleFromContext(context.Background())
		assert.False(t, ok)
	})
}

func TestUserIDAndRoleIndependently(t *testing.T) {
	ctx := WithUserID(context.Background(), "user-1")
	ctx = WithRole(ctx, "admin")

	userID, ok := UserIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "user-1", userID)

	role, ok := RoleFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "admin", role)
}
