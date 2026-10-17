# ADR-025: Сервис тренировок (Workouts)

## Контекст

Проект развивается от монолита к микросервисам. Уже реализованы Auth, Gateway, Profiles, Exercises, Habits. Следующий — Workouts.

В монолите тренировки жили в таблицах `app.workouts` и `app.workout_exercises`:
- пользователь создаёт тренировку (статус `planned`),
- добавляет в неё упражнения из справочника `app.exercises`,
- отмечает выполненные,
- тренировка переходит в `completed`,
- пересчитывается `workout.workout_score` и агрегат `users.user_workout_score`.

При распиле на микросервисы нужно сохранить эту бизнес-логику, но убрать решения, которые были возможны только в монолите.

**Что отличается от монолита:**

| | Монолит | Микросервисы |
|---|---|---|
| Справочник упражнений | JOIN `app.exercises` в SQL | gRPC-вызов `Exercises.GetExercise(id)` (ADR-008, ADR-009) |
| Профиль пользователя | `app.users` (вес, рост, пол) | gRPC-вызов `Profiles.GetProfile(user_id)` (ADR-001) |
| `user_workout_score` | Хранится в `users` | Не хранится, `SUM` по `workouts` (ADR-002) |
| HTTP-слой | Свой `workouts_transport_http` | Gateway проксирует через gRPC (ADR-009) |
| Транзакции read-then-write | Optimistic locking через `version` | Оставляем optimistic locking — обоснование ниже |

## Решение

### 1. Модель данных

Две таблицы: **workouts** (сами тренировки) и **workout_exercises** (упражнения внутри тренировки). Обе — в `workouts_db`.

**`workouts`:**

| Поле | Тип | Комментарий |
|---|---|---|
| `id` | UUID PK | |
| `version` | INTEGER NOT NULL DEFAULT 0 | Optimistic locking |
| `user_id` | UUID NOT NULL | Владелец. Из metadata |
| `status` | VARCHAR(16) NOT NULL | `planned` / `in_progress` / `completed` / `cancelled` |
| `started_at` | TIMESTAMPTZ | NULL пока не начата |
| `completed_at` | TIMESTAMPTZ | NULL пока не завершена |
| `workout_score` | INTEGER NOT NULL DEFAULT 0 | Пересчитывается при мутации упражнений |
| `intensity` | INTEGER | NULL, 1–10, задаёт пользователь |
| `personal_score_coefficient` | INTEGER NOT NULL | Снапшот на момент создания (ADR-001) |
| `created_at` | TIMESTAMPTZ NOT NULL DEFAULT NOW() | |
| `updated_at` | TIMESTAMPTZ NOT NULL DEFAULT NOW() | |

**`workout_exercises`:**

| Поле | Тип | Комментарий |
|---|---|---|
| `id` | UUID PK | |
| `version` | INTEGER NOT NULL DEFAULT 0 | Optimistic locking |
| `workout_id` | UUID NOT NULL | FK → `workouts.id`, `ON DELETE CASCADE` |
| `exercise_id` | UUID NOT NULL | Ссылка на Exercises **без cross-service FK** (ADR-004) |
| `difficulty` | SMALLINT NOT NULL | **Снапшот** `difficulty` упражнения на момент добавления (ADR-008) |
| `weight` | INTEGER | Для weight-типа |
| `sets` | INTEGER | Для weight-типа |
| `reps` | INTEGER | Для weight-типа |
| `duration` | INTEGER | Для duration-типа, ≤ 86400 (24 часа) |
| `completed` | BOOLEAN NOT NULL DEFAULT false | |
| `exercise_load` | INTEGER NOT NULL DEFAULT 0 | Вычисляется доменом, снапшот |
| `created_at` | TIMESTAMPTZ NOT NULL DEFAULT NOW() | |
| `updated_at` | TIMESTAMPTZ NOT NULL DEFAULT NOW() | |

**Индексы:**
- `idx_workouts_user_id` на `workouts(user_id)` — для `GetWorkouts` и `GetUserScore`.
- `idx_workouts_user_status` на `workouts(user_id, status)` — для `GetUserScore` (только `completed`).
- `idx_workout_exercises_workout_id` на `workout_exercises(workout_id)` — для `GetWorkoutExercises`.

**Почему `exercise_id` без FK.** Ссылка между сервисами. Exercises живёт в своей БД, в `workouts_db` его нет. Целостность обеспечивается на уровне usecase: перед INSERT Workouts зовёт `Exercises.GetExercise(id)` — если 404, возвращает `NotFound`.

**Почему `exercise_load` хранится, а не считается на лету.** Это снапшот на момент отметки. Если пользователь изменит `weight`/`reps` в упражнении, старые расчёты не должны пересчитываться ретроактивно — тот же принцип, что `difficulty` в ADR-008. Пересчёт `workout_score` использует уже зафиксированный `exercise_load`.

**Почему `difficulty` снапшотится в `workout_exercises`, а не читается live из Exercises.** ADR-008 фиксирует: `difficulty` — снапшот. Если admin завтра поменяет сложность упражнения с 5 на 9, старые тренировки не должны пересчитываться — пользователь уже получил score по старой сложности. Снапшот в `workout_exercises.difficulty` гарантирует, что `recalculateScore` использует значение на момент создания упражнения в тренировке.

**Побочный плюс:** `recalculateScore` не зовёт `Exercises.GetExercise` для `difficulty`. Единственные вызовы Exercises — при `CreateWorkoutExercise` (валидация `type` и чтение `difficulty` для снапшота) и `PatchWorkoutExercise` (валидация `type`). N+1 из `recalculateScore` исчезает.

### 2. Score и коэффициент

**Формула score** (переносится из монолита без изменений, `domain/score.go`):

```
K = coefficient / 100_000
effective_load = Σ (exercise_load_i × difficulty_i) для completed упражнений
score = min(10_000, round(10_000 × K × effective_load / (1 + K × effective_load)))
```

`exercise_load_i` и `difficulty_i` берутся из `workout_exercises` (обе колонки — снапшоты). `recalculateScore` не зовёт Exercises.

Реализация — целочисленная (`saturatedScore`, `addLoadWithLimit`), чтобы избежать floating-point погрешностей и переполнения. Константы:
- `maximumWorkoutScore = 10_000`
- `scoreSaturationScale = 100_000`
- `minimumExerciseDifficulty = 1`, `maximumExerciseDifficulty = 10`
- `minimumScoreCoefficient = 1`, `maximumScoreCoefficient = 10`

**Коэффициент** (переносится из монолита, `domain/coefficient.go`):

```
если профиль не заполнен (weight/height = NULL) → 1
BMI = weight_kg / (height_m ^ 2)
coeff = clamp(1 + (BMI - 18) × 9 / 17, 1, 10)
если sex = male → coeff += 0.5, cap 10
```

**Снапшот.** Коэффициент вычисляется **один раз** при `CreateWorkout` и сохраняется в `workouts.personal_score_coefficient`. Дальнейшие изменения профиля не влияют на старые тренировки (принцип из ADR-001).

**`exercise_load`** (в `domain/workout_exercise.go`):
- weight-тип: `weight × reps × sets`
- duration-тип: `duration`

Считается в конструкторе `NewWorkoutExercise` и при `ApplyPatch` для полей, влияющих на нагрузку.

### 3. Optimistic locking (`version`), не пессимистичный `FOR UPDATE`

**Отличие от ADR-007 F-2.** В Profiles и Exercises мы выбрали `SELECT ... FOR UPDATE`: патчи редкие, ресурс админский/личный, ждать 200мс не больно. В Workouts другая природа:

- Патчи частые (пользователь тренируется в реальном времени).
- `recalculateScore` — это read-modify-write: читает `workout_exercises`, зовёт `Exercises.GetExercise` для difficulty (N gRPC-вызовов), пишет `workout.workout_score`. Под `FOR UPDATE` workout был бы заблокирован на всё время gRPC-вызовов.
- Конфликты редкие (один пользователь в одном окне).

**Как работает:**
- `UPDATE app.workouts SET ..., version = version + 1 WHERE id = $1 AND version = $2`.
- Если `RowsAffected == 0` → `ErrConflict`.
- Gateway возвращает `409 Conflict`. Клиент перечитывает и повторяет.

**Тот же паттерн для `workout_exercises`.**

### 3.1. Транзакция вокруг мутации и пересчёта score

Любая операция, которая меняет `workout_exercises` **и** пересчитывает `workouts.workout_score`, выполняется в **одной транзакции** через `WorkoutRepository.WithTx`.

Порядок внутри `WithTx`:
1. `UPDATE workout_exercises` (с version) — если 0 rows, `ErrConflict`.
2. `recalculateScore` → `UPDATE workouts` (с version) — если 0 rows, `ErrConflict`.
3. Commit.

**Почему без транзакции плохо.** Между шагом 1 и 2 другой клиент может патчить эту же тренировку. Оба читают `workout_exercises`, оба считают score, оба пытаются обновить `workouts` с одним `version`. Один выигрывает, второй получает 409 — но **его патч упражнения уже сохранён** (шаг 1 закоммичен). Клиент видит ошибку, хотя данные изменились, а `workout_score` рассинхронизирован с реальностью до следующей мутации.

**С транзакцией** — либо обе операции прошли, либо ни одна. При конфликте откат обеих, клиент перечитывает и повторяет. Инвариант «`workout_score` всегда соответствует содержимому `workout_exercises`» держится.

**Что это требует от репозиториев:**
- `WorkoutRepository.WithTx(ctx, fn func(WorkoutRepository) error) error` — по образцу `HabitRepository`/`ExerciseRepository`/`ProfileRepository`.
- `WorkoutExerciseRepository` получает метод `WithTx` с **тем же** `*postgres.Tx` (репозитории работают в одной транзакции через общий `tx`).
- В bootstrap оба репозитория получают общий `*postgres.PgxPool`.

### 4. Агрегат vs два репозитория

`workouts` и `workout_exercises` — два агрегата. `workout_exercises` ссылается на `workouts` через FK, но это не «один большой агрегат». Следствия:

- `WorkoutRepository` и `WorkoutExerciseRepository` — два разных репозитория.
- `recalculateScore` — не метод репозитория, а helper в `usecases/common.go`: он координирует чтение из обоих репозиториев и обновление `workouts.workout_score`.
- В монолите `WorkoutExercisesRepository.UpdateWorkoutScore` писал в `workouts` из репозитория `workout_exercises` — нарушение границ. В новом проекте: `WorkoutRepository.Update(ctx, workout)` с version.

**Почему не один репозиторий.** У `workouts` есть свои инварианты (status machine), у `workout_exercises` — свои (тип-специфичные поля, `exercise_load`). Смешивание размывает ответственность и мешает тестировать.

### 5. `user_workout_score` не хранится

По ADR-002: `SUM(workout_score) WHERE user_id = ? AND status = 'completed'`, считается на лету.

В Workouts это отдельный gRPC-метод **`GetUserScore(user_id) → int`**. Gateway в `GET /users/me` параллельно зовёт `Profiles.GetProfile` и `Workouts.GetUserScore`, склеивает в один ответ (ADR-002, ADR-009).

`UpdateUserWorkoutScore` из монолита **не переносится**.

### 6. Зависимости (sync gRPC)

Habits не зовёт другие сервисы (ADR-023). Workouts — зовёт:

| Кто | Кого | Когда | Зачем |
|---|---|---|---|
| Workouts | Profiles | `CreateWorkout` | `GetProfile(user_id)` → коэффициент |
| Workouts | Exercises | `CreateWorkoutExercise`, `PatchWorkoutExercise`, `recalculateScore` | `GetExercise(id)` → `type` (валидация) и `difficulty` (score) |

Циклических зависимостей нет (ADR-009). Gateway остаётся единственной точкой входа.

**Профиль может быть не заполнен** (lazy-create, ADR-001). Тогда коэффициент = 1 — тренировка создаётся, но score будет минимальным. Gateway в этом случае вернёт `profile_completed=false` на `/users/me` — клиент сам решает, блокировать создание тренировки или нет.

### 7. gRPC-методы

```
WorkoutsService:
  CreateWorkout(CreateWorkoutRequest) returns (Workout)
  GetWorkout(GetWorkoutRequest) returns (Workout)
  GetWorkouts(GetWorkoutsRequest) returns (GetWorkoutsResponse)
  PatchWorkout(PatchWorkoutRequest) returns (Workout)
  DeleteWorkout(DeleteWorkoutRequest) returns (DeleteWorkoutResponse)

  CreateWorkoutExercise(CreateWorkoutExerciseRequest) returns (WorkoutExercise)
  GetWorkoutExercises(GetWorkoutExercisesRequest) returns (GetWorkoutExercisesResponse)
  PatchWorkoutExercise(PatchWorkoutExerciseRequest) returns (WorkoutExercise)
  DeleteWorkoutExercise(DeleteWorkoutExerciseRequest) returns (DeleteWorkoutExerciseResponse)

  GetUserScore(GetUserScoreRequest) returns (GetUserScoreResponse)
```

**`user_id` приходит из metadata.** Gateway пробрасывает `x-user-id`, Workouts читает из context (как Habits/Profiles/Exercises).

**`GetUserScore`** — для агрегации в Gateway (`GET /users/me`).

**Права доступа:** только для **своего** `user_id`. Gateway кладёт `user_id` из JWT в metadata, Workouts проверяет: `metadata.user_id == request.user_id`. Если не совпало — `PermissionDenied` (HTTP 403).

Чужие очки через этот метод получить нельзя. Leaderboard (когда появится) будет читать batch-методом `GetUsersScores(ids)` с другим auth (service-to-service, без user metadata) — отдельный ADR.

### 8. HTTP-маршруты через Gateway

| Метод | Путь | Назначение |
|---|---|---|
| GET | `/workouts` | Список тренировок |
| POST | `/workouts` | Создать |
| GET | `/workouts/{workoutId}` | Одна тренировка |
| PATCH | `/workouts/{workoutId}` | Патч (status, started_at, completed_at, intensity) |
| DELETE | `/workouts/{workoutId}` | Удалить |
| POST | `/workouts/{workoutId}/exercises` | Добавить упражнение |
| GET | `/workouts/{workoutId}/exercises` | Список упражнений тренировки |
| PATCH | `/workouts/{workoutId}/exercises/{exerciseId}` | Патч упражнения |
| DELETE | `/workouts/{workoutId}/exercises/{exerciseId}` | Удалить упражнение |

Плюс **агрегация `/users/me`** в Gateway: параллельно `Profiles.GetMyProfile` + `Workouts.GetUserScore`. Это не отдельный маршрут, а изменение существующего (сейчас `/users/me` возвращает только профиль).

#### Агрегация `/users/me` и fallback

Хендлер Gateway вызывает два downstream параллельно через `errgroup.WithContext`.

**Профиль (Profiles) — обязателен:**
- Profiles 5xx → весь `/users/me` возвращает 503. Без профиля ответ бессмысленен.

**Очки (Workouts) — опционален, fallback:**
- Workouts успех → `workout_score: <int>`, `workout_score_available: true`.
- Workouts 5xx / CB открыт / timeout → `workout_score: 0`, `workout_score_available: false`. Профиль всё равно отдаётся, HTTP-статус 200.

**Обоснование:**
- `/users/me` — самый частый эндпоинт. Пользователь должен видеть свой профиль, даже если Workouts временно лёг.
- `workout_score_available: false` сигнализирует клиенту «score недоступен, не показывай 0 как реальный». Клиент сам решает: скрыть поле или показать `—`.
- Circuit Breaker на клиенте к Workouts → мгновенный fail-fast, fallback срабатывает без задержки.

**Fallback — только для этого эндпоинта.** Остальные маршруты Gateway (`/login`, `/workouts`, `/exercises`, `/habits`) fail-fast при ошибке downstream. Это осознанно: там fallback даст ложное «успешно», что хуже, чем явная ошибка.

### 9. Валидация

- `status` — только `planned`, `in_progress`, `completed`, `cancelled`. Переходы — по state machine в `domain/workout.go`.
- `intensity` — 1–10 или NULL.
- `weight`, `sets`, `reps` — обязательны для weight-типа, NULL для duration-типа.
- `duration` — обязателен для duration-типа, NULL для weight-типа. **Максимум 86400 (24 часа).** Валидация в VO `Duration`.
- `completed` — можно поставить только когда тренировка в `in_progress`. Снять можно всегда (кроме `completed`/`cancelled` статуса тренировки).
- Нельзя перевести тренировку в `completed`, если ни одно упражнение не отмечено completed.

**Инварианты `started_at` / `completed_at` — домен ставит автоматически, пользователь не присылает:**

- `status = in_progress` и `started_at == nil` → домен ставит `started_at = time.Now().UTC()`.
- `status = completed` и `completed_at == nil` → домен ставит `completed_at = time.Now().UTC()`.
- Уже заполненное поле не перезаписывается (идемпотентность при повторном патче).

В proto и HTTP-DTO полей `started_at` / `completed_at` в `PatchWorkout` **нет** — они управляются только доменом.

Валидация типа-специфичных полей — `WorkoutExercise.ValidateForWorkoutExerciseType(type)`. Это доменное правило, тестируется в `domain`.

### 10. Идентификаторы и порты

По ADR-005:
- gRPC — **50055** (Docker) / **50065** (локально).
- HTTP (health/ready/metrics) — **8085** (Docker) / **8095** (локально).
- PostgreSQL — **5436** на хосте.

### 11. Что НЕ делаем

- **События `workout.completed` в Kafka** — ADR-024, после Notification. Сначала Workouts стабильно работает через sync gRPC.
- **Batch `Exercises.GetExercisesByIds`** — для MVP N+1 при `recalculateScore` приемлем. Триггер для batch — когда `recalculateScore` станет горячим.
- **Пагинация `GetWorkouts`** — MVP возвращает весь список. Триггер: 100+ тренировок у одного пользователя и p95 > 500ms. Тогда — `page`/`limit` в proto.
- **Кэш `GetUserScore` в Redis** — по ADR-002, при росте нагрузки. Триггер — `GET /users/me` станет bottleneck.
- **Пессимистичные блокировки** — обоснование в п.3.
- **`UpdateUserWorkoutScore`** — не переносим, `user_workout_score` не хранится (ADR-002).

### 12. Совместимость с ADR-024 (будущая Kafka)

Workouts разрабатывается **без** публикации событий — так закрывается базовая бизнес-логика до момента, когда Kafka и Notification будут готовы.

**Когда придёт ADR-024**, изменения будут только аддитивными:

- Новые файлы: `ports/event_publisher.go`, `infrastructure/kafka/publisher.go`, типы событий.
- Одна строка в `PatchWorkout` (при переходе в `completed`): `publisher.Publish(ctx, WorkoutCompleted{...})`.
- Инфраструктура: `config.go`, `compose.app.yml`, `.env.example`, `go.mod`.

**Что НЕ меняется при добавлении Kafka:**
- `domain/` — score, коэффициент, state machine, `exercise_load` остаются как есть.
- `infrastructure/postgres_repo/` — схема БД не меняется.
- `interfaces/workoutsgrpc/` — `GetUserScore` остаётся синхронным.
- 8 из 9 usecase — без изменений.
- Sync gRPC к Profiles и Exercises — не заменяется на async (ответ нужен в момент операции).

**Добавление Outbox** (если понадобится по триггеру ADR-024) — замена реализации `EventPublisher`, usecase не трогаются. Интерфейс `Publish(ctx, event) error` стабилен.

## Последствия

**Плюсы:**

- Score-формула и коэффициент переносятся из монолита без потери качества — они уже проверены на реальных пользователях.
- Optimistic locking развязывает `recalculateScore` от блокировок — gRPC-вызовы в Exercises не держат workout под замком.
- `exercise_load` — снапшот: пересчёт score при изменении справочника упражнений невозможен, старая тренировка остаётся «как было».
- `user_workout_score` не хранится → нет дрейфа, Gateway считает на лету.
- Ссылки на Exercises/Profiles — по id через gRPC, никаких cross-service FK.

**Минусы:**

- **N+1 при `GetWorkoutExercises` и `recalculateScore`.** Одна тренировка на 10 упражнений = 10 вызовов `Exercises.GetExercise`. Для MVP — приемлемо (latency ~10-30мс с CB/Retry). При росте — batch-метод.
- **Optimistic locking даёт 409 конфликты.** При параллельном патче одного упражнения (два окна) второй клиент получит конфликт. Для тренировки это редкий сценарий, клиент перечитывает.
- **`GetUserScore` — это `SUM` по всем completed-тренировкам пользователя.** На больших историях — медленно. Индекс `(user_id, status)` помогает, но при миллионах тренировок понадобится материализованное представление.
- **Профиль может быть не заполнен** — тренировка создастся, но score минимальный. Не блокируем создание — клиент сам решает.

**Когда пересмотреть:**

- Если `recalculateScore` станет горячим → batch `GetExercisesByIds` в Exercises (триггер — p95 > 100мс).
- Если `GetUserScore` тормозит → кэш в Redis на 30–60с, или материализованное представление (триггер — p95 > 100мс).
- Если конфликты 409 станут частыми → пересмотреть optimistic locking на пессимистичный (триггер — >5% запросов в 409).
- Если появится публичный лидерборд → Leaderboard перестаёт звать `Workouts.GetUserScore` синхронно, начинается кэш/события (ADR-024, Уровень 3).

## Связанные решения

- [ADR-001: Профиль пользователя](001-profile.md) — снапшот коэффициента, lazy-create.
- [ADR-002: user_workout_score](002-user-workout-score.md) — не храним, SUM на лету.
- [ADR-004: Паттерны микросервисов](004-microservices-patterns.md) — Database per Service, sync gRPC.
- [ADR-005: Распределение портов](005-port-allocation.md) — 50055 / 8085 / 5436.
- [ADR-007: Известные проблемы](007-known-issues.md) — F-2 (пессимистичная блокировка в Profiles/Exercises — здесь другой выбор, обосновано).
- [ADR-008: Жизненный цикл упражнения](008-exercise-lifecycle.md) — `difficulty` snapshot, `type` immutable, N+1 принят осознанно.
- [ADR-009: Карта зависимостей](009-service-dependencies.md) — Workouts → Profiles, Workouts → Exercises.
- [ADR-011: Naming conventions](011-naming-and-code-style.md) — стиль кода.
- [ADR-023: Сервис привычек](023-habits.md) — ориентир по структуре.

## Статус

**В работе.**

- ⬜ Domain (VO, сущности, score, coefficient) — покрыт unit-тестами.
- ⬜ Infrastructure (postgres_repo, mappers, integration tests).
- ⬜ Usecases (9 usecase + common + recalculateScore + моки + unit-тесты).
- ⬜ gRPC-сервер + bootstrap + config + main.
- ⬜ Compose + Taskfile + `.env.enc` + CI matrix.
- ⬜ Gateway-хендлеры + маршруты + агрегация `/users/me`.