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
| `deleted_at` | TIMESTAMPTZ | NULL пока не удалено (soft delete) |
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
| `duration` | INTEGER | Для duration-типа, 1 ≤ duration ≤ 86400 (24 часа) |
| `completed` | BOOLEAN NOT NULL DEFAULT false | |
| `exercise_load` | INTEGER NOT NULL DEFAULT 0 | Вычисляется доменом, снапшот |
| `created_at` | TIMESTAMPTZ NOT NULL DEFAULT NOW() | |
| `updated_at` | TIMESTAMPTZ NOT NULL DEFAULT NOW() | |

**Индексы:**
- `idx_workouts_user_id_active` — partial, на `workouts(user_id) WHERE deleted_at IS NULL` — для `GetWorkouts`.
- `idx_workouts_user_status_active` — partial, на `workouts(user_id, status) WHERE deleted_at IS NULL` — для `GetUserScore`.
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

Реализация — целочисленная (`CalculateScore`, `AddLoadWithLimit`, `DivideRoundUp` — публичные; `saturatedScore` — приватная, вызывается только из `CalculateScore`), чтобы избежать floating-point погрешностей и переполнения. Константы:
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

**Когда запускается `recalculateScore`.** Только при мутации `workout_exercises` (`Create` / `Patch` / `Delete`) в тренировках со статусом `planned` или `in_progress`. Для `completed` / `cancelled` мутации запрещены на уровне домена (`Workout.CanModifyWorkoutExercise() == false`), поэтому usecase даже не доходит до `recalculateScore`.

Следствие: `workout.workout_score` меняется только пока тренировка активна. После `completed` он фиксирован — как и `personal_score_coefficient` (снапшот).

### 3. Optimistic locking (`version`), не пессимистичный `FOR UPDATE`

**Отличие от ADR-007 F-2.** В Profiles и Exercises мы выбрали `SELECT ... FOR UPDATE`: патчи редкие, ресурс админский/личный, ждать 200мс не больно. В Workouts другая природа:

- Патчи частые (пользователь тренируется в реальном времени).
- `CreateWorkoutExercise` и `PatchWorkoutExercise` **до** транзакции делают gRPC-вызов `Exercises.GetExercise(id)` (валидация `type`). Под `FOR UPDATE` workout был бы заблокирован на время сетевого вызова в другой сервис — а это непредсказуемо (до таймаута).
- Конфликты редкие (один пользователь в одном окне).

**Как работает:**
- `UPDATE app.workouts SET ..., version = version + 1 WHERE id = $1 AND version = $2`.
- Если `RowsAffected == 0` → `ErrConflict`.
- Gateway возвращает `409 Conflict`. Клиент перечитывает и повторяет.

**Тот же паттерн для `workout_exercises`.**

### 3.1. Транзакция вокруг мутации и пересчёта score

Любая операция, которая меняет `workout_exercises` **и** пересчитывает `workouts.workout_score`, выполняется в **одной транзакции** через `WorkoutRepository.WithTx`.

Порядок внутри `WithTx`:
1. `WorkoutExerciseRepository.Update(ctx, exercise)` — `UPDATE workout_exercises` с version. Если 0 rows — `ErrConflict`, транзакция откатывается.
2. `recalculateScore` — читает `workout_exercises` (уже обновлённые в tx), считает score, вызывает `WorkoutRepository.Update(ctx, workout)` — `UPDATE workouts` с version. Если 0 rows — `ErrConflict`, откат.
3. Commit.

**Почему без транзакции плохо.** Между шагом 1 и 2 другой клиент может патчить эту же тренировку. Оба читают `workout_exercises`, оба считают score, оба пытаются обновить `workouts` с одним `version`. Один выигрывает, второй получает 409 — но **его патч упражнения уже сохранён** (шаг 1 закоммичен). Клиент видит ошибку, хотя данные изменились, а `workout_score` рассинхронизирован с реальностью до следующей мутации.

**С транзакцией** — либо обе операции прошли, либо ни одна. При конфликте откат обеих, клиент перечитывает и повторяет. Инвариант «`workout_score` всегда соответствует содержимому `workout_exercises`» держится.

**Что это требует от репозиториев:**
- `WorkoutRepository.WithTx(ctx, fn func(WorkoutRepository, WorkoutExerciseRepository) error) error` — открывает транзакцию, создаёт **оба** репозитория, привязанных к одному `*postgres.Tx`, передаёт их в callback. По образцу существующих `WithTx`, но с двумя репозиториями в аргументе (потому что `recalculateScore` ходит в оба).
- Оба репозитория получают `NewWithTx(pool, tx)` — конструктор с транзакцией (или `NewRepository(pool)` для работы без транзакции).
- В bootstrap создаются обычные репозитории на `*postgres.PgxPool`; `WithTx` создаёт временные с `tx` внутри callback.

### 3.2. Soft delete (удаление тренировки)

DELETE тренировки — **soft delete**. Строка не удаляется физически, в `workouts.workouts` ставится `deleted_at = NOW()`.

**Зачем.** Пользователь может случайно удалить тренировку, где уже начислены очки и выполнены упражнения. Физическое удаление необратимо: `user_workout_score` мгновенно падает на `workout_score`, история упражнений теряется. Soft delete сохраняет данные — восстановление возможно (через SQL или будущий `RestoreWorkout`, если понадобится).

**Что меняется в схеме.**
- Колонка `workouts.workouts.deleted_at TIMESTAMPTZ` (NULL — активна).
- Partial-индексы `WHERE deleted_at IS NULL` (см. п.1).

**Что НЕ меняется.**
- `workout_exercises` **не удаляются** при soft delete workout. `ON DELETE CASCADE` срабатывает только при физическом `DELETE` (например, в purge-скрипте, если появится).
- `workout_score` при удалении **не сбрасывается** — остаётся как был. Просто перестаёт учитываться в `GetUserScore` (см. ниже).

**`workout_exercises` — hard delete, не soft.**

В отличие от `workouts`, `workout_exercises` — **компонент агрегата**, а не самостоятельная сущность. Поэтому:

- На `workout_exercise_id` никто не ссылается — ни БД, ни другой сервис. Физическое удаление ничего не ломает.
- История привязана к тренировке целиком, не к отдельному упражнению. `workout_score` пересчитывается при любой мутации: удалил упражнение и добавил обратно — тот же score.
- Терминальные статусы `Workout` (`completed`, `cancelled`) блокируют мутации — финальный состав тренировки неизменяем (`CanModifyWorkoutExercise() == false`).
- При физическом удалении `Workout` (purge, если понадобится) `ON DELETE CASCADE` сносит все его `workout_exercises` — консистентно, без «полуживых» записей.

**Правило проекта (общее для всех сервисов):**

> **Soft delete — для aggregate roots и сущностей со внешними ссылками** (`Exercise` — ADR-008, `Habit` — ADR-023, `Workout` — здесь).
> **Hard delete — для компонентов агрегата** (`WorkoutExercise`).

Это же правило определяет, что у `workout_exercises` нет `deleted_at` в схеме (п.1), нет `isDeleted` в domain (`workout_exercise.go`), нет `MarkDeleted` в репозитории — только `Delete`.

**Поведение по методам репозитория:**

| Метод | Фильтр `deleted_at IS NULL`? | Обоснование |
|---|---|---|
| `WorkoutRepository.GetByID` | **Нет** | Возвращает и удалённые с флагом `is_deleted`. Аналогично `GetExercise` в ADR-008 |
| `WorkoutRepository.GetByUserID` | **Да** | Список тренировок — только активные |
| `WorkoutRepository.Update` | **Да** | Нельзя патчить удалённую |
| `WorkoutRepository.MarkDeleted` | **Да** | Уже удалена → 0 rows → идемпотентность |
| `WorkoutRepository.GetUserScore` | **Да** | Удалённые не входят в SUM |

**Методы `WorkoutExerciseRepository` не меняются** — они работают со своей таблицей. Но usecase `CreateWorkoutExercise`, `PatchWorkoutExercise`, `DeleteWorkoutExercise`, `GetWorkoutExercises` **проверяют `workout.IsDeleted()`** и возвращают `ErrWorkoutNotFound`, если тренировка удалена. Мутации упражнений на удалённой тренировке запрещены.

**Идемпотентность (ADR-007 F-3).** `DeleteWorkout.Execute` перехватывает `ErrWorkoutNotFound` от `MarkDeleted` и возвращает `nil`. Повторный DELETE — не ошибка.

**Proto.** В `Workout` добавлено `bool is_deleted`. Gateway возвращает его клиенту — по аналогии с `is_deleted` в Exercises и Habits.

**Gateway.** `GET /workouts/{id}` возвращает 200 с `is_deleted: true` для удалённой. `GET /workouts` (список) — только активные. `DELETE /workouts/{id}` возвращает 204 (идемпотентно). Аналогично Habits (ADR-023 п.6).

**Purge не делаем.** Удалённые тренировки остаются в БД навсегда. Триггер для purge-скрипта (если появится) — рост БД > 10 ГБ или GDPR-требование «стереть всё через N дней».

### 4. Агрегат `Workout` и его компоненты

`Workout` — **aggregate root**. `WorkoutExercise` — **компонент агрегата** (entity внутри `Workout`), не самостоятельный aggregate root.

Из этого следуют три следствия, уже зафиксированные в других разделах:

- **Hard delete для `workout_exercises`** (п.3.2) — компоненты удаляются физически, без soft delete. `ON DELETE CASCADE` срабатывает при удалении корня.
- **Терминальные статусы `Workout` блокируют мутации компонентов** (п.9) — целостность агрегата важнее, чем удобство.
- **`workout_exercises` не имеет `user_id`, `deleted_at`** — атрибуты корня, не компонента.

**Почему два репозитория при одном агрегате.** Техническое удобство:

- `workout_exercises` — отдельная таблица, отдельные SQL-запросы (`INSERT` / `UPDATE` / `DELETE` / `SELECT`).
- В одном репозитории смешались бы методы работы с `workouts` (6 шт.) и с `workout_exercises` (5 шт.) — неудобно тестировать и ревьюить.
- `WithTx` передаёт **оба** репозитория в callback (п.3.1) — атомарная мутация агрегата возможна.

Это компромисс между DDD-чистотой (один репозиторий на агрегат) и практичностью (два репозитория по таблицам). Осознанный выбор — как в Exercises и Habits.

**Про `recalculateScore`.** Не метод репозитория, а helper в `usecases/common.go`. Координирует чтение из обоих репозиториев и обновление `workouts.workout_score` внутри `WithTx`. В монолите `WorkoutExercisesRepository.UpdateWorkoutScore` писал в `workouts` из репозитория `workout_exercises` — нарушение границ. В новом проекте: `WorkoutRepository.Update(ctx, workout)` с version, вызывается из `recalculateScore`.

### 5. `user_workout_score` не хранится

По ADR-002: `SUM(workout_score) WHERE user_id = ? AND status = 'completed' AND deleted_at IS NULL`, считается на лету.

В Workouts это отдельный gRPC-метод **`GetUserScore() → int64`** (без параметра, `user_id` из metadata). Gateway в `GET /users/me` параллельно зовёт `Profiles.GetMyProfile` и `Workouts.GetUserScore`, склеивает в один ответ (ADR-002, ADR-009).

`UpdateUserWorkoutScore` из монолита **не переносится**.

### 6. Зависимости (sync gRPC)

Habits не зовёт другие сервисы (ADR-023). Workouts — зовёт:

| Кто | Кого | Когда | Зачем |
|---|---|---|---|
| Workouts | Profiles | `CreateWorkout` | `GetProfile(user_id)` → коэффициент |
| Workouts | Exercises | `CreateWorkoutExercise`, `PatchWorkoutExercise` | `GetExercise(id)` → `type` (валидация) и `difficulty` (снапшот при создании) |

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

**`GetUserScore`** — для агрегации в Gateway (`GET /users/me`). `user_id` берётся только из metadata (ставит Gateway из JWT), параметра в запросе нет. Аналог `Profiles.GetMyProfile`.

**Proto для `GetUserScore`:**

- `GetUserScoreRequest` — пустой (как `GetMyProfileRequest`).
- `GetUserScoreResponse` — `int64 score`. `int64` — запас на случай больших агрегатов (за карьеру пользователь может накопить больше 2^31 очков).

**`PatchWorkoutRequest.intensity` — `NullableInt32` wrapper.**

Поле `intensity` в БД nullable (`INTEGER NULL`). Чтобы различать три состояния в PATCH-запросе, используется wrapper (по образцу Profiles, ADR-001):

```protobuf
message NullableInt32 {
  optional int32 value = 1;
}

message PatchWorkoutRequest {
  string        id        = 1;
  optional string status  = 2;
  NullableInt32 intensity = 3;
}
```

Три состояния:

- `intensity = nil` — «не трогать».
- `intensity = NullableInt32{}` (пустой wrapper) — «сбросить в NULL».
- `intensity = NullableInt32{value: 5}` — «установить 5».

**`PatchWorkoutExerciseRequest` — без wrapper'ов.**

В БД `weight`, `sets`, `reps`, `duration`, `completed` либо NOT NULL, либо NULL невалиден по типу упражнения (`ValidateForWorkoutExerciseType`). Третьего состояния нет — `optional int32` / `optional bool` достаточно (nil = «не трогать», не-nil = «установить»).

В domain это отражено в типах патчей:

- `WorkoutPatch.Status *Status` — NOT NULL в БД, два состояния.
- `WorkoutPatch.Intensity nullable.Nullable[Intensity]` — nullable в БД, три состояния.
- `WorkoutExercisePatch.Weight *Weight`, `Sets *Sets`, `Reps *Reps`, `Duration *Duration`, `Completed *bool` — без nullable, третье состояние невалидно.

**Права доступа:** только свои очки. Других вариантов нет — параметра `user_id` в запросе нет, взять чужое невозможно.

Leaderboard (когда появится) будет читать batch-методом `GetUsersScores(ids)` с другим auth (service-to-service, без user metadata) — отдельный ADR.

### 8. HTTP-маршруты через Gateway

| Метод | Путь | Назначение |
|---|---|---|
| GET | `/workouts` | Список тренировок |
| POST | `/workouts` | Создать |
| GET | `/workouts/{workoutId}` | Одна тренировка |
| PATCH | `/workouts/{workoutId}` | Патч (status, intensity) |
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
- `exercise_id` — **immutable** после создания `workout_exercise`. Менять упражнение нельзя — только `DeleteWorkoutExercise` + `CreateWorkoutExercise`. В `PatchWorkoutExerciseRequest` поля `exercise_id` нет (аналог `type` в ADR-008).
- `weight` — обязателен для weight-типа, NULL для duration-типа. **Допустимый диапазон: 1–1000** (кг). `1 ≤ weight ≤ 1000`. Валидация в VO `Weight`.
- `sets` — обязателен для weight-типа, NULL для duration-типа. **Допустимый диапазон: 1–100**. `1 ≤ sets ≤ 100`. Валидация в VO `Sets`.
- `reps` — обязателен для weight-типа, NULL для duration-типа. **Допустимый диапазон: 1–1000**. `1 ≤ reps ≤ 1000`. Валидация в VO `Reps`.
- `duration` — обязателен для duration-типа, NULL для weight-типа. **Допустимый диапазон: 1–86400** (24 часа). `1 ≤ duration ≤ 86400`. Валидация в VO `Duration`.
- `completed` — можно поставить только когда тренировка в `in_progress`. Снять можно всегда (кроме `completed`/`cancelled` статуса тренировки).
- Нельзя перевести тренировку в `completed`, если ни одно упражнение не отмечено completed.
**Полная таблица переходов `Workout.Status`:**

| Из | Можно в | Нельзя |
|---|---|---|
| `planned` | `in_progress`, `cancelled`, `planned` (self) | `completed` |
| `in_progress` | `completed`, `cancelled`, `in_progress` (self) | `planned` |
| `completed` | `completed` (self) | всё остальное |
| `cancelled` | `cancelled` (self) | всё остальное |

**Правила переходов:**

- **Переход `planned → completed` запрещён.** Обязательный маршрут: `planned → in_progress → completed`. Гарантирует, что у завершённой тренировки `started_at` всегда заполнен (ставится при переходе в `in_progress`).
- **Переход `in_progress → planned` запрещён** — нельзя «отменить» начало тренировки.
- **Переход в `cancelled` разрешён** из `planned` и `in_progress`.
- **Переход из `completed` / `cancelled` в любой другой статус запрещён.** Терминальные состояния (метод `Status.IsTerminal() == true`).
- **Self-transition — разрешён как idempotent no-op.** PATCH с тем же статусом возвращает 200, но ничего не меняет (включая `updated_at`). Это семантика REST PATCH: повторный запрос с тем же значением — no-op, не ошибка. Без этого пользователь, дважды кликнувший «Начать тренировку», получил бы `409 Conflict` на второй клик.

Валидация числовых полей (`1 ≤ weight ≤ 1000`, `1 ≤ sets ≤ 100`, `1 ≤ reps ≤ 1000`, `1 ≤ duration ≤ 86400`) — в VO (`Weight`, `Sets`, `Reps`, `Duration`), тестируется в `domain`.

**Инварианты `started_at` / `completed_at` — домен ставит автоматически, пользователь не присылает:**

- `status = in_progress` и `started_at == nil` → домен ставит `started_at = time.Now().UTC()`.
- `status = completed` и `completed_at == nil` → домен ставит `completed_at = time.Now().UTC()`.
- Уже заполненное поле не перезаписывается (идемпотентность при повторном патче).
- Переход `planned → completed` запрещён (см. выше), поэтому к моменту `completed_at` поле `started_at` гарантированно заполнено.

В proto и HTTP-DTO полей `started_at` / `completed_at` в `PatchWorkout` **нет** — они управляются только доменом.

Валидация типа-специфичных полей — `WorkoutExercise.validateForWorkoutExerciseType(type)` (приватный метод; вызывается из `NewWorkoutExercise` и `ApplyPatch`). Это доменное правило, тестируется в `domain`.

### 10. Идентификаторы и порты

По ADR-005:
- gRPC — **50055** (Docker) / **50065** (локально).
- HTTP (health/ready/metrics) — **8085** (Docker) / **8095** (локально).
- PostgreSQL — **5436** на хосте.

### 11. Что НЕ делаем

- **События `workout.completed` в Kafka** — ADR-024, после Notification. Сначала Workouts стабильно работает через sync gRPC.
- **Пагинация `GetWorkouts`** — MVP возвращает весь список. Триггер: 100+ тренировок у одного пользователя и p95 > 500ms. Тогда — `page`/`limit` в proto.
- **Кэш `GetUserScore` в Redis** — по ADR-002, при росте нагрузки. Триггер — `GET /users/me` станет bottleneck.
- **Пессимистичные блокировки (`FOR UPDATE`)** — не используем в Workouts, обоснование в п.3 (в Profiles/Exercises используем — ADR-007 F-2).
- **`UpdateUserWorkoutScore`** — не переносим, `user_workout_score` не хранится (ADR-002).
- **Purge удалённых тренировок** — soft delete оставляет строки навсегда. Триггер для purge-скрипта: рост БД > 10 ГБ или GDPR-требование.
- **`RestoreWorkout`** — восстанавливать удалённые тренировки через API не планируем. `deleted_at` ставится, но обратной операции в proto нет. Если понадобится — отдельный usecase и RPC.
- **CB/Retry на клиентах к Profiles/Exercises — на шаге 4 (usecases).** В skeleton настроены только `GRPC_CLIENT_TLS_MODE` и `WORKOUTS_TIMEOUT` из env (по образцу Gateway). CB/Retry добавятся, когда появятся реальные вызовы — тогда можно осмысленно подобрать пороги (`CBMinRequests`, `CBErrorRate`) и per-target retry-политику (ADR-020 «Когда пересмотреть»).

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
- **9 из 10 usecase — без изменений** (только `PatchWorkout` получает одну строку `publisher.Publish`).
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

- **Sync gRPC в Exercises перед мутацией.** `CreateWorkoutExercise` и `PatchWorkoutExercise` ждут `Exercises.GetExercise(id)` до начала транзакции. Если Exercises лежит — операция падает (CB/Retry помогают, но не устраняют зависимость).
- **Optimistic locking даёт 409 конфликты.** При параллельном патче одного упражнения (два окна) второй клиент получит конфликт. Для тренировки это редкий сценарий, клиент перечитывает.
- **`GetUserScore` — это `SUM` по всем completed-тренировкам пользователя.** На больших историях — медленно. Индекс `(user_id, status)` помогает, но при миллионах тренировок понадобится материализованное представление.
- **Профиль может быть не заполнен** — тренировка создастся, но score минимальный. Не блокируем создание — клиент сам решает.
- **Fallback `/users/me` скрывает деградацию.** При падении Workouts клиент получит `workout_score: 0, workout_score_available: false`. Если клиент не проверяет флаг — покажет «0 очков» как реальное значение.

**Когда пересмотреть:**

- Если `GetUserScore` тормозит → кэш в Redis на 30–60с, или материализованное представление (триггер — p95 > 100мс).
- Если конфликты 409 станут частыми → пересмотреть optimistic locking на пессимистичный (триггер — >5% запросов в 409).
- Если `Exercises.GetExercise` станет узким местом перед мутацией → batch `GetExercisesByIds` в Exercises для `GET /workouts/{id}/exercises` (в Gateway-агрегации) — триггер p95 > 100мс.
- Если появится публичный лидерборд → Leaderboard перестаёт звать `Workouts.GetUserScore` синхронно, начинается кэш/события (ADR-024, Уровень 3).

## Связанные решения

- [ADR-001: Профиль пользователя](001-profile.md) — снапшот коэффициента, lazy-create.
- [ADR-002: user_workout_score](002-user-workout-score.md) — не храним, SUM на лету.
- [ADR-004: Паттерны микросервисов](004-microservices-patterns.md) — Database per Service, sync gRPC.
- [ADR-005: Распределение портов](005-port-allocation.md) — 50055 / 8085 / 5436.
- [ADR-007: Известные проблемы](007-known-issues.md) — F-2 (пессимистичная блокировка в Profiles/Exercises — здесь другой выбор, обосновано).
- [ADR-008: Жизненный цикл упражнения](008-exercise-lifecycle.md) — `difficulty` snapshot, `type` immutable.
- [ADR-009: Карта зависимостей](009-service-dependencies.md) — Workouts → Profiles, Workouts → Exercises.
- [ADR-011: Naming conventions](011-naming-and-code-style.md) — стиль кода.
- [ADR-023: Сервис привычек](023-habits.md) — ориентир по структуре.
- [ADR-024: Kafka и сервис Notification](024-kafka-notification.md) — будущие события `workout.completed` (п.12).

## Статус

**В работе.**

- ✅ Skeleton сервиса (proto, config, bootstrap, миграции, domain stubs) — PR `feat/workouts-skeleton-fixes`.
- ✅ Domain: value objects, сущности (`Workout`, `WorkoutExercise`), state machine (`ApplyPatch`), ошибки, `exercise_load` — ветка `feat/workouts-domain`.
- ✅ Domain: `CalculateScore`, `CalculatePersonalScoreCoefficient`, `saturatedScore` (приватная), `AddLoadWithLimit`, `DivideRoundUp` — ветка `feat/workouts-score-coefficient`.
- ⬜ Domain: unit-тесты (VO, сущности, `ApplyPatch`, state machine) — ветка `feat/workouts-domain-tests`.
- ⬜ Infrastructure (postgres_repo, mappers, integration tests).
- ⬜ Usecases (10 usecase + common + recalculateScore + моки + unit-тесты).
- ⬜ gRPC-сервер + bootstrap + config + main (реализация хендлеров).
- ⬜ Compose + Taskfile + `.env.enc` + CI matrix.
- ⬜ Gateway-хендлеры + маршруты + агрегация `/users/me`.