CREATE TABLE IF NOT EXISTS workouts.workout_exercises (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version       INTEGER NOT NULL DEFAULT 0,
    workout_id    UUID NOT NULL,
    exercise_id   UUID NOT NULL,
    difficulty    SMALLINT NOT NULL,
    weight        INTEGER,
    sets          INTEGER,
    reps          INTEGER,
    duration      INTEGER,
    completed     BOOLEAN NOT NULL DEFAULT false,
    exercise_load INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_workout_exercises_workout
        FOREIGN KEY (workout_id)
        REFERENCES workouts.workouts (id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_workout_exercises_workout_id
    ON workouts.workout_exercises (workout_id);

ALTER TABLE workouts.workout_exercises ADD CONSTRAINT chk_workout_exercises_difficulty
    CHECK (difficulty BETWEEN 1 AND 10);

ALTER TABLE workouts.workout_exercises ADD CONSTRAINT chk_workout_exercises_weight
    CHECK (weight IS NULL OR (weight >= 1 AND weight <= 1000));

ALTER TABLE workouts.workout_exercises ADD CONSTRAINT chk_workout_exercises_sets
    CHECK (sets IS NULL OR (sets >= 1 AND sets <= 100));

ALTER TABLE workouts.workout_exercises ADD CONSTRAINT chk_workout_exercises_reps
    CHECK (reps IS NULL OR (reps >= 1 AND reps <= 1000));

ALTER TABLE workouts.workout_exercises ADD CONSTRAINT chk_workout_exercises_duration
    CHECK (duration IS NULL OR (duration >= 1 AND duration <= 86400));