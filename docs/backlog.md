# Backlog

Открытые задачи проекта, зафиксированные в ADR. Здесь — что делать дальше.
Для каждой задачи указан ADR — источник контекста и обоснования.

## Правила

- Каждая строка ссылается на ADR, где задача зафиксирована.
- При закрытии задачи — обновить статус в ADR. Выполненное **не висит в «В работе»**: сервис переезжает в таблицу «Состояние сервисов» со статусом ✅, секция детализации удаляется.
- Статусы: ✅ — сделано, 🚧 — в работе, ⬜ — не начато.
- Задачи «по триггеру» не начинаем до наступления триггера — они здесь как напоминание.

## 📊 Состояние сервисов

| Сервис | Статус | ADR |
|---|---|---|
| Auth | ✅ | [ADR-001](adr/001-profile.md), [ADR-007](adr/007-known-issues.md) |
| Gateway | ✅ | [ADR-009](adr/009-service-dependencies.md) |
| Profiles | ✅ | [ADR-001](adr/001-profile.md) |
| Exercises | ✅ | [ADR-008](adr/008-exercise-lifecycle.md) |
| Habits | ✅ | [ADR-023](adr/023-habits.md) |
| Workouts | 🚧 | [ADR-025](adr/025-workouts.md) |
| Leaderboard | ⬜ | — |
| Notification | ⬜ | [ADR-024](adr/024-kafka-notification.md) |

## 🚧 В работе

### Workouts (ADR-025)

Сервис тренировок. Полный контекст — [ADR-025](adr/025-workouts.md).
Зовёт Profiles (коэффициент) и Exercises (справочник). Публикует `workout.completed` (ADR-024, в будущем).

**Когда все шаги ✅ — секция удаляется, Workouts переезжает в «Состояние сервисов» со статусом ✅.**

| Шаг | Что | Ветка | Статус | Оценка |
|---|---|---|---|---|
| 1 | ADR-025 (схема БД, контракты, зависимости, score) | `docs/adr-025-workouts` | ✅ | ~2 ч |
| 2a | Domain: VO, сущности, state machine, `exercise_load`, ошибки | `feat/workouts-domain` | ✅ | ~2 ч |
| 2b | Domain: расчёт score и coefficient | `feat/workouts-score-coefficient` | ✅ | ~1 ч |
| 2c | Domain: unit-тесты (VO, сущности, `ApplyPatch`) | `feat/workouts-domain-tests` | ✅ | ~2 ч |
| 3 | Infrastructure (postgres_repo, mapper, интеграционные тесты) | `feat/workouts-infrastructure` | ⬜ | ~4 ч |
| 4 | Usecases (Create, Get, Patch, Delete, Score) + моки + unit-тесты | `feat/workouts-usecases` | ⬜ | ~5 ч |
| 5 | gRPC-сервер + bootstrap + config + main | `feat/workouts-grpc-and-bootstrap` | ⬜ | ~3 ч |
| 6 | Compose + Taskfile + .env.enc + CI matrix | `feat/workouts-compose-and-taskfile` | ⬜ | ~2 ч |
| 7 | Gateway-хендлеры + маршруты + агрегация `/users/me` | `feat/workouts-gateway` | ⬜ | ~3 ч |

Сделано: шаги 1, 2a, 2b, 2c (~7 ч). Осталось: шаги 3–7 (~17 ч).

## ⏭️ Следующее (после Workouts)

### Kafka + Notification (ADR-024)

Фиксируется в [ADR-024](adr/024-kafka-notification.md). Начнётся после мерджа шага 7 Workouts — тогда переедет в «В работе».

### Leaderboard

ADR ещё нет. Появится, когда начнётся (после Kafka + Notification, либо параллельно — по триггеру).

## 🟡 Стоит сделать в ближайшее время

| Что | ADR | Оценка | Комментарий |
|-----|-----|--------|-------------|
| — | — | — | Раздел пуст. Все ближайшие задачи закрыты. |

## 🟢 Низкий / техдолг

| Что | ADR | Оценка | Комментарий |
|-----|-----|--------|-------------|
| Алерт на `/ready` Gateway | [ADR-017](adr/017-alerting.md) | ~1 ч | Нужен отдельный механизм — healthcheck смотрит на `/health` |
| Алерт на возраст S3-дампов | [ADR-014](adr/014-backups-and-dr.md) | ~1 ч | Fail-open для S3 не покрыт |
| govulncheck по расписанию | [ADR-012](adr/012-ci-cd-optimization.md) | ~30 мин | Отдельный workflow |
| Dependabot / Renovate | [ADR-012](adr/012-ci-cd-optimization.md) | ~30 мин | Авто-PR на зависимости |
| Docker logging options | [ADR-018](adr/018-storage-retention.md) | ~15 мин | `daemon.json` + перезапуск Docker |
| Rotation SSH на Ed25519 | [ADR-006](adr/006-secrets-management.md), [ADR-021](adr/021-independent-deployment.md) | ~30 мин | `appleboy/ssh-action` уже обновлён |

## 🟡 После недели наблюдений за CB/Retry

| Что | ADR | Комментарий |
|-----|-----|-------------|
| Алерты `CircuitBreakerOpen`, `HighRetryRate` | [ADR-020](adr/020-circuit-breaker-and-retry.md) | Нужна статистика порогов |
| Скорректировать дефолты CB | [ADR-020](adr/020-circuit-breaker-and-retry.md) | `CBMinRequests`, `CBErrorRate` |

## ⏳ Ждёт триггера (не начинать сейчас)

| Что | ADR | Триггер |
|-----|-----|---------|
| Bulkheads (семафор на downstream) | [ADR-019](adr/019-microservices-evolution.md), [ADR-020](adr/020-circuit-breaker-and-retry.md) | Реальная нагрузка / инцидент «Gateway упал от затора» |
| Fallback / degradation по эндпоинтам | [ADR-019](adr/019-microservices-evolution.md) | Появление фронта, обсуждение UX |
| Auto-rollback (Level 2) | [ADR-012](adr/012-ci-cd-optimization.md) | Дежурный / SLA / болезненный ручной откат |
| Event bus + Outbox + Saga | [ADR-019](adr/019-microservices-evolution.md) | Cross-service запись в две БД / projections (Leaderboard) |
| Distributed tracing | [ADR-019](adr/019-microservices-evolution.md) | 5+ сервисов или fan-out (1 запрос → 3+ downstream) |
| Несколько хостов / k8s / managed БД | [ADR-019](adr/019-microservices-evolution.md) | Реальная нагрузка, сервисов > 10 |
| mTLS + `GRPC_CLIENT_TLS_MODE=verify` | [ADR-016](adr/016-network-security.md) | Выход из одной docker-сети / внешний gRPC-потребитель |
| Статический анализ импортов `pkg/**` | [ADR-021](adr/021-independent-deployment.md) | Если `pkg/**` начнёт меняться часто |
| Blue-green / canary | [ADR-012](adr/012-ci-cd-optimization.md), [ADR-021](adr/021-independent-deployment.md) | Появление SLA |
| RS256 + JWKS вместо HS256 | [ADR-004](adr/004-microservices-patterns.md) | Второй потребитель JWT / требование не раздавать общий секрет / Auth стал bottleneck валидации |

## Что НЕ делаем (явно зафиксировано)

| Что | ADR |
|-----|-----|
| k8s, Kafka, Service mesh | [ADR-019](adr/019-microservices-evolution.md) |
| Микросервис на каждый CRUD | [ADR-019](adr/019-microservices-evolution.md) |
| Event sourcing / CQRS | [ADR-004](adr/004-microservices-patterns.md) |
| Distributed-транзакции между сервисами | [ADR-003](adr/003-transactions.md) |
| Автоматический инкремент semver | [ADR-021](adr/021-independent-deployment.md) |
| Per-service версионирование | [ADR-021](adr/021-independent-deployment.md) |