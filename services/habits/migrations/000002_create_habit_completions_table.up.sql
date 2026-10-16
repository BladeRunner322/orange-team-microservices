CREATE TABLE IF NOT EXISTS habits.habit_completions (
    habit_id   UUID NOT NULL,
    date       DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Составной PK (habit_id, date) даёт два эффекта сразу:
    --   1. Гарантия «одна отметка на привычку в день» — вторая вставка
    --      за ту же дату упадёт с unique violation.
    --   2. Индекс по (habit_id, date), который покрывает ListByHabit,
    --      IsCompleted, Uncomplete и Streak. Отдельные индексы не нужны.
    PRIMARY KEY (habit_id, date),

    -- ON DELETE CASCADE: при физическом удалении привычки отметки уедут
    -- следом. При soft delete (UPDATE ... SET deleted_at) НЕ срабатывает.
    CONSTRAINT fk_habit_completions_habit
        FOREIGN KEY (habit_id)
        REFERENCES habits.habits (id)
        ON DELETE CASCADE
);