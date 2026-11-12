# Orange Team Microservices — Context

**Дата фиксации:** 2026-11-12
**Причина:** передача в новый чат перед шагом 5 Workouts (gRPC-сервер + bootstrap)
**Что нового с прошлой фиксации (2026-10-17):** Workouts — domain, usecases и infrastructure полностью закрыты (шаги 1–4e из плана ADR-025). Осталось: gRPC-сервер (5), compose/CI (6), Gateway (7).

---

## 1. Проект

**Репозиторий:** `github.com/BladeRunner322/orange-team-microservices` (монорепо)

**Стек:** Go-микросервисы, чистая архитектура + DDD.

**Сервисы:**
- ✅ Готовы в проде: Auth, Gateway, Profiles, Exercises, Habits
- 🚧 В работе: Workouts — domain + usecases + infrastructure закрыты, осталось gRPC + compose + Gateway
- ⬜ В плане: Leaderboard, Notification

**Инфраструктура:**
- 20 контейнеров: 5 сервисов + 4 PostgreSQL + 2 Redis + Caddy + Prometheus + Alertmanager + node-exporter + Grafana + Loki + Promtail + WARP + 4 migrate-контейнера
- Прод: Selectel VDS, `root@136.234.4.93`, `/root/projects/orange-team-microservices`, Ubuntu 24.04, 8 ГБ / 4 vCPU / 80 ГБ NVMe
- Домен: `sololevelingms.duckdns.org` (DuckDNS)
- HTTPS через Caddy + Let's Encrypt
- Внутренние порты на `127.0.0.1`, снаружи только 22/80/443, ufw + fail2ban
- gRPC между сервисами — `insecure` (осознанно, ADR-016)
- Compose разделён на `infra/compose.app.yml` + `infra/compose.infra.yml` (ADR-021)
- Релиз `v0.1.0` выпущен (semver через git tag)

---

## 2. Состояние по сервисам

| Сервис | Статус | Порт gRPC | HTTP | PostgreSQL |
|---|---|---|---|---|
| Auth | ✅ | 50051 | 8080 | 5432 |
| Gateway | ✅ | — | 8081 | — |
| Profiles | ✅ | 50052 | 8082 | 5433 |
| Exercises | ✅ | 50053 | 8083 | 5434 |
| Habits | ✅ | 50054 | 8084 | 5435 |
| Workouts | 🚧 | 50055 | 8085 | 5436 |
| Leaderboard | ⬜ | 50056 | — | — |
| Notification | ⬜ | — | 8085 (HTTP-only) | 5436 |

**Порты локально = Docker + 10** (ADR-005).

---

## 3. Roadmap (ADR-019) — статус

- **Уровень 1 — Независимый деплой:** ✅ 100% (ADR-021, релиз v0.1.0)
- **Уровень 2 — Failure isolation:** ✅ Circuit Breaker + Retry (ADR-020). ⬜ Bulkheads, fallback — по триггеру.
- **Уровень 3 — Асинхронность (Kafka):** ⬜ ADR-024 запланирован, ждёт завершения Workouts
- **Уровень 4 — Distributed tracing:** ⬜ по триггеру (5+ сервисов или fan-out)
- **Уровень 5 — k8s / несколько хостов:** ⬜ по триггеру

---

## 4. Ключевые архитектурные принципы

- **Database per Service** (ADR-004)
- **Sync gRPC** между сервисами (ADR-004)
- **`user_workout_score` не хранится** — считается из Workouts (ADR-002)
- **Lazy-create профиля** (ADR-001)
- **Exercises:** `type` immutable, `difficulty` snapshot, soft delete (ADR-008)
- **Gateway** — только прокси + auth + агрегация (ADR-009)
- **Naming (ADR-011):** суффикс `Raw`, `ToDomainX`, `Scan(row)`, `ModelToDomain` с id, package-комментарий один раз
- **Circuit Breaker:** бизнес-ошибки не считаются failure
- **Retry:** только read-only методы (whitelist). Мутации не ретраим, даже если идемпотентны.
- **Порядок интерсепторов (клиент):** `CB → Retry → ClientMetrics → Timeout → UserID → RequestID → invoker`
- **Тестирование (ADR-022):** unit (domain, usecases, mappers, middleware, pkg/*) + integration (postgres_repo, redis_repo, pkg/ratelimit, pkg/postgres). Обвязка — не покрываем. Целевое покрытие: домен 95–100%, usecases 85–95%, infra (unit) 70–85%, integration (repo) 85–95%.

---

## 5. ADR — все 25

- ADR-001: Профиль пользователя (lazy-create, снапшот коэффициента)
- ADR-002: user_workout_score (не хранится, SUM на лету)
- ADR-003: Транзакции (границы сервиса, без distributed)
- ADR-004: Паттерны микросервисов (Database per Service, sync gRPC, отказ от event bus)
- ADR-005: Распределение портов (50051–50060 gRPC, 8080–8089 HTTP, +10 локально)
- ADR-006: Управление секретами (SOPS + age)
- ADR-007: Известные проблемы (техдолг, почти весь закрыт)
- ADR-008: Жизненный цикл упражнения (type immutable, difficulty snapshot, soft delete)
- ADR-009: Карта зависимостей сервисов
- ADR-010: Общие инфраструктурные пакеты (`pkg/` — только то, что идентично в 3+ сервисах)
- ADR-011: Naming conventions и Code style
- ADR-012: Оптимизация CI/CD (matrix + docker cache)
- ADR-013: Требования к инфраструктуре
- ADR-014: Бэкапы и DR (Selectel S3)
- ADR-015: HTTPS через Caddy
- ADR-016: Сетевая безопасность (bind 127.0.0.1, ufw, fail2ban)
- ADR-017: Alerting (Alertmanager + Telegram через WARP)
- ADR-018: Retention метрик (15d) и логов (30d)
- ADR-019: Эволюция в сторону микросервисов (5 уровней)
- ADR-020: Circuit Breaker и Retry — реализован
- ADR-021: Независимый деплой — реализован
- ADR-022: Стратегия тестирования — реализован
- ADR-023: Сервис привычек (Habits) — реализован
- ADR-024: Kafka + Notification — в плане, ждёт Workouts
- ADR-025: Сервис тренировок (Workouts) — в работе (шаги 1–4e ✅, 5–7 остались)

---

## 6. Что сделано в сессиях

### До 2026-10-17

Кратко: Level A/B/C закрыты, Уровень 1 полностью, Уровень 2 частично (CB+Retry), техдолг ADR-007 закрыт почти весь, тестирование по ADR-022 внедрено, Habits полностью готов.

### Сессия 2026-10-17 … 2026-11-12 (текущая)

**Workouts (ADR-025) — шаги 1–4e закрыты:**

| Шаг | Что | Ветка |
|---|---|---|
| 1 | ADR-025 (схема БД, контракты, зависимости, score) | `docs/adr-025-workouts` |
| 2a | Domain: VO, сущности, state machine, `exercise_load`, ошибки | `feat/workouts-domain` |
| 2b | Domain: расчёт score и coefficient | `feat/workouts-score-coefficient` |
| 2c | Domain: unit-тесты | `feat/workouts-domain-tests` |
| 3a | Usecases: реализация 10 сценариев | `feat/workouts-usecases` |
| 3b | Usecases: моки + unit-тесты | `feat/workouts-usecases-tests` |
| 4a | Infrastructure: модели, мапперы, рефакторинг `updated_at` | `feat/workouts-infrastructure` |
| 4b | Infrastructure: репозитории | `feat/workouts-infrastructure` |
| 4c | Infrastructure: интеграционные тесты (Testcontainers) | `feat/workouts-infrastructure` |
| 4d | Infrastructure: gRPC-клиенты к Profiles, Exercises | `feat/workouts-infrastructure` |
| 4e | Infrastructure: CB/Retry на gRPC-клиентах | `feat/workouts-infrastructure` |

**Осталось:** шаг 5 (gRPC-сервер + bootstrap + config + main), шаг 6 (compose + Taskfile + `.env.enc` + CI matrix), шаг 7 (Gateway-хендлеры + `/users/me` агрегация). ~8 часов.

---

## 7. Структура проекта (актуальная)

    orange-team-microservices/
    ├── .github/workflows/ci-cd.yml        # CI/CD (matrix по 5 сервисам, tag-triggered semver)
    ├── api/                               # proto + Postman коллекции
    │   ├── auth/, profiles/, exercises/, habits/, workouts/, leaderboard/
    │   └── postman/
    ├── docs/
    │   ├── adr/                           # ADR-001 … ADR-025
    │   ├── backlog.md                     # актуальные задачи
    │   └── context/                       # ← новый каталог для handoff-документов
    ├── internal/gen/                      # сгенерированный proto-код
    ├── pkg/
    │   ├── ctxkeys/                       # user_id, role, request_id в context
    │   ├── grpc/{client,interceptors,server}/
    │   ├── health/, logger/, metrics/, nullable/
    │   ├── postgres/, pgerrors/, pgconfig/, redis/, ratelimit/
    ├── scripts/
    ├── services/
    │   ├── auth/                          # ✅ ГОТОВ
    │   ├── gateway/                       # ✅ ГОТОВ (5 клиентов: Auth, Profiles, Exercises, Habits, Workouts)
    │   ├── profiles/                      # ✅ ГОТОВ
    │   ├── exercises/                     # ✅ ГОТОВ
    │   ├── habits/                        # ✅ ГОТОВ
    │   ├── workouts/                      # 🚧 domain ✅, usecases ✅, infrastructure ✅, gRPC/bootstrap ⬜
    │   └── leaderboard/                   # ⬜ в плане
    ├── .env.enc, .sops.yaml, Taskfile.yml
    └── README.md

**Не в git:** `certs/`, `coverage/`, `logs/`, `secrets/`, `backup-metrics/`, `.env`, `services/*/.env`, `scripts/rclone.conf`.

---

## 8. Инфраструктурные детали (для справки)

**Compose:** `COMPOSE_FILE=infra/compose.app.yml:infra/compose.infra.yml`, `name: orange-team-microservices` в обоих файлах.

**CI/CD:** matrix по изменённым сервисам (`changed-services`), semver через `git tag`, `--no-deps` при деплое, smoke-тесты после деплоя.

**Тестирование:**
- `task test` — unit, только testable-пакеты (домен, usecases, mappers, middleware, pkg/*)
- `task test-integration` — integration с Testcontainers, покрытие по integration-пакетам (repo, pkg/postgres, pkg/ratelimit)
- `task test-cover` — HTML-отчёт

**Переменные `Taskfile.yml`:**
- `UNIT_COVERAGE_EXCLUDE` — паттерны исключений из unit-покрытия
- `INTEGRATION_PACKAGES` — явный список integration-пакетов (включая `./services/workouts/internal/infrastructure/postgres_repo/...`)

---

## 9. Точка фиксации — Workouts infrastructure merge-ready

**Ветка `feat/workouts-infrastructure`** — PR в работе, все 5 подшагов (4a–4e) готовы, включая интеграционные тесты и CB/Retry.

**Что НЕ закоммичено:**
- PR не создан (нужно создать после проверки)

**Незакоммиченные изменения:**
- Нет. Всё, что связано с 4a–4e, закоммичено в ветку.

**Что готово к мерджу:**
- `feat/workouts-domain` ✅
- `feat/workouts-score-coefficient` ✅
- `feat/workouts-domain-tests` ✅
- `feat/workouts-usecases` ✅
- `feat/workouts-usecases-tests` ✅
- `feat/workouts-infrastructure` 🚧 (PR создаётся)

---

## 10. Открытые задачи и триггеры

**Workouts (ADR-025) — в работе:** см. разделы 14–22 ниже.

**После Workouts (ADR-024) — Kafka + Notification:** 10 шагов, ждёт завершения Workouts.

**Открытые триггеры (не начинать сейчас):**
- Bulkheads (семафор на downstream) — триггер: реальная нагрузка
- Fallback / degradation по эндпоинтам — триггер: появление фронта
- Алерты на CB/Retry (`CircuitBreakerOpen`, `HighRetryRate`) — после недели наблюдений
- RS256 + JWKS — триггер: второй потребитель JWT
- Distributed tracing — триггер: 5+ сервисов
- k8s / несколько хостов — триггер: сервисов > 10

**Техдолг (низкий приоритет):**
- Алерт на `/ready` Gateway
- Алерт на возраст S3-дампов
- govulncheck по расписанию
- Dependabot / Renovate
- Docker logging options
- Rotation SSH на Ed25519

---

## 11. Правила работы — ссылка на `Rules.md`

Формат: **Ветка → Правка → Было/Стало → Коммит → PR → Следующий пункт**.

- «Было/Стало» в 4 обратных кавычках, язык внешнего блока по содержимому
- Commits: conventional, английский, одна строка
- PR: 5 блоков (Что / Контекст / Что внутри / Что не в этом PR / Связанные решения). Название PR — отдельно от тела.
- Ссылки на ADR — полные: `[ADR-016: Сетевая безопасность](docs/adr/016-network-security.md)`
- Язык общения — русский, коммиты/ветки/PR title — английский, содержимое файлов — русский
- Не коммитить самому — только команды
- Git pull: `git pull` (без origin). Push: `git push -u origin <branch>` (первый раз), потом `git push`.
- Не давать блоки в ```text, если содержимое — markdown или код
- Не забывать про экранирование 4 обратными кавычками

**Проверки перед PR:**
- `go build ./services/workouts/...`
- `go test ./services/workouts/... -short`
- `task test-integration` (если затронуты repo)

---

## 12. Что НЕ трогать без причины

- `main` — только через PR
- Не выносить `bootstrap.go` / `main.go` в `pkg/` (ADR-010)
- Не делать cross-service транзакции, saga, event bus без нового ADR
- Не менять `127.0.0.1:` bind на внутренних портах без ADR-016
- Не закрывать 22 порт без bastion / VPN
- Не дублировать `TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` в GitHub Secrets
- Не присылать расшифрованные `.env`, `rclone.conf`, ключи
- Не менять диапазоны VO-валидации без причины (заданы в domain VO)

---

## 13. Gotchas (актуальные)

- **Docker bypass ufw.** Bind на `127.0.0.1` — обязателен.
- **Compose имя проекта.** `name: orange-team-microservices` в обоих файлах.
- **COMPOSE_FILE в SSH-скрипте.** `appleboy/ssh-action` не пробрасывает env — выставлять явно.
- **Promtail structured metadata** — `request_id` и `user_id` в structured metadata, не labels.
- **Alertmanager ${VAR}.** Не работает в конфиге — только file-based secrets.
- **`WithMax` в retry** — количество попыток, не retry. `WithMax(3)` = 3 попытки (1 initial + 2 retry).
- **`saturatedScore` в домене Workouts — приватная**, `CalculateScore` — публичная обёртка. Usecase вызывает только `CalculateScore`.
- **`WithTx` в Workouts принимает TWO репозитория** (`WorkoutRepository`, `WorkoutExerciseRepository`) — потому что `recalculateScore` нужен доступ к обеим таблицам в одной транзакции.
- **`updated_at` в Workouts — nullable** (как в Auth/Profiles/Exercises/Habits), `*time.Time` в модели и домене. До первого UPDATE — nil.
- **`ErrConflict`** — sentinel для optimistic locking (version mismatch). Не `ErrWorkoutNotFound`.
- **`GetByID` возвращает и удалённые** (с `IsDeleted=true`), `GetByUserID` — только активные.
- **`workout_exercises` — hard delete, `workouts` — soft delete** (ADR-025 п.3.2).
- **`difficulty` и `exercise_load` — снапшоты** (не пересчитываются при изменении справочника).
- **`personal_score_coefficient` — снапшот** на момент создания тренировки.
- **`GetUserScore` — SUM на лету**, не хранится (ADR-002).

---

## 14. Сервис Workouts — общее

**Порт:** gRPC 50055 / HTTP 8085 (локально 50065 / 8095).
**PostgreSQL:** 5436.
**Зависимости:** Profiles (`GetProfile` → коэффициент), Exercises (`GetExercise` → type + difficulty).
**Публикует:** `workout.completed` (Kafka, ADR-024 — в будущем).

**Схема БД:** `workouts.workouts` + `workouts.workout_exercises` (FK `ON DELETE CASCADE`).

**Ключевые правила из ADR-025:**
- `status`: `planned → in_progress → completed`, `cancelled` из `planned`/`in_progress`. Терминальные (`completed`/`cancelled`) блокируют мутации. Self-transition — idempotent no-op.
- `started_at`/`completed_at` ставятся доменом автоматически при переходах. Пользователь их не присылает.
- `intensity` (1–10 или NULL) — задаёт пользователь, патчится через `nullable.Nullable[Intensity]`.
- `workout_score` — пересчитывается при мутации `workout_exercises` в `planned`/`in_progress`. После `completed` — фиксирован.
- `personal_score_coefficient` — снапшот при `CreateWorkout`.
- `difficulty` в `workout_exercises` — снапшот при `CreateWorkoutExercise` (не пересчитывается при изменении справочника).
- `exercise_load` — снапшот, вычисляется доменом. weight-тип: `weight × reps × sets`, duration-тип: `duration`.
- `exercise_id` — immutable. Поменять упражнение нельзя — только Delete + Create.
- `workouts` — soft delete (`deleted_at`), `workout_exercises` — hard delete (компонент агрегата).
- Optimistic locking через `version` (в обеих таблицах), `UPDATE ... WHERE version = $N`. 0 rows → `ErrConflict`.
- `GetUserScore` — `SUM(workout_score) WHERE user_id AND status='completed' AND deleted_at IS NULL`. Не хранится (ADR-002).
- `WithTx` передаёт **два** репозитория (`WorkoutRepository` + `WorkoutExerciseRepository`) в callback — атомарная мутация агрегата.
- `recalculateScore` — helper в usecases, вызывается внутри `WithTx` после мутации упражнения.
- `GetByID` возвращает и удалённые (с `IsDeleted=true`). `GetByUserID` — только активные.
- `MarkDeleted` — `WHERE id AND user_id AND deleted_at IS NULL`. 0 rows → `ErrWorkoutNotFound`. Usecase перехватывает → nil (идемпотентность).

---

## 15. Структура `services/workouts/`

    services/workouts/
    ├── cmd/main.go                                    # ✅ точка входа (не трогали после skeleton)
    ├── config/config.go                               # ✅ + CB/Retry поля (шаг 4e)
    ├── .env.example                                   # ✅ + WORKOUTS_CB_*, WORKOUTS_RETRY_* (шаг 4e)
    ├── migrations/
    │   ├── 000001_create_workouts_table.up.sql        # ✅ updated_at nullable
    │   ├── 000001_create_workouts_table.down.sql
    │   ├── 000002_create_workout_exercises_table.up.sql   # ✅ updated_at nullable
    │   └── 000002_create_workout_exercises_table.down.sql
    ├── internal/
    │   ├── domain/
    │   │   ├── value_objects.go                       # ✅ VO (Weight, Sets, Reps, Duration, Difficulty, Intensity, Status, ExerciseType)
    │   │   ├── value_objects_test.go                  # ✅
    │   │   ├── errors.go                              # ✅ + ErrConflict
    │   │   ├── workout.go                             # ✅ Workout + ApplyPatch (state machine)
    │   │   ├── workout_test.go                        # ✅
    │   │   ├── workout_exercise.go                    # ✅ WorkoutExercise + ApplyPatch + validateForWorkoutExerciseType + recalculateExerciseLoad
    │   │   ├── workout_exercise_test.go               # ✅
    │   │   ├── workout_patch.go                       # ✅ WorkoutPatch
    │   │   ├── workout_patch_test.go                  # ✅
    │   │   ├── workout_exercise_patch.go              # ✅ WorkoutExercisePatch
    │   │   ├── workout_exercise_patch_test.go         # ✅
    │   │   ├── coefficient.go                         # ✅ CalculatePersonalScoreCoefficient
    │   │   ├── coefficient_test.go                    # ✅
    │   │   ├── score.go                               # ✅ CalculateScore (public), saturatedScore (private), AddLoadWithLimit, DivideRoundUp
    │   │   ├── score_test.go                          # ✅
    │   │   └── test_helpers_test.go                   # ✅ хелперы для nullable.Intensity
    │   ├── application/
    │   │   ├── ports/
    │   │   │   ├── repository.go                      # ✅ WorkoutRepository, WorkoutExerciseRepository interfaces
    │   │   │   └── clients.go                         # ✅ ProfilesClient, ExercisesClient + UserProfile, ExerciseInfo
    │   │   └── usecases/
    │   │       ├── common.go                          # ✅ userIDFromContext, recalculateScore
    │   │       ├── create_workout.go                  # ✅ + test
    │   │       ├── get_workout.go                     # ✅ + test
    │   │       ├── get_workouts.go                    # ✅ + test
    │   │       ├── patch_workout.go                   # ✅ + test
    │   │       ├── delete_workout.go                  # ✅ + test
    │   │       ├── create_workout_exercise.go         # ✅ + test (принимает *domain.Weight, не *int)
    │   │       ├── get_workout_exercises.go           # ✅ + test
    │   │       ├── patch_workout_exercise.go          # ✅ + test
    │   │       ├── delete_workout_exercise.go         # ✅ + test
    │   │       ├── get_user_score.go                  # ✅ + test
    │   │       ├── mock_test.go                       # ✅ моки для репозиториев и клиентов
    │   │       └── test_helpers_test.go               # ✅ хелперы для VO и Workout
    │   ├── infrastructure/
    │   │   ├── postgres_repo/
    │   │   │   ├── workout_model.go                   # ✅ UpdatedAt *time.Time
    │   │   │   ├── workout_exercise_model.go          # ✅ UpdatedAt *time.Time
    │   │   │   ├── workout_mapper.go                  # ✅
    │   │   │   ├── workout_exercise_mapper.go         # ✅
    │   │   │   ├── workout_repository.go              # ✅ WithTx, Create, GetByID, GetByUserID, Update, MarkDeleted, GetUserScore
    │   │   │   ├── workout_exercise_repository.go     # ✅ Create, GetByID, ListByWorkout, Update, Delete
    │   │   │   ├── main_test.go                       # ✅ TestMain + Testcontainers
    │   │   │   ├── workout_repository_integration_test.go           # ✅
    │   │   │   └── workout_exercise_repository_integration_test.go  # ✅
    │   │   └── clients/
    │   │       ├── profiles_grpc.go                   # ✅ NewProfilesClient, GetProfile (маппинг proto → ports.UserProfile)
    │   │       └── exercises_grpc.go                  # ✅ NewExercisesClient, GetExercise (маппинг proto → ports.ExerciseInfo)
    │   ├── bootstrap/bootstrap.go                     # 🚧 TODO шаг 5 (wiring уже есть, но workoutsgrpc — заглушки)
    │   └── interfaces/
    │       └── workoutsgrpc/
    │           ├── server.go                          # 🚧 TODO шаг 5 (10 RPC — заглушки return nil, nil)
    │           └── mapper.go                          # 🚧 TODO шаг 5 (ToDomainWorkoutPatch, ToDomainWorkoutExercisePatch, ToProtoWorkout, ToProtoWorkoutExercise, ToDomainID)
    └── ...

---

## 16. Домен — ключевые сигнатуры

**VO:**
- `NewWeight(int) (Weight, error)` — 1–1000
- `NewSets(int) (Sets, error)` — 1–100
- `NewReps(int) (Reps, error)` — 1–1000
- `NewDuration(int) (Duration, error)` — 1–86400
- `NewDifficulty(int) (Difficulty, error)` — 1–10
- `NewIntensity(int) (Intensity, error)` — 1–10
- `NewStatus(string) (Status, error)` — planned/in_progress/completed/cancelled
- `NewExerciseType(string) (ExerciseType, error)` — weight/duration

**Workout:**
- `NewWorkout(userID uuid.UUID, coefficient int) Workout` — status=planned, version=0, createdAt=now, updatedAt=nil
- `RestoreWorkout(id, version, userID, status, startedAt, completedAt, workoutScore, intensity, coefficient, isDeleted, createdAt, updatedAt *time.Time) Workout`
- `(w *Workout) ApplyPatch(patch WorkoutPatch) error` — state machine + инварианты
- `(w Workout) CanModifyWorkoutExercise() bool` — `!status.IsTerminal()`
- `(w *Workout) SetWorkoutScore(score int)` — обновляет score + updatedAt

**WorkoutExercise:**
- `NewWorkoutExercise(workoutID, exerciseID uuid.UUID, difficulty Difficulty, weight *Weight, sets *Sets, reps *Reps, duration *Duration, completed bool, exerciseType ExerciseType) (WorkoutExercise, error)` — валидирует + recalc load
- `RestoreWorkoutExercise(id, version, workoutID, exerciseID, difficulty, weight, sets, reps, duration, completed, exerciseLoad, createdAt, updatedAt *time.Time) WorkoutExercise`
- `(e *WorkoutExercise) ApplyPatch(patch WorkoutExercisePatch, t ExerciseType) error` — валидация до мутации, recalc load, updatedAt

**Патчи:**
- `WorkoutPatch{Status *Status; Intensity nullable.Nullable[Intensity]}` — 2 + 3 состояния
- `WorkoutExercisePatch{Weight *Weight; Sets *Sets; Reps *Reps; Duration *Duration; Completed *bool}` — 2 состояния (exercise_id immutable)

**Coefficient:**
- `CalculatePersonalScoreCoefficient(p UserProfileData) int` — BMI-based, male +0.5, clamp [1,10], round-half-up
- `UserProfileData{Sex *string; WeightGrams *int; HeightCM *int}`

**Score:**
- `CalculateScore(effectiveLoad, coefficient int) int` — публичная, тонкая обёртка
- `saturatedScore(effectiveLoad, coefficient int) int` — приватная, `int64` внутри
- `AddLoadWithLimit(current, load, difficulty, limit int) int` — cap по limit
- `DivideRoundUp(value, divisor int) int`
- Константы: `MaximumWorkoutScore = 10_000`, `ScoreSaturationScale = 100_000`, `MaximumScaledLoad = 100_000_000`

---

## 17. Ошибки домена (sentinel)

**Lookup:**
- `ErrWorkoutNotFound`
- `ErrWorkoutExerciseNotFound`

**Бизнес-правила:**
- `ErrInvalidStatusTransition`
- `ErrCannotCompleteEmptyWorkout`
- `ErrCannotModifyWorkoutExercise`
- `ErrCannotMarkExerciseCompleted`
- `ErrInvalidWorkoutExerciseForType`

**VO:**
- `ErrInvalidStatus`, `ErrInvalidWeight`, `ErrInvalidSets`, `ErrInvalidReps`, `ErrInvalidDuration`, `ErrInvalidDifficulty`, `ErrInvalidIntensity`, `ErrInvalidExerciseType`

**Auth:**
- `ErrUnauthenticated`

**Conflict (optimistic locking):**
- `ErrConflict`

---

## 18. Usecases — 10 сценариев + common

**`common.go`:**
- `userIDFromContext(ctx, log) (uuid.UUID, error)` — извлекает из `ctxkeys.UserIDFromContext`, парсит в UUID, `ErrUnauthenticated` если нет
- `recalculateScore(ctx, workoutRepo, exerciseRepo, userID, workoutID) error` — читает workout + exercises, считает `effectiveLoad` (только completed, с `AddLoadWithLimit`), `CalculateScore`, `SetWorkoutScore`, `Update`. **Вызывается только внутри `WithTx`.**

**Usecases (все с проверкой владельца и `IsDeleted`):**
- `CreateWorkout` — `GetProfile` (Profiles) → coefficient → `NewWorkout` → `Create`
- `GetWorkout(id)` — `GetByID` (включая удалённые)
- `GetWorkouts()` — `GetByUserID` (только активные)
- `PatchWorkout(id, patch)` — `WithTx` → `GetByID` → `ApplyPatch` → (если `completed`) проверка «есть хотя бы одно completed» → `Update`
- `DeleteWorkout(id)` — `MarkDeleted`, идемпотентен
- `CreateWorkoutExercise(workoutID, exerciseID, weight *domain.Weight, sets *domain.Sets, reps *domain.Reps, duration *domain.Duration, completed bool)` — проверка `CanModifyWorkoutExercise`, `GetExercise` (Exercises), `NewWorkoutExercise` → `WithTx` → `Create` + `recalculateScore`
- `GetWorkoutExercises(workoutID)` — `ListByWorkout`
- `PatchWorkoutExercise(workoutID, exerciseID, patch domain.WorkoutExercisePatch)` — `GetByID` → `ApplyPatch` → `WithTx` → `Update` + `recalculateScore`
- `DeleteWorkoutExercise(workoutID, exerciseID)` — `WithTx` → `Delete` + `recalculateScore`, идемпотентен (0 rows → nil)
- `GetUserScore()` — `SUM` по completed

**Сигнатура `CreateWorkoutExercise` — принимает VO (`*domain.Weight`), не `*int`.** Это осознанно: консистентно с Profiles (`PatchMyProfile` принимает `domain.ProfilePatch` с VO). Маппинг `*int32 → *domain.Weight` будет в `workoutsgrpc/mapper.go` на шаге 5.

---

## 19. Infrastructure — детали

### `WithTx` (WorkoutRepository)

```go
func (r *WorkoutRepository) WithTx(
    ctx context.Context,
    fn func(ports.WorkoutRepository, ports.WorkoutExerciseRepository) error,
) error {
    if r.tx != nil {
        exRepo := &WorkoutExerciseRepository{tx: r.tx, opTimeout: r.opTimeout}
        return fn(r, exRepo)
    }
    return r.pool.WithTx(ctx, func(tx postgres.Tx) error {
        txWorkoutRepo := &WorkoutRepository{tx: tx, opTimeout: r.opTimeout}
        txExerciseRepo := &WorkoutExerciseRepository{tx: tx, opTimeout: r.opTimeout}
        return fn(txWorkoutRepo, txExerciseRepo)
    })
}
```

`WorkoutExerciseRepository` **не имеет** своего `WithTx` — транзакция приходит извне через `WithTx` у `WorkoutRepository`.

### Клиенты

**`profiles_grpc.go`:**
- `NewProfilesClient(ctx, cfg grpcclient.Config) (*ProfilesClient, error)` — `grpcclient.New` + `profiles.NewProfilesServiceClient(conn)`
- `GetProfile(ctx, userID uuid.UUID) (ports.UserProfile, error)` — вызов + маппинг `proto.UserProfile → ports.UserProfile`, включая:
  - `uuid.Parse(resp.UserId)`
  - `optionalString(resp.Sex)` — `"" → nil`
  - `optionalWeightGrams(resp.WeightKg)` — `0 → nil`, кг → граммы через `math.Round(kg*1000)`
  - `optionalInt(resp.HeightCm)` — `0 → nil`
- `IsHealthy(ctx)` — `conn.GetState()` (Ready/Idle/Connecting = ok)
- `Close()`

**`exercises_grpc.go`:**
- `NewExercisesClient(ctx, cfg) (*ExercisesClient, error)`
- `GetExercise(ctx, id uuid.UUID) (ports.ExerciseInfo, error)` — вызов + маппинг: `uuid.Parse(resp.Id)`, `resp.Type`, `int(resp.Difficulty)`
- `IsHealthy(ctx)`, `Close()`

**Нюанс:** в `profiles.proto` поля `sex`, `weight_kg`, `height_cm` — не optional. Нулевое значение интерпретируется как «не задано» → `nil` в `ports.UserProfile`. Это **известное ограничение** существующего proto (не наша задача сейчас). Альтернатива — переделать proto на `optional`, но это отдельная задача.

### CB/Retry (шаг 4e)

**`config.go`:**
- `WORKOUTS_CB_ENABLED=true`
- `WORKOUTS_CB_MAX_REQUESTS=1`
- `WORKOUTS_CB_INTERVAL=60s`
- `WORKOUTS_CB_TIMEOUT=30s`
- `WORKOUTS_CB_MIN_REQUESTS=10`
- `WORKOUTS_CB_ERROR_RATE=0.5`
- `WORKOUTS_RETRY_ENABLED=true`
- `WORKOUTS_RETRY_MAX_ATTEMPTS=3`
- `WORKOUTS_RETRY_BASE_DELAY=100ms`
- `WORKOUTS_RETRY_MAX_DELAY=1s`
- `WORKOUTS_RETRY_IDEMPOTENT_METHODS=/profiles.ProfilesService/GetProfile,/exercises.ExercisesService/GetExercise`

**`bootstrap.buildGRPCConfig`** — заполняет все поля `grpcclient.Config` (CB + Retry) из `config.Config`.

**.env.example** — блок CB/Retry идентичен Gateway (те же комментарии, та же структура), отличается только префикс `WORKOUTS_` и whitelist.

---

## 20. Что осталось — шаги 5, 6, 7

### Шаг 5 — gRPC-сервер + bootstrap + config + main

**Ветка:** `feat/workouts-grpc-and-bootstrap`
**Оценка:** ~3 часа

**Файлы:**
1. **`services/workouts/internal/interfaces/workoutsgrpc/mapper.go`** — маппинг proto ↔ domain:
   - `statusFromProto(*string) (*domain.Status, error)` — nil → nil, иначе `domain.NewStatus`
   - `intensityFromProto(*workouts.NullableInt32) (nullable.Nullable[domain.Intensity], error)` — 3 состояния
   - `weightFromProto(*int32) (*domain.Weight, error)` — nil → nil, `domain.NewWeight(int(*raw))`
   - `setsFromProto(*int32) (*domain.Sets, error)`
   - `repsFromProto(*int32) (*domain.Reps, error)`
   - `durationFromProto(*int32) (*domain.Duration, error)`
   - `completedFromProto(*bool) *bool`
   - `ToDomainWorkoutPatch(*workouts.PatchWorkoutRequest) (domain.WorkoutPatch, error)`
   - `ToDomainWorkoutExercisePatch(*workouts.PatchWorkoutExerciseRequest) (domain.WorkoutExercisePatch, error)`
   - `ToDomainID(string) (uuid.UUID, error)` — `uuid.Parse`
   - `ToProtoWorkout(domain.Workout) *workouts.Workout` — маппинг полей, `timestamppb.New` для `*time.Time`
   - `ToProtoWorkouts([]domain.Workout) []*workouts.Workout`
   - `ToProtoWorkoutExercise(domain.WorkoutExercise) *workouts.WorkoutExercise`
   - `ToProtoWorkoutExercises([]domain.WorkoutExercise) []*workouts.WorkoutExercise`

2. **`services/workouts/internal/interfaces/workoutsgrpc/server.go`** — 10 RPC-методов с маппингом доменных ошибок в gRPC-коды:
   - `ErrWorkoutNotFound` / `ErrWorkoutExerciseNotFound` → `codes.NotFound`
   - `ErrUnauthenticated` → `codes.Unauthenticated`
   - `ErrInvalidStatusTransition` / `ErrCannotCompleteEmptyWorkout` / `ErrCannotModifyWorkoutExercise` / `ErrCannotMarkExerciseCompleted` / `ErrInvalidWorkoutExerciseForType` → `codes.FailedPrecondition` или `codes.InvalidArgument`
   - `ErrConflict` → `codes.Aborted`
   - VO-ошибки (`ErrInvalidStatus`, `ErrInvalidWeight`, ...) → `codes.InvalidArgument`
   - `default` → `codes.Internal`
   - `GetUserScore` возвращает `int64`

3. **`services/workouts/internal/bootstrap/bootstrap.go`** — уже собирает все зависимости. Проверить:
   - `workoutsgrpc.NewServer(...)` вызывается со всеми 10 usecases (уже есть)
   - `workouts.RegisterWorkoutsServiceServer(s, server)` регистрирует (уже есть)
   - `/ready` использует `profilesClient.IsHealthy` и `exercisesClient.IsHealthy` (уже есть)
   - `buildGRPCConfig` заполняет CB/Retry (уже есть, шаг 4e)

4. **`services/workouts/config/config.go`** — проверить, что все обязательные поля есть (уже есть).

5. **`services/workouts/cmd/main.go`** — проверить graceful shutdown (уже есть по образцу Exercises).

**Проверки перед PR:**
- `go build ./services/workouts/...`
- `go test ./services/workouts/... -short`
- `task test-integration` (workouts-тесты уже зелёные)
- `task postman -- workouts` (когда Gateway-маршруты будут, шаг 7; на шаге 5 — опционально)

### Шаг 6 — Compose + Taskfile + .env.enc + CI matrix

**Ветка:** `feat/workouts-compose-and-taskfile`
**Оценка:** ~2 часа

**Что нужно:**
1. **`infra/compose.app.yml`** — добавить `postgres-workouts`, `migrate-workouts`, `workouts` (по образцу habits). Порты 5436/50055/8085.
2. **`infra/compose.infra.yml`** — ничего (инфра не меняется).
3. **`infra/prometheus.yml`** — добавить job `workouts` (`targets: ['workouts:8085']`).
4. **`infra/alerts.yml`** — добавить `job="workouts"` в regex `ServiceDown` и `HighGRPCErrorRate`.
5. **`taskfiles/workouts.yml`** — новый файл (по образцу `taskfiles/habits.yml`): `run`, `logs`, `psql`, `migrate-create`, `restart`, `build`.
6. **`Taskfile.yml`** — `includes: workouts: ./taskfiles/workouts.yml`.
7. **`.env.example`** (корневой) — добавить `WORKOUTS_POSTGRES_PORT=5436`, `WORKOUTS_GRPC_PORT=50055`, `WORKOUTS_HTTP_PORT=8085`.
8. **`scripts/backup-db.sh`** — добавить `workouts` в массив `SERVICES`.
9. **`.github/workflows/ci-cd.yml`** — добавить `workouts` в `changed-services` (paths), в matrix, в `build-and-push`, в deploy (migrate + up -d --no-deps), в smoke-test readiness (если нужно).
10. **`.sops.yaml`** — добавить `path_regex: services/workouts/\.env$` (уже может быть).
11. **`.env.enc`** (root) и **`services/workouts/.env.enc`** — зашифровать через `task sops:encrypt`.
12. **`services/habits/.env.example`** — для справки, как образец.

**После:** `task docker-up`, `curl http://localhost:8085/health`, `task postman -- workouts` (когда Gateway будет, шаг 7).

### Шаг 7 — Gateway-хендлеры + маршруты + `/users/me`

**Ветка:** `feat/workouts-gateway`
**Оценка:** ~3 часа

**Что нужно:**

1. **`services/gateway/internal/application/ports/workouts_client.go`** — новый интерфейс `WorkoutsClientInterface`:
   - `CreateWorkout(ctx)`
   - `GetWorkout(ctx, id)`
   - `GetWorkouts(ctx)`
   - `PatchWorkout(ctx, req)`
   - `DeleteWorkout(ctx, id)`
   - `CreateWorkoutExercise(ctx, req)`
   - `GetWorkoutExercises(ctx, workoutID)`
   - `PatchWorkoutExercise(ctx, req)`
   - `DeleteWorkoutExercise(ctx, workoutID, exerciseID)`
   - `GetUserScore(ctx)`
   - `IsHealthy(ctx) error`
   - `Close()`

2. **`services/gateway/internal/infrastructure/clients/workouts_grpc.go`** — реализация (по образцу `habits_grpc.go`).

3. **`services/gateway/internal/interfaces/http/handlers/workouts/`** — 9 хендлеров:
   - `handler.go` — хендлеры
   - `dto.go` — DTO (CreateWorkoutRequest пустой, PatchWorkoutRequest с `Nullable`, CreateWorkoutExerciseRequest с `ExerciseID uuid.UUID`)
   - `mapper.go` — proto ↔ HTTP-DTO
   - `mock_test.go`, `handler_test.go`, `mapper_test.go` — unit-тесты

4. **`services/gateway/internal/bootstrap/bootstrap.go`** — заменить `proxy.Handler` для `/workouts/*` на реальные хендлеры. Добавить `workoutsClient` в `App`.

5. **`services/gateway/internal/interfaces/http/handlers/profiles/handler.go`** — `GetUserHandler` теперь агрегирует `Profiles.GetMyProfile` + `Workouts.GetUserScore`:
   - `errgroup.WithContext` для параллельных вызовов
   - Profiles 5xx → весь `/users/me` 503 (профиль обязателен)
   - Workouts 5xx/CB/timeout → `workout_score: 0, workout_score_available: false`, HTTP 200 (профиль отдаётся)
   - Расширить `UserProfileResponse`: `WorkoutScore int64`, `WorkoutScoreAvailable bool`

6. **`.env.example`** (gateway) — добавить `WORKOUTS_GRPC_ADDR=workouts-service:50055`.
7. **`services/gateway/.env.enc`** — перешифровать с новой переменной.
8. **`Taskfile.yml`** — `INTEGRATION_PACKAGES` уже включает `services/workouts/...`, проверить.
9. **CI matrix** — Gateway пересобирается при изменении `api/workouts/**` и `internal/gen/api/workouts/**` (уже есть в `changed-services`).
10. **Postman collection** — `api/postman/workouts_collection.json` (по образцу `habits_collection.json`). Порядок: Create, Get list, Get by ID, Patch (status → in_progress → completed), Delete, Create exercise, Get exercises, Patch exercise (completed → recalc score), Delete exercise, Get user score.

---

## 21. Ключевые решения в этом PR (для контекста нового чата)

1. **`updated_at` — nullable.** В миграциях `workouts` и `workout_exercises` — `TIMESTAMPTZ` без `NOT NULL`. В домене и моделях — `*time.Time`. Обоснование: консистентность с Auth/Profiles/Exercises/Habits. До первого UPDATE — nil.

2. **`ErrConflict` для optimistic locking.** Не `ErrWorkoutNotFound`, потому что usecase перед `Update` всегда делает `GetByID` — «не найдено» в `Update` означает «конкурентное изменение/удаление», а не «записи не было».

3. **`saturatedScore` — приватная.** Только `CalculateScore` (публичная) вызывается из usecase. `AddLoadWithLimit`, `DivideRoundUp` — публичные (нужны в `recalculateScore`).

4. **`WithTx` передаёт ДВА репозитория.** `recalculateScore` ходит и в `workouts`, и в `workout_exercises` — обе операции должны быть в одной транзакции. Альтернатива (один репозиторий на агрегат) отвергнута в ADR-025 п.4.

5. **`CreateWorkoutExercise` принимает VO (`*domain.Weight`), не `*int`.** Консистентно с Profiles (`PatchMyProfile` → `domain.ProfilePatch`). Маппинг `*int32 → *domain.Weight` — в `workoutsgrpc/mapper.go` (шаг 5).

6. **Клиенты: `proto → ports`, не `proto → JSON`.** В Gateway клиенты возвращают proto (HTTP-хендлеры сами маппят в JSON). В Workouts — возвращают `ports.UserProfile` / `ports.ExerciseInfo`, потому что usecases ждут именно их. Маппинг внутри клиента.

7. **Нулевые proto-значения → nil в ports.** `sex=""` → `nil`, `weight_kg=0` → `nil`, `height_cm=0` → `nil`. Причина: в `profiles.proto` поля не optional, «не задано» приходит как 0. Без этого `CalculatePersonalScoreCoefficient` поделит на ноль.

8. **CB/Retry с дефолтами Gateway (ADR-020).** Whitelist read-only: `/profiles.ProfilesService/GetProfile`, `/exercises.ExercisesService/GetExercise`. Тонкая настройка порогов — после недели наблюдений.

9. **Compose: `postgres-workouts` (5436), `migrate-workouts`, `workouts` (50055, 8085).** Уже подготовлены в скелете (шаг 1), но в `compose.app.yml` ещё нет — добавить в шаге 6.

10. **`outbox`/`events` — не делаем сейчас.** ADR-024 (Kafka + Notification) — после Workouts. `workout.completed` будет публиковаться в `PatchWorkout` (одна строка).

---

## 22. Точка фиксации

**Текущий статус:**
- Ветка `feat/workouts-infrastructure` — все 5 подшагов (4a–4e) закоммичены и запушены. PR создаётся.
- Остальные ветки (`feat/workouts-domain`, `-score-coefficient`, `-domain-tests`, `-usecases`, `-usecases-tests`) — смерджены в `main` (по идее, или ждут).

**Что сделать первым в новом чате:**
1. Прочитать `Rules.md`, этот файл (`Context.md`), `docs/adr/025-workouts.md`, `docs/backlog.md`.
2. Прочитать текущий `services/workouts/internal/interfaces/workoutsgrpc/server.go` (заглушки) и `mapper.go` (TODO).
3. Создать ветку `feat/workouts-grpc-and-bootstrap` от `main`.
4. Начать с `mapper.go` — по порядку: `ToDomainXxx` (proto → domain) сначала, потом `ToProtoXxx`.

**Формат работы:** по `Rules.md` — Ветка → Правка → Было/Стало → Коммит → PR → «Подтверждаешь?».
