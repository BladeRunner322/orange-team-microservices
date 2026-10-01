package domain

// Exercise — доменная сущность упражнения.
//
// Поля (приватные, доступ через геттеры):
//   - id
//   - name
//   - description
//   - difficulty (1..10)
//   - exerciseType (weight | duration)
//   - isDeleted (soft delete, см. ADR-008)
//   - createdAt, updatedAt
//
// TODO: конструкторы NewExercise, RestoreExercise
// TODO: методы ApplyPatch(patch ExercisePatch), Delete(), геттеры
// TODO: метод MarkDeleted() для soft delete
