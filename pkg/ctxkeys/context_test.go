package ctxkeys

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

func TestRequestID(t *testing.T) {
	t.Run("установка и чтение", func(t *testing.T) {
		ctx := WithRequestID(context.Background(), "req-123")

		got, ok := RequestIDFromContext(ctx)

		assert.True(t, ok)
		assert.Equal(t, "req-123", got)
	})

	t.Run("пустой контекст", func(t *testing.T) {
		_, ok := RequestIDFromContext(context.Background())
		assert.False(t, ok)
	})

	t.Run("пустая строка — тоже валидное значение", func(t *testing.T) {
		ctx := WithRequestID(context.Background(), "")

		got, ok := RequestIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "", got)
	})
}

func TestKeysAreIndependent(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserID(ctx, "user-1")
	ctx = WithRole(ctx, "admin")
	ctx = WithRequestID(ctx, "req-abc")

	userID, ok := UserIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "user-1", userID)

	role, ok := RoleFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "admin", role)

	requestID, ok := RequestIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "req-abc", requestID)
}

func TestMetadataConstants(t *testing.T) {
	// Константы публичные — используются как HTTP-заголовки и gRPC
	// metadata. Изменение значения — ломающее для клиентов.
	t.Run("MetadataUserID", func(t *testing.T) {
		assert.Equal(t, "x-user-id", MetadataUserID)
	})

	t.Run("MetadataRole", func(t *testing.T) {
		assert.Equal(t, "x-user-role", MetadataRole)
	})

	t.Run("MetadataRequestID", func(t *testing.T) {
		assert.Equal(t, "x-request-id", MetadataRequestID)
	})
}
