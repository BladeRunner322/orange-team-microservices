# ADR-002: `user_workout_score`

## Контекст

В монолите поле `app.users.user_workout_score` — денормализация. Пересчитывалось при каждом изменении тренировки:

```sql
UPDATE app.users
SET user_workout_score = (
    SELECT COALESCE(SUM(workout_score), 0)
    FROM app.workouts
    WHERE user_id = $1 AND status = 'completed'
)
```

Триггеры пересчёта:

- завершение/изменение workout exercise (`recalculateScore`)
- патч тренировки (`PatchWorkout`)
- удаление тренировки (`DeleteWorkout`)

Транзакции вокруг этих операций **не было** — то есть в монолите уже существовал риск рассинхрона `workout_score` и `user_workout_score`.

В микросервисах `workouts` и `users` — разные БД. Прямой `SELECT ... FROM workouts` из Users **невозможен**.

## Решение

`user_workout_score` **не хранится как отдельное поле** ни в одном сервисе.

- **Источник истины:** таблица `workouts` (в Workouts-сервисе).
- **Вычисление:** `SUM(workout_score) WHERE user_id = ? AND status = 'completed'`, считает Workouts по запросу.
- **Как отдаётся клиенту:** `GET /users/me` в Gateway делает два **параллельных** gRPC-вызова:
  - `Users.GetProfile(user_id)` → профиль
  - `Workouts.GetUserScore(user_id)` → сумма очков
  - склеивает в один HTTP-ответ

**Никаких событий `workout.completed` в Users.** Никаких distributed-транзакций.

## Последствия

**Плюсы:**

- Нет дрейфа: score всегда вычисляется из источника истины.
- Нет eventual consistency между сервисами.
- Нет межсервисных транзакций.
- Простота: один SQL-запрос в Workouts.
- Логика совпадает с монолитом (SUM по завершённым).

**Минусы:**

- `GET /users/me` делает 2 gRPC-вызова вместо одного. Latency = `max(t_users, t_workouts)`, а не сумма, так как параллельно.
- `SUM(workout_score)` — полносканирующий запрос по `workouts`. Нужен индекс `(user_id, status)`.

**Когда пересмотреть:**

- Если `GET /users/me` станет горячим (десятки RPS на пользователя) — кэш в Redis на 30–60 секунд или event-driven кэш в Users.
- Если пользователей станет очень много и SUM начнёт тормозить — материализованное представление или инкрементальное обновление через события.

## Связанные решения

- [ADR-003: Транзакции](003-transactions.md) — прямое следствие этого ADR.