CREATE SCHEMA IF NOT EXISTS workouts;

CREATE TABLE IF NOT EXISTS workouts.workouts (
    id                         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version                    INTEGER NOT NULL DEFAULT 0,
    user_id                    UUID NOT NULL,
    status                     VARCHAR(16) NOT NULL DEFAULT 'planned',
    started_at                 TIMESTAMPTZ,
    completed_at               TIMESTAMPTZ,
    workout_score              INTEGER NOT NULL DEFAULT 0,
    intensity                  INTEGER,
    personal_score_coefficient INTEGER NOT NULL,
    deleted_at                 TIMESTAMPTZ,
    created_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workouts_user_id_active
    ON workouts.workouts (user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_workouts_user_status_active
    ON workouts.workouts (user_id, status)
    WHERE deleted_at IS NULL;

ALTER TABLE workouts.workouts ADD CONSTRAINT chk_workouts_status
    CHECK (status IN ('planned', 'in_progress', 'completed', 'cancelled'));

ALTER TABLE workouts.workouts ADD CONSTRAINT chk_workouts_intensity
    CHECK (intensity IS NULL OR intensity BETWEEN 1 AND 10);

ALTER TABLE workouts.workouts ADD CONSTRAINT chk_workouts_coefficient
    CHECK (personal_score_coefficient BETWEEN 1 AND 10);