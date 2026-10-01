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

-- TODO: constraints
--   - UNIQUE (name) — с учётом soft-delete: либо partial UNIQUE INDEX WHERE deleted_at IS NULL,
--     либо UNIQUE по name как в монолите
--   - CHECK difficulty BETWEEN 1 AND 10
--   - CHECK type IN ('weight', 'duration')
--   - CHECK length(name) BETWEEN 3 AND 100
--   - CHECK length(description) BETWEEN 1 AND 1000
-- TODO: индекс по deleted_at для быстрого фильтра активных
