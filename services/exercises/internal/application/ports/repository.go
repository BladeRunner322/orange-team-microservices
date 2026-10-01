// Package ports — интерфейсы, которые domain требует от инфраструктуры.
package ports

// TODO: интерфейс Repository с методами:
//   - GetExercises(ctx) ([]domain.Exercise, error)         // только активные
//   - GetExercise(ctx, id) (domain.Exercise, error)        // включая удалённые
//   - Create(ctx, exercise) (domain.Exercise, error)
//   - Update(ctx, exercise) (domain.Exercise, error)       // для PATCH и soft delete
