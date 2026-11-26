# Backlog

Открытые задачи проекта, зафиксированные в ADR. Здесь — что делать дальше.
Для каждой задачи указан ADR — источник контекста и обоснования.

## Правила

- Каждая строка ссылается на ADR, где задача зафиксирована.
- При закрытии задачи — обновить статус в ADR. Выполненное **не висит в «В работе»**: сервис переезжает в таблицу «Состояние сервисов» со статусом ✅, секция детализации удаляется.
- Статусы сервисов: ✅ — готов, 🚧 — в работе, ⬜ — не начат.
- Статусы этапов: ✅ — завершён, ⬜ — ещё не завершён.
- Задачи «по триггеру» не начинаем до наступления триггера — они здесь как напоминание.

## 📊 Состояние сервисов

| Сервис | Статус | ADR |
|---|---|---|
| Auth | ✅ | [ADR-001](adr/001-profile.md), [ADR-007](adr/007-known-issues.md) |
| Gateway | ✅ | [ADR-009](adr/009-service-dependencies.md) |
| Profiles | ✅ | [ADR-001](adr/001-profile.md) |
| Exercises | ✅ | [ADR-008](adr/008-exercise-lifecycle.md) |
| Habits | ✅ | [ADR-023](adr/023-habits.md) |
| Workouts | ✅ | [ADR-025](adr/025-workouts.md) |
| Leaderboard | 🚧 | [ADR-026](adr/026-leaderboard.md) |
| Notification | ⬜ | [ADR-024](adr/024-kafka-notification.md) |

## 🚧 В работе

### Leaderboard (ADR-026)

Workouts считает очки за период. Leaderboard обновляет готовые топы по расписанию и отдаёт их через Gateway. Правила рейтинга и зависимости описаны в [ADR-026](adr/026-leaderboard.md).

| Шаг | Что | Ветка | Статус |
|---|---|---|---|
| 1 | ADR-026, индекс ADR и план в backlog | `docs/adr-026-leaderboard` | ✅ |
| 2 | Запускаемый скелет: main, конфигурация, bootstrap, gRPC, health-проверка, Dockerfile и `.env.example` | `feat/leaderboard-skeleton` | ✅ |
| 3a | Domain: периоды UTC, строка рейтинга, готовый топ | `feat/leaderboard-domain` | ⬜ |
| 3b | Domain: unit-тесты границ дня, недели, месяца и года | `feat/leaderboard-domain` | ⬜ |
| 4a | Use cases: чтение готового топа, обновление одного периода, порт источника очков | `feat/leaderboard-usecases` | ⬜ |
| 4b | Use cases: моки, тесты пустого топа, сбоя обновления и смены периода | `feat/leaderboard-usecases` | ⬜ |
| 5a | Workouts: SQL для суммирования очков за период, сортировки и лимита | `feat/workouts-period-ranking` | ⬜ |
| 5b | Workouts: интеграционные тесты SQL, включая границы периода, фильтрацию, равные очки и лимит | `feat/workouts-period-ranking` | ⬜ |
| 5c | Workouts: внутренний gRPC-метод, мапперы, валидация и проверка доступа | `feat/workouts-period-ranking` | ⬜ |
| 6a | Leaderboard: gRPC-клиент Workouts и потокобезопасное хранилище топов в памяти | `feat/leaderboard-grpc-and-bootstrap` | ⬜ |
| 6b | Leaderboard: воркер, немедленная загрузка, таймауты и остановка по контексту | `feat/leaderboard-grpc-and-bootstrap` | ⬜ |
| 6c | Leaderboard: рабочие gRPC-обработчики, мапперы и подключение клиента Workouts и воркера в bootstrap | `feat/leaderboard-grpc-and-bootstrap` | ⬜ |
| 7a | Локальный стек: Compose, Taskfile, `.env.enc` и проверка запуска вместе с Workouts | `feat/leaderboard-compose-and-ci` | ⬜ |
| 7b | CI/CD: проверка сборки, образ и деплой | `feat/leaderboard-compose-and-ci` | ⬜ |
| 8a | Gateway: gRPC-клиент Leaderboard, конфигурация и bootstrap | `feat/leaderboard-gateway` | ⬜ |
| 8b | Gateway: HTTP-хендлеры, мапперы, три маршрута и проверка `503` | `feat/leaderboard-gateway` | ⬜ |
| 8c | Postman, проверка деплоя, актуализация README и закрытие задачи | `feat/leaderboard-gateway` | ⬜ |

## ⏭️ Следующее

### Kafka + Notification (ADR-024)

Следующий этап после MVP Leaderboard. План внедрения Kafka и сервиса Notification описан в [ADR-024](adr/024-kafka-notification.md).

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
| k8s, Service mesh | [ADR-019](adr/019-microservices-evolution.md) |
| Микросервис на каждый CRUD | [ADR-019](adr/019-microservices-evolution.md) |
| Event sourcing / CQRS | [ADR-004](adr/004-microservices-patterns.md) |
| Distributed-транзакции между сервисами | [ADR-003](adr/003-transactions.md) |
| Автоматический инкремент semver | [ADR-021](adr/021-independent-deployment.md) |
| Per-service версионирование | [ADR-021](adr/021-independent-deployment.md) |
