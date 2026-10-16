CREATE SCHEMA IF NOT EXISTS habits;

CREATE TABLE IF NOT EXISTS habits.habits (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    deleted_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ
);

-- Partial unique index: имя уникально среди активных привычек одного пользователя.
-- Удалённые (deleted_at IS NOT NULL) не участвуют.
CREATE UNIQUE INDEX IF NOT EXISTS idx_habits_user_name_active
    ON habits.habits (user_id, name)
    WHERE deleted_at IS NULL;

-- Partial index: список активных привычек пользователя — самый частый запрос.
CREATE INDEX IF NOT EXISTS idx_habits_user_id_active
    ON habits.habits (user_id)
    WHERE deleted_at IS NULL;

-- Для быстрого фильтра удалённых.
CREATE INDEX IF NOT EXISTS idx_habits_deleted_at
    ON habits.habits (deleted_at);

-- name: 1..100 символов.
ALTER TABLE habits.habits ADD CONSTRAINT chk_habits_name_length
    CHECK (char_length(name) BETWEEN 1 AND 100);

-- description: 0..500 символов (пустое описание допустимо).
ALTER TABLE habits.habits ADD CONSTRAINT chk_habits_description_length
    CHECK (char_length(description) BETWEEN 0 AND 500);