package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSecret   = "test-secret-32-bytes-long-xxxxxxxx"
	testIssuer   = "auth-service"
	testAudience = "orange-team"
)

func newManager() *Manager {
	return NewManager(testSecret, testIssuer, testAudience, 15*time.Minute)
}

func TestManager_GenerateAndValidate(t *testing.T) {
	t.Run("успешная генерация и валидация", func(t *testing.T) {
		m := newManager()
		ctx := context.Background()

		token, err := m.Generate(ctx, "user-123", "admin")
		require.NoError(t, err)
		require.NotEmpty(t, token)

		info, err := m.Validate(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, "user-123", info.UserID)
		assert.Equal(t, "admin", info.Role)
	})
}

func TestManager_Validate_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("мусор вместо токена", func(t *testing.T) {
		m := newManager()

		_, err := m.Validate(ctx, "not-a-token")

		assert.Error(t, err)
	})

	t.Run("пустая строка", func(t *testing.T) {
		m := newManager()

		_, err := m.Validate(ctx, "")

		assert.Error(t, err)
	})

	t.Run("токен подписан другим секретом", func(t *testing.T) {
		other := NewManager("another-secret-32-bytes-long-yyy", testIssuer, testAudience, 15*time.Minute)
		token, err := other.Generate(ctx, "user-123", "admin")
		require.NoError(t, err)

		m := newManager()
		_, err = m.Validate(ctx, token)

		assert.Error(t, err)
	})

	t.Run("токен с чужим issuer", func(t *testing.T) {
		other := NewManager(testSecret, "other-service", testAudience, 15*time.Minute)
		token, err := other.Generate(ctx, "user-123", "admin")
		require.NoError(t, err)

		m := newManager()
		_, err = m.Validate(ctx, token)

		assert.Error(t, err)
	})

	t.Run("токен с чужим audience", func(t *testing.T) {
		other := NewManager(testSecret, testIssuer, "other-audience", 15*time.Minute)
		token, err := other.Generate(ctx, "user-123", "admin")
		require.NoError(t, err)

		m := newManager()
		_, err = m.Validate(ctx, token)

		assert.Error(t, err)
	})

	t.Run("истёкший токен", func(t *testing.T) {
		m := NewManager(testSecret, testIssuer, testAudience, -1*time.Hour)
		token, err := m.Generate(ctx, "user-123", "admin")
		require.NoError(t, err)

		_, err = m.Validate(ctx, token)

		assert.Error(t, err)
	})
}
