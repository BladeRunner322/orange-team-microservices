# Backlog

Открытые задачи проекта, зафиксированные в ADR. Здесь — что делать дальше.
Для каждой задачи указан ADR — источник контекста и обоснования.

## Правила

- Каждая строка ссылается на ADR, где задача зафиксирована.
- При закрытии задачи — обновить статус в ADR, потом удалить строку отсюда.
- Задачи «по триггеру» не начинаем до наступления триггера — они здесь как напоминание.

## 🟡 Стоит сделать в ближайшее время

| # | Что | ADR | Оценка | Комментарий |
|---|-----|-----|--------|-------------|
| 1 | RS256 + JWKS вместо HS256 | [ADR-004](adr/004-microservices-patterns.md) | ~2.5 ч | Убирает общий секрет по сервисам. Триггер — второй потребитель JWT |
| 2 | gitleaks / trufflehog в CI | [ADR-006](adr/006-secrets-management.md) | ~1 ч | Защита от случайного коммита секретов |
| 3 | Закрепить версии monitoring-образов | [ADR-013](adr/013-infrastructure-requirements.md) | ~15 мин | `:latest` → `v3`, `11`, `3`, `3`. Триггер — первый инцидент |

## 🟢 Низкий / техдолг

| # | Что | ADR | Оценка | Комментарий |
|---|-----|-----|--------|-------------|
| 4 | Алерт на `/ready` Gateway | [ADR-017](adr/017-alerting.md) | ~1 ч | Нужен отдельный механизм — healthcheck смотрит на `/health` |
| 5 | Алерт на возраст S3-дампов | [ADR-014](adr/014-backups-and-dr.md) | ~1 ч | Fail-open для S3 не покрыт |
| 6 | govulncheck по расписанию | [ADR-012](adr/012-ci-cd-optimization.md) | ~30 мин | Отдельный workflow |
| 7 | Dependabot / Renovate | [ADR-012](adr/012-ci-cd-optimization.md) | ~30 мин | Авто-PR на зависимости |
| 8 | Docker logging options | [ADR-018](adr/018-storage-retention.md) | ~15 мин | `daemon.json` + перезапуск Docker |
| 9 | Rotation SSH на Ed25519 | [ADR-006](adr/006-secrets-management.md), [ADR-021](adr/021-independent-deployment.md) | ~30 мин | `appleboy/ssh-action` уже обновлён |
| 10 | F-2: read-then-write без транзакции | [ADR-007](adr/007-known-issues.md) | ~2 ч | `PatchMyProfile`, `PatchExercise` |
| 11 | Рефакторинг Auth/Profiles по ADR-011 | [ADR-011](adr/011-naming-and-code-style.md) | ~2 ч | N-1, N-3, CS-3, CS-9 |

## 🟡 После недели наблюдений за CB/Retry

| # | Что | ADR | Комментарий |
|---|-----|-----|-------------|
| 12 | Алерты `CircuitBreakerOpen`, `HighRetryRate` | [ADR-020](adr/020-circuit-breaker-and-retry.md) | Нужна статистика порогов |
| 13 | Скорректировать дефолты CB | [ADR-020](adr/020-circuit-breaker-and-retry.md) | `CBMinRequests`, `CBErrorRate` |

## ⏳ Ждёт триггера (не начинать сейчас)

| # | Что | ADR | Триггер |
|---|-----|-----|---------|
| 14 | Bulkheads (семафор на downstream) | [ADR-019](adr/019-microservices-evolution.md), [ADR-020](adr/020-circuit-breaker-and-retry.md) | Реальная нагрузка / инцидент «Gateway упал от затора» |
| 15 | Fallback / degradation по эндпоинтам | [ADR-019](adr/019-microservices-evolution.md) | Появление фронта, обсуждение UX |
| 16 | Auto-rollback (Level 2) | [ADR-012](adr/012-ci-cd-optimization.md) | Дежурный / SLA / болезненный ручной откат |
| 17 | Event bus + Outbox + Saga | [ADR-019](adr/019-microservices-evolution.md) | Cross-service запись в две БД / projections (Leaderboard) |
| 18 | Distributed tracing | [ADR-019](adr/019-microservices-evolution.md) | 5+ сервисов или fan-out (1 запрос → 3+ downstream) |
| 19 | Несколько хостов / k8s / managed БД | [ADR-019](adr/019-microservices-evolution.md) | Реальная нагрузка, сервисов > 10 |
| 20 | mTLS + `GRPC_CLIENT_TLS_MODE=verify` | [ADR-016](adr/016-network-security.md) | Выход из одной docker-сети / внешний gRPC-потребитель |
| 21 | Статический анализ импортов `pkg/**` | [ADR-021](adr/021-independent-deployment.md) | Если `pkg/**` начнёт меняться часто |
| 22 | Blue-green / canary | [ADR-012](adr/012-ci-cd-optimization.md), [ADR-021](adr/021-independent-deployment.md) | Появление SLA |

## Что НЕ делаем (явно зафиксировано)

| Что | ADR |
|-----|-----|
| k8s, Kafka, Service mesh | [ADR-019](adr/019-microservices-evolution.md) |
| Микросервис на каждый CRUD | [ADR-019](adr/019-microservices-evolution.md) |
| Event sourcing / CQRS | [ADR-004](adr/004-microservices-patterns.md) |
| Distributed-транзакции между сервисами | [ADR-003](adr/003-transactions.md) |
| Автоматический инкремент semver | [ADR-021](adr/021-independent-deployment.md) |
| Per-service версионирование | [ADR-021](adr/021-independent-deployment.md) |