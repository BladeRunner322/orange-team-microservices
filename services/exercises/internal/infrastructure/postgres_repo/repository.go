// Package postgres_repo — реализация ports.Repository для PostgreSQL.
package postgres_repo

// TODO: структура Repository с пулом, конструктор NewRepository(pool).
// TODO: методы GetExercises, GetExercise, Create, Update.
//       GetExercises — WHERE deleted_at IS NULL, ORDER BY id.
//       GetExercise  — без фильтра (включая удалённые).
//       Update       — SET ... + updated_at = NOW(), RETURNING все поля.
