# ADR-022: Стратегия тестирования

## Контекст

В проекте три уровня тестов:

- **Unit** — `go test -short`, без внешних зависимостей.
- **Integration** — `go test -tags=integration`, с Testcontainers (Postgres, Redis).
- **E2E** — Postman-коллекции против запущенного стека.

Нигде не зафиксировано, что покрываем каким уровнем, а что не покрываем вообще. Из-за этого:

- В `task test` было 12 строк «0.0%» для обвязки (`cmd`, `config`, `bootstrap`, `internal/gen`, `pkg/logger`, `pkg/health`, `pkg/grpc/server`, `pkg/nullable`, `pkg/redis`, `pkg/metrics`, `pkg/ratelimit`). Выглядело как дыры, хотя осознанно так — у этих пакетов либо нет логики для unit-тестов, либо она покрывается через integration/e2e.
- Недавний PR формализовал exclusion в `Taskfile.yml` через переменную `COVERAGE_EXCLUDE`. Но без ADR — через месяц будет непонятно, что было решением, а что случайностью.

## Решение

### Что покрываем unit-тестами

- `domain` — VO, сущности, `ApplyPatch`, доменные ошибки.
- `usecases` — сценарии с моками репозиториев.
- `mappers` — proto ↔ domain, model ↔ domain.
- `middleware` — чистые функции над `http.Request` / `http.ResponseWriter`.
- `httputil` — `SendJSON`, `SendError`, `GrpcErrorToHTTP`, `StatusRecorder`.
- `pkg/ctxkeys`, `pkg/grpc/client`, `pkg/grpc/interceptors` — инфраструктура с логикой.

### Что покрываем integration-тестами (Testcontainers)

- `infrastructure/postgres_repo` — SQL, маппинг ошибок, транзакции.
- `infrastructure/redis_repo` — TTL, set-операции.
- `pkg/ratelimit` — Lua-скрипт, атомарность Token Bucket.
- `pkg/postgres` — `WithTx`, `mapErrors` с реальным Postgres.

Integration-тесты помечены `//go:build integration` и в `task test` не участвуют.

### Что НЕ покрываем

- `cmd/` — точка входа, `main`.
- `config/` — envconfig-обёртка.
- `bootstrap/` — DI-сборка, тестируется через e2e.
- `internal/gen/` — сгенерированный proto-код.
- `interfaces/*grpc` — обвязка над usecase, тестируется e2e.
- `application/ports` — интерфейсы, нет кода для покрытия.
- `pkg/logger`, `pkg/health`, `pkg/nullable`, `pkg/redis`, `pkg/metrics`, `pkg/grpc/server` — тонкие обёртки.

### Как это зафиксировано в коде

В `Taskfile.yml` две переменные:

- `UNIT_COVERAGE_EXCLUDE` — паттерны пакетов, исключаемых из unit-покрытия. Две группы: обвязка (нет логики для unit-тестов) и пакеты с только integration-тестами (0% в unit-режиме вводит в заблуждение).

```
internal/gen|/cmd$|/config$|/bootstrap$|/interfaces/.*grpc$|/application/ports$|pkg/logger$|pkg/health$|pkg/grpc/server$|pkg/nullable$|pkg/redis$|pkg/metrics$|pkg/ratelimit$|/infrastructure/postgres_repo$|/infrastructure/redis_repo$|/infrastructure/clients$
```

- `INTEGRATION_PACKAGES` — явный список пакетов, покрываемых integration-тестами.

```
./pkg/postgres/... ./pkg/ratelimit/... ./services/auth/internal/infrastructure/postgres_repo/... ./services/auth/internal/infrastructure/redis_repo/... ./services/profiles/internal/infrastructure/postgres_repo/... ./services/exercises/internal/infrastructure/postgres_repo/...
```

- `task test` и `task test-cover` фильтруют через `go list ./... | grep -vE '{{.UNIT_COVERAGE_EXCLUDE}}'`.
- `task test-integration` использует `-cover` и явный список `{{.INTEGRATION_PACKAGES}}`, показывает покрытие по integration-пакетам.

### Правила

1. **Новый код** в `domain`, `usecases`, `mappers`, `middleware`, `httputil` — покрывается unit-тестами.
2. **Работа с внешними системами** (Postgres, Redis) — integration-тесты с Testcontainers.
3. **Обвязка** — не покрывается. Новый пакет-обёртка добавляется в `COVERAGE_EXCLUDE`.
4. **Total coverage** считается только по testable-пакетам. Цель — не «100%», а отсутствие непокрытых ветвлений в бизнес-логике.

## Последствия

**Плюсы:**

- `task test` показывает реальную картину, не разбавленную 0% обвязки.
- Новый сервис сразу знает, что писать, что не писать.
- Не тратим время на unit-тесты для `main.go`.

**Минусы:**

- `COVERAGE_EXCLUDE` надо поддерживать — при добавлении нового пакета-обвязки не забыть в список.
- «Total coverage» в отчёте не совпадает с наивным «процент по всем пакетам». Может смутить при первом взгляде.

## Связанные решения

- [ADR-011: Naming conventions и Code style](011-naming-and-code-style.md)