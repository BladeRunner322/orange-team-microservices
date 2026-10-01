package usecases

// PatchExercise — usecase для обновления упражнения (admin-only).
//
// TODO: структура + конструктор + Execute(ctx, id, patch)
// Патч содержит только name, description, difficulty (type immutable — ADR-008).
// Если упражнение удалено — вернуть ErrExerciseNotFound.
