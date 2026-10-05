# ADR-007: Известные проблемы и технический долг

## Контекст

При разработке Auth, Gateway и Profiles накопились замечания: где-то
баг, где-то осознанное упрощение, где-то недоделка. Часть уже
устранена (см. PR #17 — интеграция Profiles, PR #19 — security & config
fixes), но остаётся список того, что нужно залатать до продакшена или
осознанно отложить.

Без фиксации в одном месте через месяц будет непонятно, что было
осознанным решением, а что — упущением. Этот ADR закрывает именно
этот пробел.

Ничто из перечисленного не блокирует локальный запуск и dev-сценарии.

## Решение

Фиксируем список в одном документе. Часть пунктов — блокеры для
продакшена, часть — приятные улучшения.

### Безопасность

**B-1. JWT: не проверяются `iss` и `aud`.**
`services/auth/internal/infrastructure/jwt/manager.go` вызывает
`jwt.ParseWithClaims` без `jwt.WithIssuer` и `jwt.WithAudience`.
Подпись проверяется, но claims `iss`/`aud` — нет.
Токен, подписанный тем же секретом, но выпущенный для другого
сервиса/аудитории, пройдёт валидацию. Пока секрет только у Auth —
риск низкий; при переходе на общий секрет между сервисами — дыра.
**Что делать:** добавить `jwt.WithIssuer(m.issuer)`, `jwt.WithAudience(m.audience)`.
**Приоритет:** до продакшена.
**Статус:** ✅ Закрыто в PR #19 (2026-10-01).

**B-2. Rate limit: доверие `X-Forwarded-For` без whitelist прокси.**
`services/gateway/internal/interfaces/http/middleware/rate_limit.go`,
`clientIP()`. При публичной доступности Gateway клиент подделает
заголовок и обойдёт лимит по IP (а с ним — и защиту от брутфорса
`/login`).
**Что делать:** либо конфигурируемый список trusted proxies,
либо не доверять XFF (в docker-сети прокси нет).
**Приоритет:** до продакшена.
**Статус:** ✅ Закрыто в PR #19 (2026-10-01).

**B-3. `GetProfile(user_id)` в Profiles не проверяет права доступа.**
По gRPC-контракту `user_id` приходит в теле запроса, никакой
авторизации. Любой, кто дотянется до gRPC-порта Profiles, читает
чужой профиль. Осознанное решение (ADR-001), но стоит явно
задокументировать, что gRPC-порты доступны только внутри сети.
**Что делать:** ADR-комментарий или mTLS между сервисами.
**Приоритет:** документирование — сейчас.

### Функциональные баги

**F-1. `Register` в Auth глотает ошибки БД.**
`services/auth/internal/application/usecases/register.go`:
`if _, err := repo.FindByEmail(...); err == nil { duplicate }`.
Любая ошибка ≠ `ErrUserNotFound` (timeout, сеть, ошибка запроса)
молча игнорируется, код идёт на INSERT.
**Что делать:** `if !errors.Is(err, domain.ErrUserNotFound) { return ... }` до проверки на duplicate.
**Приоритет:** до продакшена.
**Статус:** ✅ Закрыто в PR #19 (2026-10-01).

**F-2. `PatchMyProfile` (Profiles) и `PatchExercise` (Exercises) — read-then-write без транзакции.**
`services/profiles/internal/application/usecases/patch_my_profile.go`,
`services/exercises/internal/application/usecases/patch_exercise.go`.

Оба usecase читают сущность через `GetByXxx` / `GetExercise`, применяют
патч в памяти, затем пишут через `Update`. Между SELECT и UPDATE есть
окно: если два патча одного ресурса идут параллельно, оба читают
одну версию, оба применяют свой патч к ней, второй UPDATE затирает
результат первого — **lost update** (классический last-write-wins).

Ресурс админский (профиль — свой, упражнение — только admin), патчи
редкие, вероятность одновременного патча одного ресурса низкая.
Осознанно принято как компромисс.

**Что делать:** `SELECT ... FOR UPDATE` внутри транзакции через
`pool.WithTx` (пессимистичная блокировка), либо оптимистичная через
колонку `version` + `WHERE version = $N` + `409 Conflict` на 0 rows.
**Приоритет:** низкий. Триггер на исправление — жалобы на потерянные
правки или рост числа админов.

**F-3. Идемпотентность операций удаления.**

Три операции удаления/отзыва в проекте:

| Операция | Поведение при повторе | Где |
|---|---|---|
| `Auth.Logout` | идемпотентен | `services/auth/internal/application/usecases/logout.go` |
| `Exercises.DeleteExercise` | **идемпотентен** | `services/exercises/internal/application/usecases/delete_exercise.go` |
| `Profiles.DeleteMyProfile` | не идемпотентен → NotFound | `services/profiles/internal/application/usecases/delete_my_profile.go` |

**Решение.** DELETE-подобные операции проектируем **идемпотентными**:
повторный вызов возвращает успех, не NotFound. Это защищает от
retry-проблем: прокси, load balancer или клиент может безопасно
повторить запрос после network timeout, не получив ложный 404.

- `Logout` — идемпотентен изначально, оставляем.
- `DeleteExercise` — идемпотентен: soft delete, повторный `MarkDeleted`
  даёт `0 rows affected`, usecase возвращает `nil`.
- `DeleteMyProfile` — **отстаёт**, привести к идемпотентному стилю
  при следующем касании Profiles.

**Приоритет:** низкий для `DeleteMyProfile` (рефакторинг Profiles),
не блокер.

**Связанные контракты:** `DELETE /exercises/{id}` в Gateway после
реализации должен возвращать `204` и на первый, и на повторный вызов.

### Конфигурация

**C-1. `GATEWAY_TIMEOUT` и `pkg/grpc/client.Config.Timeout` не применяются.**
Флаги объявлены, но нигде не используются. Если Auth/Profiles зависнет,
вызов будет ждать TCP-таймаут.
**Что делать:** на gRPC-клиентах применять
`context.WithTimeout(ctx, cfg.Timeout)` на каждый вызов, либо задать
дедлайн на уровне клиента.
**Приоритет:** до продакшена.
**Статус:** ✅ Закрыто в PR #19 (2026-10-01).

**C-2. `ENABLE_TLS` / `TLSCertFile` / `TLSKeyFile` в Gateway не влияют.**
`clients.NewAuthClient` / `NewProfilesClient` хардкодят `TLSModeInsecure`.
Флаг есть — эффекта нет.
**Что делать:** прокидывать TLSMode из конфига, либо убрать неиспользуемые переменные.
**Приоритет:** средний (вводит в заблуждение).
**Статус:** ✅ Закрыто (2026-10-06). `ENABLE_TLS`/`TLSCertFile`/`TLSKeyFile` убраны из конфига Gateway.
Добавлено поле `GRPC_CLIENT_TLS_MODE` (значения `disabled`/`insecure`/`verify`),
прокидывается в `NewAuthClient`, `NewProfilesClient`, `NewExercisesClient`.

### Наблюдаемость

**O-1. `/ready` в Gateway проверяет только Redis.**
Не проверяет доступность Auth и Profiles. Если Auth лёг —
Gateway считает себя готовым, но защищённые запросы будут падать.
**Что делать:** добавить в `/ready` проверку gRPC-коннектов к Auth/Profiles.
**Приоритет:** средний.
**Статус:** ✅ Закрыто (2026-10-05). `/ready` проверяет Redis + gRPC-каналы
к Auth, Profiles, Exercises. Каждый клиент имеет метод `IsHealthy(ctx) error`,
основанный на `grpc.conn.GetState()`. Readiness-логика вынесена в `pkg/health`.

### CI/CD

**CI-1. Интеграционные тесты не запускаются в CI.**
`.github/workflows/ci-cd.yml` выполняет только `go test -v ./... -short`
(юнит-тесты). Интеграционные тесты (`-tags=integration`, Testcontainers)
не прогоняются. Регрессии в интеграциях (репозитории Auth/Profiles,
rate limiter) CI не поймает.
**Что делать:** добавить шаг с `go test -tags=integration ./...` —
Docker на GitHub Actions доступен, Testcontainers работает.
**Приоритет:** средний.
**Статус:** ✅ Закрыто (2026-10-05). Добавлен шаг `Run Integration Tests`
с `-tags=integration -timeout=15m`. См. ADR-012.

### Мелкие

**S-1. `Upsert` в Profiles пишет только user_id + created_at.**
`services/profiles/internal/infrastructure/postgres_repo/repository.go`.
Остальные поля игнорируются. Имя вводит в заблуждение — оно про
lazy-create, а не про «сохранить всё».
**Что делать:** переименовать в `EnsureExists` или сделать полноценный upsert.
**Приоритет:** низкий.

**S-2. `test_logger.NewTestLogger` без `file`.**
`pkg/logger/test_logger.go` создаёт `&Logger{Logger: ...}` без `file`.
При `Close()` — паника на nil.
**Что делать:** `Close()` проверять `if l.file != nil`.
**Приоритет:** низкий.
**Статус:** ✅ Закрыто (2026-10-05). `Logger.Close()` теперь возвращает nil при `file == nil`.

**S-3. Две обёртки `ResponseWriter` в Gateway middleware.**
`logger.go` (`responseWriterWrapper`) и `metrics.go`
(`metricsResponseWriter`) вложены друг в друга. Работает, но хрупко —
при добавлении третьей легко рассинхронить статус.
**Что делать:** одна общая обёртка в `httputil`.
**Приоритет:** низкий.

**S-4. `log.Info` в usecase дублируется с `LoggingInterceptor`.**

Все usecase (Auth, Profiles, Exercises) логируют `Info` в конце операции: `"user registered successfully"`, `"profile patched"`, `"exercise fetched"`. При этом `LoggingInterceptor` уже логирует каждый gRPC-вызов с method, status и duration. Получается две записи на один успешный запрос.

Для мутаций (`CreateExercise`, `PatchExercise`, `DeleteExercise`, `Register`) `Info` иногда рассматривают как бизнес-аудит: «кто что сделал». Но в Exercises аудит неполный — сервис не знает `user_id` (его использует только Gateway), поэтому запись «упражнение X удалено» без «кем» ценности не несёт. Полные данные всегда можно получить из БД.

**Решение (принято в PR по Exercises):**
- **Exercises — вариант B.** `log.Info` в usecase убран полностью. Логирование успешных операций — только на уровне `LoggingInterceptor`. `log.Warn` и `log.Error` в usecase остаются: они несут информацию, которой у интерсептора нет (`invalid name` vs `invalid difficulty` vs `db is down`).
- **Auth, Profiles — пока вариант A** (Info везде). Устаревший стиль, привести к B при следующем касании соответствующего usecase. Отдельным PR.
- **Правило для новых сервисов (Habits, Workouts, Leaderboard):** вариант B. `Info` в usecase не пишем с самого начала.

**Приоритет:** низкий для Auth/Profiles (рефакторинг), не блокер.

**S-5. `pkg/postgres` не различает constraint при unique violation.**

`pkg/postgres/adapters.go`, функция `mapErrors`. Postgres возвращает
один и тот же код `23505` для любого нарушения unique-индекса, вне
зависимости от того, какой именно constraint сработал. `mapErrors`
смотрит только на код, но не на `pgErr.ConstraintName`, и возвращает
общую `ErrViolatesUnique`.

Репозитории маппят её в **доменную** ошибку — сейчас это корректно,
потому что в каждой таблице один unique-индекс:

- Auth → `email` → `ErrEmailAlreadyExists`
- Exercises → partial `name WHERE deleted_at IS NULL` → `ErrExerciseNameExists`
- Profiles → unique не используется

Проблема станет реальной, когда в любой таблице появится **второй**
unique-индекс. Тогда оба нарушения будут маппиться в одну и ту же
доменную ошибку — клиент увидит неверное сообщение («name already
exists» вместо «slug already exists»).

**Что делать:** расширить `mapErrors` или добавить хелпер
`ConstraintNameFromError(err) string` в `pkg/postgres`, чтобы
репозиторий сам решал, что маппить. Правка `pkg/` — затрагивает Auth
и Profiles одновременно.
**Приоритет:** низкий. Триггер — появление второго unique-индекса
в любой таблице любого сервиса.

### Отложено из ранее принятых ADR

- **ADR-004:** circuit breaker, retry с backoff, distributed tracing,
  mTLS — осознанно отложены «до продакшена». (gRPC deadlines на
  клиентах — закрыты, см. C-1.)
- **ADR-006:** SOPS + age. Статус «принято, внедрение — отдельным PR».
  Фактически `.env` на сервере до сих пор правятся руками.

## Последствия

**Плюсы:**
- Список в одном месте, не теряется в чатах и PR.
- Понятно, что блокирует прод, а что можно отложить.
- При добавлении нового сервиса можно пройтись по этому чек-листу.

**Минусы:**
- Список разрастётся, надо будет ревизовать.
- Часть пунктов может устареть, если код поправили — не забывать
  обновлять ADR при закрытии.

## Приоритеты

1. **До продакшена:** внедрить ADR-006 (SOPS + age — принято, но не реализовано).
2. **Средний:** — (закрыто).
3. **Низкий:**
   - **F-2** — открыт (Profiles, Exercises).
   - **F-3** — частично закрыт: Exercises идемпотентен ✅, Profiles — открыт.
   - **S-1, S-3** — открыты.
   - **S-4** — частично закрыт: Exercises следует варианту B ✅, Auth и Profiles — открыты.
   - **S-5** — открыт.
4. **Документирование:** B-3.

## Когда пересмотреть

- Перед первым продовым деплоем — пройтись по блокаторам.
- При добавлении нового сервиса (Workouts, Habits, ...) — сверить
  с этим списком, чтобы не тащить те же баги.
- После закрытия каждого пункта — обновлять статус, не копить.

## Связанные решения

- [ADR-001: Профиль пользователя](001-profile.md) — lazy-create, nullable поля.
- [ADR-003: Транзакции](003-transactions.md) — sync-вызовы между сервисами.
- [ADR-004: Паттерны микросервисов](004-microservices-patterns.md) — общий tech debt.
- [ADR-006: Управление секретами](006-secrets-management.md) — статус внедрения.
- [ADR-008: Жизненный цикл упражнения](008-exercise-lifecycle.md) — F-3 про идемпотентность DELETE.
- [ADR-011: Naming conventions и Code style](011-naming-and-code-style.md) — соглашения по именованию и стилю.