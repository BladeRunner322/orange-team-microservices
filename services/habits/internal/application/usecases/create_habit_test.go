package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BladeRunner322/orange-team-microservices/pkg/ctxkeys"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
)

func TestCreateHabit_Execute(t *testing.T) {
	newCtx := func(userID uuid.UUID) context.Context {
		return ctxkeys.WithUserID(context.Background(), userID.String())
	}

	t.Run("успешно создаёт привычку", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockHabitRepository()
		uc := NewCreateHabit(repo, logger.NewTestLogger())

		// Act
		created, err := uc.Execute(ctx, "Пить воду", "2 литра в день")

		// Assert
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, created.ID())
		assert.Equal(t, userID, created.UserID())
		assert.Equal(t, "Пить воду", created.Name().String())
		assert.Equal(t, "2 литра в день", created.Description().String())
		assert.False(t, created.IsDeleted())

		saved, err := repo.GetByID(context.Background(), created.ID())
		require.NoError(t, err)
		assert.Equal(t, created.ID(), saved.ID())
	})

	t.Run("успешно создаёт с пустым description", func(t *testing.T) {
		userID := uuid.New()
		ctx := newCtx(userID)
		repo := newMockHabitRepository()
		uc := NewCreateHabit(repo, logger.NewTestLogger())

		created, err := uc.Execute(ctx, "Читать", "")

		require.NoError(t, err)
		assert.Equal(t, "Читать", created.Name().String())
		assert.Equal(t, "", created.Description().String())
	})

	t.Run("нет user_id в context — ErrUnauthenticated", func(t *testing.T) {
		repo := newMockHabitRepository()
		uc := NewCreateHabit(repo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), "Пить воду", "")

		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})

	t.Run("невалидный user_id в context — ErrUnauthenticated", func(t *testing.T) {
		ctx := ctxkeys.WithUserID(context.Background(), "not-a-uuid")
		repo := newMockHabitRepository()
		uc := NewCreateHabit(repo, logger.NewTestLogger())

		_, err := uc.Execute(ctx, "Пить воду", "")

		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})

	t.Run("невалидное name — ErrInvalidName", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockHabitRepository()
		uc := NewCreateHabit(repo, logger.NewTestLogger())

		_, err := uc.Execute(ctx, "", "Описание")

		assert.ErrorIs(t, err, domain.ErrInvalidName)
	})

	t.Run("слишком длинное name — ErrInvalidName", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockHabitRepository()
		uc := NewCreateHabit(repo, logger.NewTestLogger())

		longName := make([]byte, 101)
		for i := range longName {
			longName[i] = 'a'
		}
		_, err := uc.Execute(ctx, string(longName), "")

		assert.ErrorIs(t, err, domain.ErrInvalidName)
	})

	t.Run("слишком длинное description — ErrInvalidDescription", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockHabitRepository()
		uc := NewCreateHabit(repo, logger.NewTestLogger())

		longDesc := make([]byte, 501)
		for i := range longDesc {
			longDesc[i] = 'a'
		}
		_, err := uc.Execute(ctx, "Пить воду", string(longDesc))

		assert.ErrorIs(t, err, domain.ErrInvalidDescription)
	})

	t.Run("имя занято у того же user — ErrHabitNameExists", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockHabitRepository()
		uc := NewCreateHabit(repo, logger.NewTestLogger())

		_, err := uc.Execute(ctx, "Пить воду", "Описание 1")
		require.NoError(t, err)

		_, err = uc.Execute(ctx, "Пить воду", "Описание 2")

		assert.ErrorIs(t, err, domain.ErrHabitNameExists)
	})

	t.Run("имя занято у другого user — можно", func(t *testing.T) {
		repo := newMockHabitRepository()
		uc := NewCreateHabit(repo, logger.NewTestLogger())

		_, err := uc.Execute(newCtx(uuid.New()), "Пить воду", "Описание 1")
		require.NoError(t, err)

		_, err = uc.Execute(newCtx(uuid.New()), "Пить воду", "Описание 2")

		require.NoError(t, err)
	})

	t.Run("ошибка репозитория при Create — пробрасывается", func(t *testing.T) {
		ctx := newCtx(uuid.New())
		repo := newMockHabitRepository()
		repo.err = errors.New("db is down")
		uc := NewCreateHabit(repo, logger.NewTestLogger())

		_, err := uc.Execute(ctx, "Пить воду", "")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}
