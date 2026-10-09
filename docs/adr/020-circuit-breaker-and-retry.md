# ADR-020: Circuit Breaker и Retry для gRPC-клиентов

## Контекст

После закрытия Level B и C в проекте работают четыре сервиса
(Auth, Gateway, Profiles, Exercises) с синхронным gRPC между
ними. `TimeoutInterceptor` в `pkg/grpc/client` ограничивает
каждый вызов дедлайном (`GATEWAY_TIMEOUT`, по умолчанию 10s).

**Проблема.** Если downstream-сервис начинает тормозить или
падать, Gateway продолжает слать ему запросы и ждать таймаут на
каждом. Это приводит к каскадному отказу:

- Profiles лежит → Gateway копит запросы (по одному на каждый
  входящий HTTP-запрос) → у Gateway заканчиваются горутины и
  соединения → Gateway падает → Caddy возвращает 502 клиентам.

`TimeoutInterceptor` ограничивает каждый отдельный вызов, но
**не ограничивает общий поток** запросов к больному downstream.
Это классический сценарий, для которого придуман Circuit Breaker.

**Второй gap.** В ADR-004 явно зафиксировано: метрики есть только
на серверной стороне gRPC (`MetricsInterceptor` собирает
`grpc_requests_total` для входящих вызовов). Исходящие вызовы
Gateway → Auth/Profiles/Exercises не инструментированы — при
инциденте непонятно, тормозит ли сам сервис, или сеть между
Gateway и сервисом.

**Третий gap.** Временные сетевые сбои (одна потерянная TCP-сессия,
кратковременный `Unavailable` при рестарте контейнера) приводят к
ошибке клиенту, хотя повторный вызов через 100ms был бы успешен.

Это ADR закрывает все три и реализует **Уровень 2 из ADR-019**
(Failure Isolation) в части Circuit Breaker + Retry. Bulkheads и
fallback остаются отдельными задачами.

## Решение

### 1. Circuit Breaker: `sony/gobreaker/v2`

**Библиотека:** `github.com/sony/gobreaker/v2`.

**Почему библиотека, а не свой CB:**

- CB — тот случай, где дьявол в деталях: half-open timing, sliding
  window, concurrent half-open, race при переходе состояний.
  Свой CB на 150-200 строк потребует тестов на каждое состояние
  и каждую гонку.
- `sony/gobreaker` — фактический стандарт в Go-проде (используется
  в Docker, Uber, GitLab), 400 строк, ноль транзитивных зависимостей.
- `OnStateChange` даёт бесплатный hook для метрик и логов.

**Почему не альтернативы:**

- `hystrix-go` — мёртв (последний коммит 2020), нет `context.Context`,
  известные race-баги.
- `failsafe-go` — всё в одном (CB + Retry + Timeout + Bulkhead),
  но это большая зависимость со своим DSL. Избыточно для одной задачи.

**Один breaker на target.** `gobreaker.NewCircuitBreaker` создаётся
в `pkg/grpc/client.New` один раз на каждый `Target`. Это значит:
если Profiles лежит — открывается breaker только для Profiles,
Auth и Exercises продолжают работать нормально.

**Состояния и переходы:**

```
CLOSED  ──(> ErrorRate ошибок за Interval, min Requests)──►  OPEN
OPEN    ──(Timeout истёк)───────────────────────────────►  HALF-OPEN
HALF-OPEN ──(MaxRequests успешных подряд)────────────────►  CLOSED
HALF-OPEN ──(1 ошибка)───────────────────────────────────►  OPEN
```

В `OPEN` состояние `cb.Execute` сразу возвращает `gobreaker.ErrOpenState`
без вызова downstream — это и есть fail-fast.

### 2. Retry: `grpc-ecosystem/go-grpc-middleware/retry`

**Библиотека:** `github.com/grpc-ecosystem/go-grpc-middleware/retry`.

**Почему:** официальный интерсептор экосистемы gRPC. Exponential
backoff с jitter, ограничение по кодам ответа, ограничение попыток.
Уже проверен в проде, не требует своей реализации.

**Коды для retry:** только `codes.Unavailable` и
`codes.DeadlineExceeded`. Не `Internal`, не `InvalidArgument`, не
`NotFound`. Логические ошибки не ретраим — только транспортные.

### 3. Whitelist идемпотентных методов

Retry применяется **только** к методам из явного whitelist.

**Критерий отбора:** read-only методы — те, у которых нет
side-эффектов в БД или в других сервисах. Повторный вызов не
изменяет состояние системы.

**Текущий whitelist:**

```
/auth.AuthService/ValidateToken
/profiles.ProfilesService/GetMyProfile
/profiles.ProfilesService/GetProfile
/exercises.ExercisesService/GetExercise
/exercises.ExercisesService/GetExercises
```

**Мутации не ретраим даже если они идемпотентны.**

Пример: `DeleteExercise` (Exercises) идемпотентен по ADR-007 F-3 —
повторный вызов возвращает успех, потому что `MarkDeleted` фильтрует
`WHERE deleted_at IS NULL`. Формально его retry безопасен.

Мы всё равно **не включаем** его в whitelist. Причины:

1. **Простота правила.** «Retry только для read-методов» не требует
   от разработчика каждый раз доказывать идемпотентность нового
   метода. Мутация → не в whitelist. Read → в whitelist. Точка.
2. **Устойчивость к эволюции.** Если завтра `DeleteExercise` начнёт
   писать в outbox-таблицу (Уровень 3 из ADR-019) — идемпотентность
   сломается. Whitelist из read-методов такого риска не создаёт.
3. **Реальной выгоды почти нет.** DELETE через Gateway — редкая
   операция от админа. Timeout на DELETE — редкость. Если он
   случится, админ нажмёт «удалить» ещё раз руками.

**Технически:** `grpc_retry` ретраит все методы, чей код ответа в
`WithCodes`. Чтобы ограничить whitelist, пишем свой тонкий
интерсептор `RetryInterceptor(methods []string, ...)`, который для
метода не из списка вызывает `invoker` напрямую, а для метода из
списка — делегирует в `grpc_retry`.

### 4. Порядок интерсепторов

**Итоговая цепочка:**

```
CB → Retry → Timeout → UserID → RequestID → invoker
```

**Обоснование порядка:**

- **CB снаружи Retry.** CB видит **финальный** результат retry-цикла.
  Если retry добился успеха со 2-й попытки — CB считает это успехом,
  breaker не открывается. Если бы CB был внутри Retry, каждый retry
  считался бы отдельным запросом в статистику breaker'а.
- **Retry снаружи Timeout.** Таймаут применяется к **каждой попытке**,
  а не ко всему retry-циклу. Иначе `3 × 10s = 30s` на один вызов
  превысит `GATEWAY_TIMEOUT` и весь бюджет запроса клиента.
- **UserID и RequestID внутри.** Это финализирующие интерсепторы,
  которые кладут данные в metadata непосредственно перед вызовом.
  Они должны отработать один раз на каждую попытку, включая retry
  (иначе в metadata второй попытки не будет `user_id`).

**Это breaking change для observability.** Раньше `TimeoutInterceptor`
был самым внешним — `grpc_request_duration_ms` на сервере включал
один вызов. Теперь при retry на сервере будет N записей, каждая
со своим временем. Метрики retry (`grpc_client_retry_attempts_total`)
дадут полную картину.

### 5. Конфигурация

**Новые поля в `pkg/grpc/client.Config`:**

```go
type Config struct {
    // ... существующие Target, TLSMode, Timeout

    // Circuit Breaker
    CBEnabled     bool
    CBMaxRequests uint32        // half-open: сколько запросов пропустить подряд
    CBInterval    time.Duration // окно подсчёта ошибок
    CBTimeout     time.Duration // пауза в OPEN перед переходом в HALF-OPEN
    CBMinRequests uint32        // минимум запросов в окне для расчёта ErrorRate
    CBErrorRate   float64       // порог доли ошибок для перехода CLOSED → OPEN

    // Retry
    RetryEnabled     bool
    RetryMaxAttempts uint32
    RetryBaseDelay   time.Duration
    RetryMaxDelay    time.Duration
    RetryMethods     []string // whitelist полных имён методов
}
```

**Значения по умолчанию:**

| Параметр | Значение | Обоснование |
|---|---|---|
| `CBEnabled` | `true` | Иначе смысла нет |
| `CBMaxRequests` | `1` | Half-open пропускает 1 запрос — минимальный риск |
| `CBInterval` | `60s` | Окно подсчёта за минуту — сглаживает всплески |
| `CBTimeout` | `30s` | После открытия ждём 30s до первой проверки |
| `CBMinRequests` | `10` | Не открывать breaker на 2-3 случайных ошибках |
| `CBErrorRate` | `0.5` | >50% ошибок = open |
| `RetryEnabled` | `true` | |
| `RetryMaxAttempts` | `3` | Стандарт: 1 + 2 retry |
| `RetryBaseDelay` | `100ms` | |
| `RetryMaxDelay` | `1s` | Ограничивает общий бюджет |
| `RetryMethods` | см. п.3 | Явный whitelist |

**Env-переменные в `services/gateway/.env`:**

```
GATEWAY_CB_ENABLED=true
GATEWAY_CB_MAX_REQUESTS=1
GATEWAY_CB_INTERVAL=60s
GATEWAY_CB_TIMEOUT=30s
GATEWAY_CB_MIN_REQUESTS=10
GATEWAY_CB_ERROR_RATE=0.5

GATEWAY_RETRY_ENABLED=true
GATEWAY_RETRY_MAX_ATTEMPTS=3
GATEWAY_RETRY_BASE_DELAY=100ms
GATEWAY_RETRY_MAX_DELAY=1s
GATEWAY_RETRY_IDEMPOTENT_METHODS=/auth.AuthService/ValidateToken,/profiles.ProfilesService/GetMyProfile,/profiles.ProfilesService/GetProfile,/exercises.ExercisesService/GetExercise,/exercises.ExercisesService/GetExercises
```

Прокидываются в `clients.NewAuthClient(...)`, `NewProfilesClient(...)`,
`NewExercisesClient(...)` — через расширенную сигнатуру или через
`grpcclient.Config` напрямую.

### 6. Метрики

В `pkg/metrics/grpc.go` добавляются **клиентские** метрики. Это
закрывает gap из ADR-004.

| Метрика | Тип | Labels | Что показывает |
|---|---|---|---|
| `grpc_client_requests_total` | Counter | `target, method, status` | Сколько исходящих вызовов, с каким кодом |
| `grpc_client_request_duration_ms` | Histogram | `target, method` | Сколько Gateway ждёт ответа |
| `grpc_client_requests_in_flight` | Gauge | `target` | Сколько вызовов висит сейчас |
| `grpc_client_circuit_breaker_state` | Gauge | `target` | `0=closed`, `1=half-open`, `2=open` |
| `grpc_client_circuit_breaker_transitions_total` | Counter | `target, from, to` | Сколько раз CB менял состояние |
| `grpc_client_retry_attempts_total` | Counter | `target, method, result` | Сколько было retry-попыток (`result: success\|failure`) |

**Источники данных:**

- `ClientMetricsInterceptor` (новый) — первые три, пишет на каждый вызов.
- `gobreaker.Settings.OnStateChange` — следующие два, вызывается
  библиотекой при каждом переходе.
- `RetryInterceptor` (свой) — последний, инкрементит на каждую
  retry-попытку.

**Что это даёт при инциденте:**

```
# Доля ошибок Gateway → Profiles
sum(rate(grpc_client_requests_total{target="profiles-service:50052",status!="ok"}[5m]))
  / sum(rate(grpc_client_requests_total{target="profiles-service:50052"}[5m]))

# Открыт ли breaker?
grpc_client_circuit_breaker_state == 2

# Сколько retry-попыток в секунду
rate(grpc_client_retry_attempts_total[5m])
```

Без этих метрик поведение CB и Retry невидимо в Grafana — это
противоречит принципу «каждое новое поведение должно быть
наблюдаемым» из ADR-017.

**Алерты — отдельная задача.** После сбора статистики за неделю:

- `CircuitBreakerOpen` — `grpc_client_circuit_breaker_state{target=~"..."} == 2`
  более 1 минуты → warning.
- `HighRetryRate` — `rate(grpc_client_retry_attempts_total{result="success"}[5m]) > 0.2`
  → warning.

В этом PR алерты не добавляем.

### 7. Что НЕ делаем в этом PR

**Bulkheads (семафор на downstream).** Ограничение числа
параллельных вызовов к каждому target. Отдельная задача Уровня 2
из ADR-019. CB уже даёт большую часть выгоды — bulkheads нужны для
тонкой настройки при высокой нагрузке.

**Fallback / degradation.** Возврат дефолтного значения по
эндпоинту при недоступности downstream (например, `profile_completed: false`
вместо 500 на `GET /users/me`). Требует продуктового решения по
каждому эндпоинту — отдельная задача.

**Свой CB.** См. п.1 — библиотека надёжнее.

**Hystrix-go.** Мёртв.

**Distributed tracing для retry.** При retry один HTTP-запрос
превращается в N gRPC-запросов с одним `request_id`. Отладка
по логам — возможна (все N записей с одним `request_id`). Span-ы
и parent-child связи — Уровень 4 из ADR-019.

## Последствия

**Плюсы:**

- **Fail-fast вместо таймаута.** Если Profiles лежит, Gateway
  отвечает клиенту за миллисекунды (`503` с сообщением
  `circuit breaker is open`), а не ждёт 10 секунд.
- **Защита от каскадного отказа.** Gateway не уходит в OOM от
  накопленных горутин при недоступности downstream.
- **Восстановление автоматически.** После `CBTimeout` breaker
  переходит в HALF-OPEN, пропускает один запрос, и если он
  успешен — закрывается. Никакого ручного вмешательства.
- **Retry сглаживает временные сбои.** Одна потерянная сессия
  или `Unavailable` при рестарте контейнера теперь не доходит
  до клиента.
- **Client-side observability.** Закрывается gap из ADR-004.
- **Осознанное правило по мутациям.** Whitelist read-методов
  защищает от дублей при retry.

**Минусы:**

- **Ещё одна зависимость** — `sony/gobreaker/v2`.
- **Сложнее отладка.** При retry на сервере N записей с одним
  `request_id`. Нужно помнить об этом при чтении логов.
- **Настройка параметров CB — компромисс.** Слишком агрессивные
  (низкий `CBMinRequests`, высокий `CBErrorRate`) → breaker
  срабатывает на нормальных всплесках. Слишком мягкие → не
  защищает. Значения по умолчанию подобраны консервативно;
  пересмотр — после недели наблюдений.
- **Breaking change для метрик таймаута.** Раньше `TimeoutInterceptor`
  был внешним — теперь внутренний, таймаут применяется к попытке,
  не к циклу. Метрики длительности могут измениться.
- **Нет алертов сразу.** Пока не накопим статистику — не знаем,
  какие пороги адекватны.

**Когда пересмотреть:**

- **После недели наблюдений.** Настроить алерты `CircuitBreakerOpen`
  и `HighRetryRate`, скорректировать дефолты.
- **При появлении каскадных отказов даже с CB.** Bulkheads —
  следующая задача Уровня 2.
- **При росте трафика.** `CBMinRequests = 10` может быть мало
  при тысячах RPS — окно в 60s будет содержать тысячи запросов.
- **При добавлении Workouts.** Workouts зовёт Profiles и Exercises.
  Retry-политика для Workouts → Profiles может отличаться от
  Gateway → Profiles. Возможно, конфиг станет per-target.
- **При появлении event bus (Уровень 3).** Outbox-паттерн может
  сделать часть мутаций идемпотентными — тогда whitelist
  пересматривается.
- **Если появится внешний gRPC-потребитель.** Сейчас `insecure` gRPC
  (ADR-016). При выходе из docker-сети — mTLS, и CB-конфиг,
  возможно, станет строже.

## Связанные решения

- [ADR-004: Паттерны микросервисов](004-microservices-patterns.md) —
  client-side gRPC-метрики как gap, CB/Retry как tech debt.
- [ADR-007: Известные проблемы](007-known-issues.md) — C-1
  (gRPC deadlines на клиентах), F-3 (идемпотентность DeleteExercise).
- [ADR-009: Карта зависимостей сервисов](009-service-dependencies.md) —
  какие вызовы существуют.
- [ADR-017: Alerting](017-alerting.md) — принцип «новое поведение
  должно быть наблюдаемым».
- [ADR-019: Эволюция в сторону микросервисов](019-microservices-evolution.md) —
  Уровень 2, Failure Isolation.

## Статус

**Planned.** Реализуется в PR `feat/circuit-breaker-retry`.

Что войдёт в PR:

- `pkg/grpc/client` — расширенный `Config`, создание `gobreaker.CircuitBreaker`.
- `pkg/grpc/interceptors` — `CircuitBreakerInterceptor`, `RetryInterceptor`,
  `ClientMetricsInterceptor`.
- `pkg/metrics/grpc.go` — 6 новых клиентских метрик.
- `services/gateway/config` — новые env-поля.
- `services/gateway/.env.example` — документация переменных.
- `pkg/grpc/client/client.go` — новый порядок интерсепторов.
- Тесты: unit для интерсепторов (mock invoker), integration для CB
  (реальный gobreaker в тесте), unit для whitelist.