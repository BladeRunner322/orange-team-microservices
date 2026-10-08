# ADR-004: Паттерны микросервисов

## Контекст

Проект построен как набор микросервисов. При проектировании мы используем одни паттерны, осознанно отказываемся от других, и откладываем третьи на потом.

Без этой фиксации через несколько месяцев будет непонятно, что было осознанным решением, а что — упущением.

## Решение

### Используем

**Database per Service.**
У каждого сервиса своя БД. Cross-service FK нет. Ссылки между сервисами — по ID без ограничений на уровне БД.

**API Gateway.**
Единая точка входа для клиентов. Gateway валидирует JWT, применяет RBAC, проксирует в downstream-сервисы. Клиент не общается с бизнес-сервисами напрямую.

**Backend for Frontend (BFF) — заложен.**
Gateway умеет агрегировать данные из нескольких сервисов в один ответ клиенту. Целевой пример: `GET /users/me` → `Profiles.GetProfile` + `Workouts.GetUserScore` (параллельно, см. ADR-002).

**Текущее состояние:** агрегация из нескольких сервисов пока не реализована — Workouts не готов. `GET /users/me` обслуживает только Profiles.

**Externalized Configuration.**
Все настройки — через переменные окружения (`.env`-файлы). В git хранятся только `.env.example` (шаблоны) и `.env.enc` (зашифрованные SOPS + age, см. ADR-006).

**Health Check API.**
У каждого сервиса `/health` (liveness) и `/ready` (readiness с проверкой зависимостей). Используется docker-compose и Prometheus.

**Docker healthcheck — какой эндпоинт использовать:**

- Auth, Profiles, Exercises — `/ready`. Каждый сервис зависит только от своей БД, `service_healthy` гарантирует, что сервис не отдаёт трафик до готовности БД.
- **Gateway — `/health`, осознанное исключение.** `/ready` Gateway проверяет Redis + Auth + Profiles + Exercises. Если healthcheck Gateway смотрит на `/ready`, то при недоступности любого downstream Gateway становится `unhealthy`, а Caddy (`depends_on: gateway: service_healthy`) не стартует. Получается каскадный отказ на старте: Auth лежит → Gateway не поднимается → HTTPS недоступен. `/health` разрывает этот цикл: Gateway стартует всегда, а `/ready` остаётся доступен для внешнего балансировщика, если он появится.

**Idempotency.**
Операции, которые могут быть вызваны дважды (register, lazy-create), идемпотентны через `ON CONFLICT DO NOTHING`.

**Stateless Services.**
Сервисы не хранят состояние между запросами. Долгоживущие данные — в БД. Временные — в Redis (refresh-токены, rate limit).

**Centralized Logging & Metrics.**
Prometheus для метрик, Loki + Promtail для логов, Grafana для визуализации. У каждого сервиса `/metrics`.

**Метрики есть только на серверной стороне gRPC.** `MetricsInterceptor` в `pkg/grpc/interceptors` собирает `grpc_requests_total` и `grpc_request_duration_ms` для входящих вызовов. Исходящие gRPC-вызовы (Gateway → Auth/Profiles/Exercises) не инструментированы: latency и error rate на клиентской стороне не видны. Это gap в observability — при инциденте непонятно, тормозит ли сам сервис, или сеть между Gateway и сервисом.

**Что делать:** client-side gRPC interceptor с метриками `grpc_client_requests_total` (по target, method, status) и `grpc_client_request_duration_ms`. Регистрируется в `pkg/grpc/client` рядом с `TimeoutInterceptor` и `UserIDClientInterceptor`. Естественно добавляется вместе с O-2.

**Приоритет:** средний. Триггер — реальная отладка latency-проблем в цепочке Gateway → downstream.

### Не используем (осознанно)

**Saga Pattern.**
Cross-service транзакций нет. По ADR-002 `user_workout_score` нигде не хранится — значит, нет второго места, куда нужно атомарно писать. Компенсаций и саг не требуется.

**Event Sourcing / CQRS.**
Не нужны на текущем масштабе. Домен простой, состояние небольшое. Введение добавит сложность без выгоды.

**Event-Driven Architecture.**
Брокера (NATS / Kafka / Redis Streams) нет. Все межсервисные вызовы — синхронный gRPC. Осознанный выбор в пользу простоты и отладки. См. ADR-003.

### Технический долг (отложено)

**Circuit Breaker — Level B (до прода).**
Не реализован. Если downstream-сервис тормозит, upstream будет ждать таймаут на каждом запросе вместо быстрого возврата ошибки. Реализовать до продакшена.

**Retry с exponential backoff — Level C (после прода).**
Межсервисные вызовы не повторяются при временных сбоях сети. Реализовать после первого прода, только для идемпотентных вызовов, чтобы не плодить дубли.

**`GRPC_CLIENT_TLS_MODE=verify` в проде.**
В dev/staging — `insecure` (self-signed сертификаты). В проде — обязательно `verify` после установки HTTPS/Caddy. Требование не зафиксировано в `.env.example`, добавить комментарий.

**gRPC deadlines на стороне клиента.**
~~На клиентах не установлен deadline.~~ **Закрыто:** `pkg/grpc/client` применяет `interceptors.TimeoutInterceptor(cfg.Timeout)` — каждый вызов ограничен deadline'ом. См. ADR-007 C-1.

**Distributed Tracing.**
`x-request-id` пробрасывается сквозным образом между сервисами (`pkg/ctxkeys` + gRPC-интерсепторы, см. O-2 в ADR-007). Этого достаточно для отладки на текущем масштабе: один `request_id` в логах Gateway, Auth, Profiles, Exercises связывает запрос через всю цепочку.

**Чего не хватает для полноценного tracing:**
- Span-и и parent-child связи между вызовами (нужен OpenTelemetry или аналог).
- Проброс `trace_id` / `span_id` в формате W3C.
- Визуализация трасс (Jaeger, Tempo).

**Триггер для перехода на OpenTelemetry:** число сервисов превысит 5, или появится сценарий с fan-out (один запрос → 3+ downstream-вызова, latency которых надо разложить по компонентам). Сейчас такого нет: цепочки короткие (Gateway → один сервис), `x-request-id` покрывает потребности отладки.

**Rate limiting между сервисами.**
Rate limit есть только на Gateway (для клиентов). Внутренние gRPC-вызовы не ограничены. Добавить при необходимости.

**Service Mesh / mTLS.**
В одной docker-сети не требуется. Понадобится при выходе за пределы одной сети (несколько хостов, разные окружения).

## Последствия

**Плюсы:**

- Простая модель взаимодействия: sync gRPC + БД per service.
- Нет распределённых транзакций, саг, outbox.
- Легко рассуждать о консистентности.
- Минимум инфраструктуры (нет брокера).

**Минусы:**

- Downstream-сервис упал → upstream тоже падает (нет деградации).
- Нет eventual consistency для проекций.
- Отсутствие Circuit Breaker (Level B) и retry (Level C) делает систему хрупкой при сбоях сети.
- При росте числа сервисов отладка без tracing будет сложной.

**Когда пересматривать:**

- Если понадобятся cross-service транзакции — вводить событийную шину и outbox.
- Если начнутся каскадные отказы из-за недоступности downstream — Circuit Breaker и retry.
- Если число сервисов превысит 5 — Distributed Tracing.
- Если сервисы выйдут за пределы одной docker-сети — mTLS.

## Связанные решения

- [ADR-001: Профиль пользователя](001-profile.md) — lazy-create, nullable поля.
- [ADR-002: `user_workout_score`](002-user-workout-score.md) — почему saga не нужна.
- [ADR-003: Транзакции](003-transactions.md) — sync-вызовы вместо распределённых транзакций.