CREATE SCHEMA IF NOT EXISTS exercises;

CREATE TABLE IF NOT EXISTS exercises.exercises (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    description VARCHAR(1000) NOT NULL,
    difficulty  SMALLINT NOT NULL,
    type        VARCHAR(16) NOT NULL,
    deleted_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ
);

-- Partial unique index: имя уникально только среди активных упражнений.
-- Удалённые (deleted_at IS NOT NULL) не участвуют — можно пересоздать
-- упражнение с тем же именем после soft delete (см. ADR-008).
CREATE UNIQUE INDEX IF NOT EXISTS idx_exercises_name_active
    ON exercises.exercises (name)
    WHERE deleted_at IS NULL;

-- Для быстрого фильтра активных в GetExercises (deleted_at IS NULL).
CREATE INDEX IF NOT EXISTS idx_exercises_deleted_at
    ON exercises.exercises (deleted_at);

-- difficulty: 1..10 (дублирует валидацию VO — defense in depth).
ALTER TABLE exercises.exercises ADD CONSTRAINT chk_exercises_difficulty
    CHECK (difficulty BETWEEN 1 AND 10);

-- type: только weight или duration.
ALTER TABLE exercises.exercises ADD CONSTRAINT chk_exercises_type
    CHECK (type IN ('weight', 'duration'));

-- name: 3..100 символов (как в domain.Name).
ALTER TABLE exercises.exercises ADD CONSTRAINT chk_exercises_name_length
    CHECK (char_length(name) BETWEEN 3 AND 100);

-- description: 1..1000 символов (как в domain.Description).
ALTER TABLE exercises.exercises ADD CONSTRAINT chk_exercises_description_length
    CHECK (char_length(description) BETWEEN 1 AND 1000);
