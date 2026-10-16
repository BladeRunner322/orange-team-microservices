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

// patchWithName — хелпер: собирает патч только с name.
func patchWithName(raw string) domain.HabitPatch {
	name, _ := domain.NewName(raw)
	return domain.NewHabitPatch(&name, nil)
}

// patchWithDescription — хелпер: собирает патч только с description.
func patchWithDescription(raw string) domain.HabitPatch {
	description, _ := domain.NewDescription(raw)
	return domain.NewHabitPatch(nil, &description)
}

func TestPatchHabit_Execute(t *testing.T) {
	newCtx := func(userID uuid.UUID) context.Context {
		return ctxkeys.WithUserID(context.Background(), userID.String())
	}

	// createHabit создаёт привычку в моковом репозитории.
	createHabit := func(t *testing.T, repo *mockHabitRepository, userID uuid.UUID, name string) domain.Habit {
		t.Helper()
		n, err := domain.NewName(name)
		require.NoError(t, err)
		d, err := domain.NewDescription("Описание")
		require.NoError(t, err)
		created, err := repo.Create(context.Background(), domain.NewHabit(userID, n, d))
		require.NoError(t, err)
		return created
	}

	t.Run("успешно патчит name", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		repo := newMockHabitRepository()
		habit := createHabit(t, repo, userID, "Старое имя")
		uc := NewPatchHabit(repo, logger.NewTestLogger())

		// Act
		updated, err := uc.Execute(newCtx(userID), habit.ID(), patchWithName("Новое имя"))

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "Новое имя", updated.Name().String())
		assert.Equal(t, "Описание", updated.Description().String())
	})

	t.Run("успешно патчит несколько полей", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		habit := createHabit(t, repo, userID, "Старое имя")

		newName, _ := domain.NewName("Новое имя")
		newDesc, _ := domain.NewDescription("Новое описание")
		patch := domain.NewHabitPatch(&newName, &newDesc)

		uc := NewPatchHabit(repo, logger.NewTestLogger())

		updated, err := uc.Execute(newCtx(userID), habit.ID(), patch)

		require.NoError(t, err)
		assert.Equal(t, "Новое имя", updated.Name().String())
		assert.Equal(t, "Новое описание", updated.Description().String())
	})

	t.Run("пустой патч — ничего не меняется", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		uc := NewPatchHabit(repo, logger.NewTestLogger())

		updated, err := uc.Execute(newCtx(userID), habit.ID(), domain.NewHabitPatch(nil, nil))

		require.NoError(t, err)
		assert.Equal(t, "Пить воду", updated.Name().String())
		assert.Equal(t, "Описание", updated.Description().String())
	})

	t.Run("привычка не найдена — ErrHabitNotFound", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		uc := NewPatchHabit(repo, logger.NewTestLogger())

		_, err := uc.Execute(newCtx(userID), uuid.New(), patchWithName("Новое имя"))

		assert.ErrorIs(t, err, domain.ErrHabitNotFound)
	})

	t.Run("чужая привычка — ErrHabitNotFound", func(t *testing.T) {
		ownerID := uuid.New()
		otherID := uuid.New()
		repo := newMockHabitRepository()
		habit := createHabit(t, repo, ownerID, "Пить воду")
		uc := NewPatchHabit(repo, logger.NewTestLogger())

		_, err := uc.Execute(newCtx(otherID), habit.ID(), patchWithName("Новое имя"))

		assert.ErrorIs(t, err, domain.ErrHabitNotFound)
	})

	t.Run("патч удалённой привычки — ErrHabitNotFound", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		habit := createHabit(t, repo, userID, "Пить воду")
		require.NoError(t, repo.MarkDeleted(context.Background(), habit.ID()))

		uc := NewPatchHabit(repo, logger.NewTestLogger())

		_, err := uc.Execute(newCtx(userID), habit.ID(), patchWithName("Новое имя"))

		assert.ErrorIs(t, err, domain.ErrHabitNotFound)
	})

	t.Run("конфликт имён — ErrHabitNameExists", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		first := createHabit(t, repo, userID, "Пить воду")
		_ = createHabit(t, repo, userID, "Читать")

		uc := NewPatchHabit(repo, logger.NewTestLogger())

		_, err := uc.Execute(newCtx(userID), first.ID(), patchWithName("Читать"))

		assert.ErrorIs(t, err, domain.ErrHabitNameExists)
	})

	t.Run("патч с тем же именем — не конфликт", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		uc := NewPatchHabit(repo, logger.NewTestLogger())

		updated, err := uc.Execute(newCtx(userID), habit.ID(), patchWithName("Пить воду"))

		require.NoError(t, err)
		assert.Equal(t, "Пить воду", updated.Name().String())
	})

	t.Run("сброс description в пустое", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		habit := createHabit(t, repo, userID, "Пить воду")

		uc := NewPatchHabit(repo, logger.NewTestLogger())

		updated, err := uc.Execute(newCtx(userID), habit.ID(), patchWithDescription(""))

		require.NoError(t, err)
		assert.Equal(t, "", updated.Description().String())
	})

	t.Run("нет user_id в context — ErrUnauthenticated", func(t *testing.T) {
		repo := newMockHabitRepository()
		uc := NewPatchHabit(repo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), uuid.New(), patchWithName("X"))

		assert.ErrorIs(t, err, domain.ErrUnauthenticated)
	})

	t.Run("ошибка репозитория — пробрасывается", func(t *testing.T) {
		userID := uuid.New()
		repo := newMockHabitRepository()
		repo.err = errors.New("db is down")
		uc := NewPatchHabit(repo, logger.NewTestLogger())

		_, err := uc.Execute(newCtx(userID), uuid.New(), patchWithName("X"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}
