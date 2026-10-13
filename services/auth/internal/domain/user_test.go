package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	email, _ := NewEmail("user@example.com")
	passHash, _ := NewPasswordHash("$2a$10$dummyhash")
	fullName, _ := NewFullName("Test User")

	user := NewUser(email, passHash, fullName)

	t.Run("генерирует id", func(t *testing.T) {
		assert.NotEqual(t, uuid.Nil, user.ID())
	})

	t.Run("поля соответствуют переданным", func(t *testing.T) {
		assert.Equal(t, email, user.Email())
		assert.Equal(t, passHash, user.PasswordHash())
		assert.Equal(t, fullName, user.FullName())
	})

	t.Run("роль по умолчанию user", func(t *testing.T) {
		assert.Equal(t, RoleUser, user.Role())
	})

	t.Run("createdAt заполнен, updatedAt nil", func(t *testing.T) {
		assert.False(t, user.CreatedAt().IsZero())
		assert.Nil(t, user.UpdatedAt())
	})

	t.Run("id уникален между вызовами", func(t *testing.T) {
		other := NewUser(email, passHash, fullName)
		assert.NotEqual(t, user.ID(), other.ID())
	})
}

func TestRestoreUser(t *testing.T) {
	id := uuid.New()
	email, _ := NewEmail("user@example.com")
	passHash, _ := NewPasswordHash("$2a$10$dummyhash")
	fullName, _ := NewFullName("Test User")
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)

	t.Run("восстанавливает все поля", func(t *testing.T) {
		user := RestoreUser(id, email, passHash, fullName, RoleAdmin, createdAt, &updatedAt)

		assert.Equal(t, id, user.ID())
		assert.Equal(t, email, user.Email())
		assert.Equal(t, passHash, user.PasswordHash())
		assert.Equal(t, fullName, user.FullName())
		assert.Equal(t, RoleAdmin, user.Role())
		assert.Equal(t, createdAt, user.CreatedAt())
		require.NotNil(t, user.UpdatedAt())
		assert.Equal(t, updatedAt, *user.UpdatedAt())
	})

	t.Run("updatedAt nil", func(t *testing.T) {
		user := RestoreUser(id, email, passHash, fullName, RoleUser, createdAt, nil)

		assert.Nil(t, user.UpdatedAt())
	})
}