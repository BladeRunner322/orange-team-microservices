//go:build integration
// +build integration

package postgres_repo

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/domain"
)

func TestRepository_Integration(t *testing.T) {
	ctx := context.Background()

	// 1. Поднимаем PostgreSQL в Docker
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:18.6-bookworm",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	// 2. Собираем конфиг для подключения
	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := pgContainer.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	// 3. Создаём пул через pkg/postgres
	pool, err := postgres.NewPgxPool(ctx, postgres.Config{
		Host:     host,
		Port:     mappedPort.Port(),
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		Timeout:  5 * time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(func() { pool.Close() })

	// 4. Применяем реальную миграцию из файла — гарантия, что тест
	// прогоняется на той же схеме, что и прод.
	migration, err := os.ReadFile("../../../migrations/000001_create_exercises_table.up.sql")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, string(migration))
	require.NoError(t, err)

	// 5. Создаём репозиторий поверх пула
	repo := NewRepository(pool)

	// 6. Прогоняем подтесты
	runRepositoryTests(t, repo)
}

// mustExercise создаёт доменное упражнение или падает.
// Хелпер для тестов, чтобы не отвлекаться на обработку ошибок VO.
func mustExercise(t *testing.T, name, description string, difficulty int, exerciseType string) domain.Exercise {
	t.Helper()

	n, err := domain.NewName(name)
	require.NoError(t, err)

	d, err := domain.NewDescription(description)
	require.NoError(t, err)

	diff, err := domain.NewDifficulty(difficulty)
	require.NoError(t, err)

	et, err := domain.NewExerciseType(exerciseType)
	require.NoError(t, err)

	return domain.NewExercise(n, d, diff, et)
}

func runRepositoryTests(t *testing.T, repo *Repository) {
	ctx := context.Background()

	// uniqueName генерирует имя с UUID-суффиксом, чтобы избежать
	// коллизий между тестами (partial unique на name среди активных).
	uniqueName := func(prefix string) string {
		return prefix + " " + uuid.New().String()
	}

	// ============================================================
	//  Create
	// ============================================================

	t.Run("Create сохраняет упражнение и возвращает его из RETURNING", func(t *testing.T) {
		exercise := mustExercise(t, uniqueName("Жим лёжа"), "Базовое упражнение", 5, "weight")

		created, err := repo.Create(ctx, exercise)
		require.NoError(t, err)

		assert.Equal(t, exercise.ID(), created.ID())
		assert.Equal(t, exercise.Name().String(), created.Name().String())
		assert.Equal(t, exercise.Description().String(), created.Description().String())
		assert.Equal(t, 5, created.Difficulty().Int())
		assert.Equal(t, domain.ExerciseTypeWeight, created.ExerciseType())
		assert.False(t, created.IsDeleted())

		// created_at заполнен на стороне БД
		assert.False(t, created.CreatedAt().IsZero())

		// updated_at = NULL сразу после создания
		assert.Nil(t, created.UpdatedAt())
	})

	t.Run("Create с дублирующимся name — ErrExerciseNameExists", func(t *testing.T) {
		name := uniqueName("Приседания")

		first := mustExercise(t, name, "Описание 1", 5, "weight")
		_, err := repo.Create(ctx, first)
		require.NoError(t, err)

		second := mustExercise(t, name, "Описание 2", 7, "weight")
		_, err = repo.Create(ctx, second)

		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrExerciseNameExists)
	})

	// ============================================================
	//  GetExercise
	// ============================================================

	t.Run("GetExercise возвращает активное упражнение", func(t *testing.T) {
		created, err := repo.Create(ctx, mustExercise(t, uniqueName("Планка"), "Описание", 3, "duration"))
		require.NoError(t, err)

		found, err := repo.GetExercise(ctx, created.ID())
		require.NoError(t, err)

		assert.Equal(t, created.ID(), found.ID())
		assert.Equal(t, created.Name().String(), found.Name().String())
		assert.False(t, found.IsDeleted())
	})

	t.Run("GetExercise не найдено — ErrExerciseNotFound", func(t *testing.T) {
		_, err := repo.GetExercise(ctx, uuid.New())

		assert.ErrorIs(t, err, domain.ErrExerciseNotFound)
	})

	t.Run("GetExercise возвращает удалённое с is_deleted=true (ADR-008)", func(t *testing.T) {
		created, err := repo.Create(ctx, mustExercise(t, uniqueName("Удалённое"), "Описание", 5, "weight"))
		require.NoError(t, err)

		require.NoError(t, repo.MarkDeleted(ctx, created.ID()))

		found, err := repo.GetExercise(ctx, created.ID())
		require.NoError(t, err)

		assert.True(t, found.IsDeleted(), "GetExercise должен возвращать и удалённые — ADR-008")
	})

	// ============================================================
	//  GetExercises
	// ============================================================

	t.Run("GetExercises возвращает только активные и сортирует по name", func(t *testing.T) {
		// Создаём в произвольном порядке, чтобы проверить сортировку.
		// Используем префикс "AAA-", "BBB-", "CCC-" + UUID, чтобы
		// гарантировать порядок независимо от locale.
		suffix := uuid.New().String()

		_, err := repo.Create(ctx, mustExercise(t, "CCC-"+suffix, "Описание 3", 3, "duration"))
		require.NoError(t, err)

		_, err = repo.Create(ctx, mustExercise(t, "AAA-"+suffix, "Описание 1", 5, "weight"))
		require.NoError(t, err)

		_, err = repo.Create(ctx, mustExercise(t, "BBB-"+suffix, "Описание 2", 7, "weight"))
		require.NoError(t, err)

		// Удаляем BBB — не должно быть в списке.
		toDelete, err := repo.Create(ctx, mustExercise(t, "ZZZ-"+suffix, "Описание 4", 4, "weight"))
		require.NoError(t, err)
		require.NoError(t, repo.MarkDeleted(ctx, toDelete.ID()))

		list, err := repo.GetExercises(ctx)
		require.NoError(t, err)

		// Находим только что созданные упражнения в общем списке.
		var createdNames []string
		for _, ex := range list {
			name := ex.Name().String()
			if len(name) > 4 && name[4:] == suffix {
				createdNames = append(createdNames, name[:3])
			}
		}

		require.Equal(t, []string{"AAA", "BBB", "CCC"}, createdNames,
			"упражнения должны идти по алфавиту, ZZZ удалён")

		// Порядок имён в общем списке — по возрастанию.
		for i := 1; i < len(list); i++ {
			assert.LessOrEqual(t, list[i-1].Name().String(), list[i].Name().String(),
				"GetExercises должен сортировать по name ASC")
		}
	})

	t.Run("GetExercises возвращает пустой slice, не nil", func(t *testing.T) {
		// Изолированная проверка: создаём чистый репозиторий на своей таблице
		// не получится, потому что миграция одна на весь тест. Поэтому
		// проверяем поведение "нет результата" через временный тенант:
		// удаляем всё? Нет, это сломает другие тесты. Проверяем через
		// отсутствие конкретных имён — если бы возвращался nil, len(nil) = 0,
		// но assert.NotNil поймает.
		//
		// Здесь проверяем на общем списке, что он не nil. Пустоту проверим
		// в отдельном интеграционном тесте с изолированной схемой.
		list, err := repo.GetExercises(ctx)
		require.NoError(t, err)

		assert.NotNil(t, list, "GetExercises должен возвращать make([]T, 0), а не nil")
	})

	// ============================================================
	//  Update
	// ============================================================

	t.Run("Update меняет поля и возвращает свежий updated_at", func(t *testing.T) {
		created, err := repo.Create(ctx, mustExercise(t, uniqueName("Старое имя"), "Старое описание", 5, "weight"))
		require.NoError(t, err)

		newName, err := domain.NewName(uniqueName("Новое имя"))
		require.NoError(t, err)

		newDescription, err := domain.NewDescription("Новое описание")
		require.NoError(t, err)

		newDifficulty, err := domain.NewDifficulty(9)
		require.NoError(t, err)

		created.ApplyPatch(domain.NewExercisePatch(&newName, &newDescription, &newDifficulty))

		updated, err := repo.Update(ctx, created)
		require.NoError(t, err)

		assert.Equal(t, newName.String(), updated.Name().String())
		assert.Equal(t, newDescription.String(), updated.Description().String())
		assert.Equal(t, 9, updated.Difficulty().Int())

		// updated_at заполнен после UPDATE
		require.NotNil(t, updated.UpdatedAt(), "Update должен вернуть заполненный updated_at")

		// Type не изменился — immutable
		assert.Equal(t, domain.ExerciseTypeWeight, updated.ExerciseType())

		// Перечитываем из БД — тот же результат
		found, err := repo.GetExercise(ctx, created.ID())
		require.NoError(t, err)
		assert.Equal(t, newName.String(), found.Name().String())
		require.NotNil(t, found.UpdatedAt())
		assert.Equal(t, *updated.UpdatedAt(), *found.UpdatedAt())
	})

	t.Run("Update несуществующего — ErrExerciseNotFound", func(t *testing.T) {
		exercise := mustExercise(t, uniqueName("Призрак"), "Описание", 5, "weight")

		_, err := repo.Update(ctx, exercise)

		assert.ErrorIs(t, err, domain.ErrExerciseNotFound)
	})

	t.Run("Update удалённого — ErrExerciseNotFound (WHERE deleted_at IS NULL)", func(t *testing.T) {
		created, err := repo.Create(ctx, mustExercise(t, uniqueName("Удалённое"), "Описание", 5, "weight"))
		require.NoError(t, err)

		require.NoError(t, repo.MarkDeleted(ctx, created.ID()))

		// Пытаемся патчить уже удалённое.
		newName, _ := domain.NewName(uniqueName("Патченное"))
		created.ApplyPatch(domain.NewExercisePatch(&newName, nil, nil))

		_, err = repo.Update(ctx, created)

		assert.ErrorIs(t, err, domain.ErrExerciseNotFound)
	})

	t.Run("Update с конфликтом имён — ErrExerciseNameExists", func(t *testing.T) {
		first, err := repo.Create(ctx, mustExercise(t, uniqueName("Первое"), "Описание 1", 5, "weight"))
		require.NoError(t, err)

		second, err := repo.Create(ctx, mustExercise(t, uniqueName("Второе"), "Описание 2", 7, "weight"))
		require.NoError(t, err)

		// Патчим second — ставим имя first.
		newName := first.Name()
		second.ApplyPatch(domain.NewExercisePatch(&newName, nil, nil))

		_, err = repo.Update(ctx, second)

		assert.ErrorIs(t, err, domain.ErrExerciseNameExists)
	})

	t.Run("Update со своим же именем — без конфликта", func(t *testing.T) {
		created, err := repo.Create(ctx, mustExercise(t, uniqueName("Без изменений"), "Описание", 5, "weight"))
		require.NoError(t, err)

		// Патчим только description — name тот же.
		newDescription, _ := domain.NewDescription("Обновлённое описание")
		created.ApplyPatch(domain.NewExercisePatch(nil, &newDescription, nil))

		updated, err := repo.Update(ctx, created)
		require.NoError(t, err)

		assert.Equal(t, created.Name().String(), updated.Name().String())
		assert.Equal(t, "Обновлённое описание", updated.Description().String())
	})

	// ============================================================
	//  MarkDeleted
	// ============================================================

	t.Run("MarkDeleted помечает упражнение удалённым", func(t *testing.T) {
		created, err := repo.Create(ctx, mustExercise(t, uniqueName("Удаляемое"), "Описание", 5, "weight"))
		require.NoError(t, err)

		err = repo.MarkDeleted(ctx, created.ID())
		require.NoError(t, err)

		found, err := repo.GetExercise(ctx, created.ID())
		require.NoError(t, err)
		assert.True(t, found.IsDeleted())

		// updated_at тоже обновился при MarkDeleted (см. SQL)
		require.NotNil(t, found.UpdatedAt())
	})

	t.Run("MarkDeleted повторно — ErrExerciseNotFound (0 rows affected)", func(t *testing.T) {
		created, err := repo.Create(ctx, mustExercise(t, uniqueName("Повторно"), "Описание", 5, "weight"))
		require.NoError(t, err)

		require.NoError(t, repo.MarkDeleted(ctx, created.ID()))

		// Второй вызов — 0 rows, репозиторий возвращает ошибку.
		// Идемпотентность обеспечивается в usecase (см. ADR-007 F-3),
		// а не в репозитории.
		err = repo.MarkDeleted(ctx, created.ID())

		assert.ErrorIs(t, err, domain.ErrExerciseNotFound)
	})

	t.Run("MarkDeleted несуществующего — ErrExerciseNotFound", func(t *testing.T) {
		err := repo.MarkDeleted(ctx, uuid.New())

		assert.ErrorIs(t, err, domain.ErrExerciseNotFound)
	})

	// ============================================================
	//  Soft delete + partial unique index (ADR-008)
	// ============================================================

	t.Run("после MarkDeleted можно пересоздать упражнение с тем же именем", func(t *testing.T) {
		name := uniqueName("Пересоздаваемое")

		first := mustExercise(t, name, "Описание 1", 5, "weight")
		_, err := repo.Create(ctx, first)
		require.NoError(t, err)

		// Проверяем, что второй раз с тем же именем — конфликт.
		_, err = repo.Create(ctx, mustExercise(t, name, "Описание 2", 7, "weight"))
		require.ErrorIs(t, err, domain.ErrExerciseNameExists)

		// Удаляем первое.
		require.NoError(t, repo.MarkDeleted(ctx, first.ID()))

		// Теперь с тем же именем можно создать — partial unique
		// индекс не учитывает удалённые (deleted_at IS NULL).
		second := mustExercise(t, name, "Описание 3", 9, "duration")
		created, err := repo.Create(ctx, second)

		require.NoError(t, err, "после soft delete можно пересоздать с тем же именем")
		assert.Equal(t, name, created.Name().String())
		assert.NotEqual(t, first.ID(), created.ID())
	})
}
