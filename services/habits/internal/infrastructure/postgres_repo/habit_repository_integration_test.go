//go:build integration
// +build integration

package postgres_repo

import (
	"context"
	"testing"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHabitRepository_Integration(t *testing.T) {
	repo := NewHabitRepository(testPool)
	runHabitRepositoryTests(t, repo)
}

// mustHabit создаёт доменную привычку или падает.
func mustHabit(t *testing.T, userID uuid.UUID, name, description string) domain.Habit {
	t.Helper()

	n, err := domain.NewName(name)
	require.NoError(t, err)

	d, err := domain.NewDescription(description)
	require.NoError(t, err)

	return domain.NewHabit(userID, n, d)
}

func runHabitRepositoryTests(t *testing.T, repo *HabitRepository) {
	ctx := context.Background()

	uniqueName := func(prefix string) string {
		return prefix + " " + uuid.New().String()
	}

	// ============================================================
	//  Create
	// ============================================================

	t.Run("Create сохраняет привычку и возвращает её из RETURNING", func(t *testing.T) {
		userID := uuid.New()
		h := mustHabit(t, userID, uniqueName("Бег"), "Пробежка утром")

		created, err := repo.Create(ctx, h)
		require.NoError(t, err)

		assert.Equal(t, h.ID(), created.ID())
		assert.Equal(t, userID, created.UserID())
		assert.Equal(t, h.Name().String(), created.Name().String())
		assert.Equal(t, h.Description().String(), created.Description().String())
		assert.False(t, created.IsDeleted())
		assert.False(t, created.CreatedAt().IsZero())
		assert.Nil(t, created.UpdatedAt())
	})

	t.Run("Create с дублирующимся name у одного user — ErrHabitNameExists", func(t *testing.T) {
		userID := uuid.New()
		name := uniqueName("Медитация")

		_, err := repo.Create(ctx, mustHabit(t, userID, name, "Описание 1"))
		require.NoError(t, err)

		_, err = repo.Create(ctx, mustHabit(t, userID, name, "Описание 2"))
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrHabitNameExists)
	})

	t.Run("Create с тем же name у разных user — ок", func(t *testing.T) {
		name := uniqueName("Чтение")

		_, err := repo.Create(ctx, mustHabit(t, uuid.New(), name, "Описание"))
		require.NoError(t, err)

		_, err = repo.Create(ctx, mustHabit(t, uuid.New(), name, "Описание"))
		require.NoError(t, err)
	})

	// ============================================================
	//  GetByID
	// ============================================================

	t.Run("GetByID возвращает активную привычку", func(t *testing.T) {
		created, err := repo.Create(ctx, mustHabit(t, uuid.New(), uniqueName("Зарядка"), "10 минут"))
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, created.ID())
		require.NoError(t, err)

		assert.Equal(t, created.ID(), found.ID())
		assert.Equal(t, created.UserID(), found.UserID())
		assert.False(t, found.IsDeleted())
	})

	t.Run("GetByID не найдено — ErrHabitNotFound", func(t *testing.T) {
		_, err := repo.GetByID(ctx, uuid.New())
		assert.ErrorIs(t, err, domain.ErrHabitNotFound)
	})

	t.Run("GetByID возвращает удалённое с IsDeleted=true", func(t *testing.T) {
		created, err := repo.Create(ctx, mustHabit(t, uuid.New(), uniqueName("Удалённая"), "Описание"))
		require.NoError(t, err)

		require.NoError(t, repo.MarkDeleted(ctx, created.ID()))

		found, err := repo.GetByID(ctx, created.ID())
		require.NoError(t, err)
		assert.True(t, found.IsDeleted())
	})

	// ============================================================
	//  GetByIDForUpdate
	// ============================================================

	t.Run("GetByIDForUpdate — параллельный FOR UPDATE ждёт ROLLBACK", func(t *testing.T) {
		created, err := repo.Create(ctx, mustHabit(t, uuid.New(), uniqueName("Блокировка"), "Описание"))
		require.NoError(t, err)

		tx, err := repo.pool.(*postgres.PgxPool).BeginTx(ctx)
		require.NoError(t, err)

		txRepo := &HabitRepository{tx: tx, opTimeout: 5 * time.Second}
		_, err = txRepo.GetByIDForUpdate(ctx, created.ID())
		require.NoError(t, err)

		done := make(chan error, 1)
		go func() {
			_, err := repo.GetByIDForUpdate(ctx, created.ID())
			done <- err
		}()

		select {
		case <-done:
			t.Fatal("параллельный FOR UPDATE не должен завершиться, пока первая транзакция открыта")
		case <-time.After(200 * time.Millisecond):
		}

		require.NoError(t, tx.Rollback(ctx))

		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("параллельный FOR UPDATE должен был завершиться после ROLLBACK")
		}
	})

	// ============================================================
	//  ListByUser
	// ============================================================

	t.Run("ListByUser возвращает только активные привычки конкретного user", func(t *testing.T) {
		userID := uuid.New()

		h1, err := repo.Create(ctx, mustHabit(t, userID, uniqueName("Первая"), "Описание 1"))
		require.NoError(t, err)

		h2, err := repo.Create(ctx, mustHabit(t, userID, uniqueName("Вторая"), "Описание 2"))
		require.NoError(t, err)

		// чужой пользователь — не должен попасть
		_, err = repo.Create(ctx, mustHabit(t, uuid.New(), uniqueName("Чужая"), "Описание"))
		require.NoError(t, err)

		// удалённая — не должна попасть
		deleted, err := repo.Create(ctx, mustHabit(t, userID, uniqueName("Удалённая"), "Описание"))
		require.NoError(t, err)
		require.NoError(t, repo.MarkDeleted(ctx, deleted.ID()))

		list, err := repo.ListByUser(ctx, userID)
		require.NoError(t, err)
		require.Len(t, list, 2)

		ids := []uuid.UUID{list[0].ID(), list[1].ID()}
		assert.Contains(t, ids, h1.ID())
		assert.Contains(t, ids, h2.ID())

		for _, h := range list {
			assert.Equal(t, userID, h.UserID())
			assert.False(t, h.IsDeleted())
		}

		// порядок по created_at ASC
		for i := 1; i < len(list); i++ {
			assert.False(t, list[i-1].CreatedAt().After(list[i].CreatedAt()))
		}
	})

	t.Run("ListByUser пустой slice, не nil", func(t *testing.T) {
		list, err := repo.ListByUser(ctx, uuid.New())
		require.NoError(t, err)
		assert.NotNil(t, list)
		assert.Empty(t, list)
	})

	// ============================================================
	//  Update
	// ============================================================

	t.Run("Update меняет name/description и возвращает свежий updated_at", func(t *testing.T) {
		created, err := repo.Create(ctx, mustHabit(t, uuid.New(), uniqueName("Старое"), "Старое описание"))
		require.NoError(t, err)

		newName, err := domain.NewName(uniqueName("Новое"))
		require.NoError(t, err)

		newDesc, err := domain.NewDescription("Новое описание")
		require.NoError(t, err)

		patched := domain.RestoreHabit(
			created.ID(),
			created.UserID(),
			newName,
			newDesc,
			false,
			created.CreatedAt(),
			created.UpdatedAt(),
		)

		updated, err := repo.Update(ctx, patched)
		require.NoError(t, err)

		assert.Equal(t, newName.String(), updated.Name().String())
		assert.Equal(t, newDesc.String(), updated.Description().String())
		require.NotNil(t, updated.UpdatedAt())

		found, err := repo.GetByID(ctx, created.ID())
		require.NoError(t, err)
		assert.Equal(t, newName.String(), found.Name().String())
		require.NotNil(t, found.UpdatedAt())
		assert.Equal(t, *updated.UpdatedAt(), *found.UpdatedAt())
	})

	t.Run("Update несуществующего — ErrHabitNotFound", func(t *testing.T) {
		h := mustHabit(t, uuid.New(), uniqueName("Призрак"), "Описание")
		_, err := repo.Update(ctx, h)
		assert.ErrorIs(t, err, domain.ErrHabitNotFound)
	})

	t.Run("Update удалённого — ErrHabitNotFound", func(t *testing.T) {
		created, err := repo.Create(ctx, mustHabit(t, uuid.New(), uniqueName("Удалённая"), "Описание"))
		require.NoError(t, err)
		require.NoError(t, repo.MarkDeleted(ctx, created.ID()))

		newName, _ := domain.NewName(uniqueName("Патч"))
		patched := domain.RestoreHabit(
			created.ID(),
			created.UserID(),
			newName,
			created.Description(),
			true,
			created.CreatedAt(),
			created.UpdatedAt(),
		)

		_, err = repo.Update(ctx, patched)
		assert.ErrorIs(t, err, domain.ErrHabitNotFound)
	})

	t.Run("Update с конфликтом имени у того же user — ErrHabitNameExists", func(t *testing.T) {
		userID := uuid.New()

		first, err := repo.Create(ctx, mustHabit(t, userID, uniqueName("Первая"), "Описание 1"))
		require.NoError(t, err)

		second, err := repo.Create(ctx, mustHabit(t, userID, uniqueName("Вторая"), "Описание 2"))
		require.NoError(t, err)

		patched := domain.RestoreHabit(
			second.ID(),
			second.UserID(),
			first.Name(),
			second.Description(),
			false,
			second.CreatedAt(),
			second.UpdatedAt(),
		)

		_, err = repo.Update(ctx, patched)
		assert.ErrorIs(t, err, domain.ErrHabitNameExists)
	})

	t.Run("Update со своим же именем — без конфликта", func(t *testing.T) {
		created, err := repo.Create(ctx, mustHabit(t, uuid.New(), uniqueName("Имя"), "Старое описание"))
		require.NoError(t, err)

		newDesc, _ := domain.NewDescription("Новое описание")
		patched := domain.RestoreHabit(
			created.ID(),
			created.UserID(),
			created.Name(),
			newDesc,
			false,
			created.CreatedAt(),
			created.UpdatedAt(),
		)

		updated, err := repo.Update(ctx, patched)
		require.NoError(t, err)
		assert.Equal(t, created.Name().String(), updated.Name().String())
		assert.Equal(t, newDesc.String(), updated.Description().String())
	})

	// ============================================================
	//  MarkDeleted
	// ============================================================

	t.Run("MarkDeleted помечает привычку удалённой", func(t *testing.T) {
		created, err := repo.Create(ctx, mustHabit(t, uuid.New(), uniqueName("Удаляем"), "Описание"))
		require.NoError(t, err)

		require.NoError(t, repo.MarkDeleted(ctx, created.ID()))

		found, err := repo.GetByID(ctx, created.ID())
		require.NoError(t, err)
		assert.True(t, found.IsDeleted())
		require.NotNil(t, found.UpdatedAt())
	})

	t.Run("MarkDeleted повторно — ErrHabitNotFound", func(t *testing.T) {
		created, err := repo.Create(ctx, mustHabit(t, uuid.New(), uniqueName("Повтор"), "Описание"))
		require.NoError(t, err)
		require.NoError(t, repo.MarkDeleted(ctx, created.ID()))

		err = repo.MarkDeleted(ctx, created.ID())
		assert.ErrorIs(t, err, domain.ErrHabitNotFound)
	})

	t.Run("MarkDeleted несуществующего — ErrHabitNotFound", func(t *testing.T) {
		err := repo.MarkDeleted(ctx, uuid.New())
		assert.ErrorIs(t, err, domain.ErrHabitNotFound)
	})

	// ============================================================
	//  Soft delete + partial unique index
	// ============================================================

	t.Run("после MarkDeleted можно пересоздать с тем же именем у того же user", func(t *testing.T) {
		userID := uuid.New()
		name := uniqueName("Пересоздаваемое")

		first, err := repo.Create(ctx, mustHabit(t, userID, name, "Описание 1"))
		require.NoError(t, err)

		_, err = repo.Create(ctx, mustHabit(t, userID, name, "Описание 2"))
		require.ErrorIs(t, err, domain.ErrHabitNameExists)

		require.NoError(t, repo.MarkDeleted(ctx, first.ID()))

		second, err := repo.Create(ctx, mustHabit(t, userID, name, "Описание 3"))
		require.NoError(t, err)
		assert.Equal(t, name, second.Name().String())
		assert.NotEqual(t, first.ID(), second.ID())
	})
}
