package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestExercise — хелпер: собирает упражнение с заданными полями.
func newTestExercise(
	name, description string,
	difficulty int,
	exerciseType string,
) Exercise {
	n, _ := NewName(name)
	d, _ := NewDescription(description)
	diff, _ := NewDifficulty(difficulty)
	et, _ := NewExerciseType(exerciseType)
	return NewExercise(n, d, diff, et)
}

func TestApplyPatch_EmptyPatch(t *testing.T) {
	exercise := newTestExercise("Жим лёжа", "Базовое упражнение", 5, "weight")
	originalID := exercise.ID()

	// Пустой патч — все поля nil.
	patch := NewExercisePatch(nil, nil, nil)
	exercise.ApplyPatch(patch)

	assert.Equal(t, originalID, exercise.ID())
	assert.Equal(t, "Жим лёжа", exercise.Name().String())
	assert.Equal(t, "Базовое упражнение", exercise.Description().String())
	assert.Equal(t, 5, exercise.Difficulty().Int())
	assert.Equal(t, ExerciseTypeWeight, exercise.ExerciseType())
}

func TestApplyPatch_NameOnly(t *testing.T) {
	exercise := newTestExercise("Старое имя", "Описание", 5, "weight")

	newName, _ := NewName("Новое имя")
	patch := NewExercisePatch(&newName, nil, nil)
	exercise.ApplyPatch(patch)

	assert.Equal(t, "Новое имя", exercise.Name().String())
	assert.Equal(t, "Описание", exercise.Description().String())
	assert.Equal(t, 5, exercise.Difficulty().Int())
}

func TestApplyPatch_DescriptionOnly(t *testing.T) {
	exercise := newTestExercise("Жим лёжа", "Старое описание", 5, "weight")

	newDescription, _ := NewDescription("Новое описание")
	patch := NewExercisePatch(nil, &newDescription, nil)
	exercise.ApplyPatch(patch)

	assert.Equal(t, "Жим лёжа", exercise.Name().String())
	assert.Equal(t, "Новое описание", exercise.Description().String())
	assert.Equal(t, 5, exercise.Difficulty().Int())
}

func TestApplyPatch_DifficultyOnly(t *testing.T) {
	exercise := newTestExercise("Жим лёжа", "Описание", 3, "weight")

	newDifficulty, _ := NewDifficulty(9)
	patch := NewExercisePatch(nil, nil, &newDifficulty)
	exercise.ApplyPatch(patch)

	assert.Equal(t, "Жим лёжа", exercise.Name().String())
	assert.Equal(t, "Описание", exercise.Description().String())
	assert.Equal(t, 9, exercise.Difficulty().Int())
}

func TestApplyPatch_AllFields(t *testing.T) {
	exercise := newTestExercise("Старое имя", "Старое описание", 1, "duration")

	newName, _ := NewName("Новое имя")
	newDescription, _ := NewDescription("Новое описание")
	newDifficulty, _ := NewDifficulty(10)
	patch := NewExercisePatch(&newName, &newDescription, &newDifficulty)
	exercise.ApplyPatch(patch)

	assert.Equal(t, "Новое имя", exercise.Name().String())
	assert.Equal(t, "Новое описание", exercise.Description().String())
	assert.Equal(t, 10, exercise.Difficulty().Int())
	// type не должен измениться — immutable (ADR-008)
	assert.Equal(t, ExerciseTypeDuration, exercise.ExerciseType())
}

func TestApplyPatch_TypeIsImmutable(t *testing.T) {
	// Упражнение типа "weight".
	exercise := newTestExercise("Жим лёжа", "Описание", 5, "weight")

	// Патч без поля type (его нет в ExercisePatch по ADR-008).
	newName, _ := NewName("Новое имя")
	newDescription, _ := NewDescription("Новое описание")
	newDifficulty, _ := NewDifficulty(7)
	patch := NewExercisePatch(&newName, &newDescription, &newDifficulty)
	exercise.ApplyPatch(patch)

	// type остался прежним.
	assert.Equal(t, ExerciseTypeWeight, exercise.ExerciseType())
}

func TestApplyPatch_DoesNotTouchDeletedAtOrCreatedAt(t *testing.T) {
	// Патч не должен влиять на deleted_at / created_at / id —
	// эти поля управляются отдельными операциями.
	name, _ := NewName("Жим")
	description, _ := NewDescription("Описание")
	difficulty, _ := NewDifficulty(5)
	exerciseType, _ := NewExerciseType("weight")
	id := uuid.New()
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	exercise := RestoreExercise(
		id, name, description, difficulty, exerciseType,
		false, createdAt, nil,
	)

	newName, _ := NewName("Обновлённое имя")
	patch := NewExercisePatch(&newName, nil, nil)
	exercise.ApplyPatch(patch)

	assert.Equal(t, id, exercise.ID())
	assert.Equal(t, createdAt, exercise.CreatedAt())
	assert.Nil(t, exercise.UpdatedAt())
	assert.False(t, exercise.IsDeleted())

	// но name изменился
	require.Equal(t, "Обновлённое имя", exercise.Name().String())
}

func TestApplyPatch_OnDeletedExercise(t *testing.T) {
	// Домен не запрещает патчить удалённое — это забота репозитория
	// (WHERE deleted_at IS NULL). Проверяем, что ApplyPatch просто
	// применяет патч к полям, не глядя на isDeleted.
	name, _ := NewName("Жим")
	description, _ := NewDescription("Описание")
	difficulty, _ := NewDifficulty(5)
	exerciseType, _ := NewExerciseType("weight")

	exercise := RestoreExercise(
		uuid.New(), name, description, difficulty, exerciseType,
		true, time.Now().UTC(), nil,
	)

	newName, _ := NewName("Новое имя")
	patch := NewExercisePatch(&newName, nil, nil)
	exercise.ApplyPatch(patch)

	assert.Equal(t, "Новое имя", exercise.Name().String())
	assert.True(t, exercise.IsDeleted(), "isDeleted не должен измениться через ApplyPatch")
}
