# ADR-011: Naming conventions и Code style

## Контекст

Проект развивается: Auth, Gateway, Profiles, Exercises, в плане — Habits, Workouts, Leaderboard. При разработке каждого сервиса принимались локальные решения по именованию и стилю кода. Из-за отсутствия единых правил накопился разнобой:

- В Auth параметры usecase с суффиксом `Str` (`emailStr`), в Exercises — с `Raw` (`nameRaw`).
- Проверка ошибок в usecases — где-то плоский стиль (два `if`), где-то вложенный.
- Логирование VO — где-то `.String()` явно, где-то нет.
- Скан моделей из БД — в Profiles inline, в Exercises метод `Scan`.
- Порядок методов в `ports.Repository` — у каждого сервиса свой.
- Обёртка ошибок `ModelToDomain` — в Profiles без контекста, в Exercises с id.

Без явных правил каждый новый сервис будет плодить новые варианты. Это затрудняет чтение и поддержку.

Этот ADR фиксирует соглашения — **не архитектурные решения, а стиль**. Отвечает на вопросы «как называть» и «как писать».

## Решение

### Naming conventions

**N-1. Мапперы proto ↔ domain.**

- **Хелперы для одного поля:** `<field>FromProto`. Пример: `nameFromProto`, `difficultyFromProto`.
- **Маппер целого объекта domain ← proto:** `ToDomain<Type>`. Пример: `ToDomainPatch`, `ToDomainID`.
- **Маппер целого объекта proto ← domain:** `ToProto<Type>`. Пример: `ToProtoExercise`, `ToProtoProfile`.

**Почему разделяем:** хелпер поля («взять поле из proto») и маппер объекта («преобразовать объект в domain») — разные роли. Разные роли → разные имена.

**Уточнение для универсальных парсеров скаляров.**

Есть функции, которые парсят **одно скалярное значение** из proto (например, `id string` → `uuid.UUID`). Такие функции называют **`ToDomain<Type>`**, а не `<field>FromProto`, если они:

- используются в **нескольких** request-типах (например, `id` в `GetExerciseRequest`, `PatchExerciseRequest`, `DeleteExerciseRequest`);
- принимают **само значение** (не request целиком), чтобы не дублировать функцию под каждый request.

Пример:
```go
func ToDomainID(raw string) (uuid.UUID, error) {
    id, err := uuid.Parse(raw)
    if err != nil {
        return uuid.Nil, fmt.Errorf("parse id: %w", err)
    }
    return id, nil
}
```

Если request с таким полем **один** — допустимо принимать request целиком:
```go
func ToDomainUserID(req *profiles.GetProfileRequest) (uuid.UUID, error)
```

Оба варианта осознанны. Разница: количество потребителей.

**N-2. Методы репозитория — CRUD-alphabetical.**

Порядок:
1. `Create`
2. `Get<Entity>` (одно)
3. `Get<Entities>` (список)
4. `Update`
5. `Delete` / `MarkDeleted` / `Remove`

**Почему так:** соответствует CRUD-модели (создать → прочитать → обновить → удалить), совпадает с алфавитным порядком имён файлов (если методы разбиты по файлам), сохраняется в godoc и IDE-автодополнении.

**N-3. Параметры usecase: суффикс `Raw`.**

Сырые значения, которые ещё не прошли валидацию в VO, получают суффикс `Raw`:

```go
func (uc *CreateExercise) Execute(
    ctx context.Context,
    nameRaw string,
    descriptionRaw string,
    difficultyRaw int,
    exerciseTypeRaw string,
) (domain.Exercise, error) {
    name, err := domain.NewName(nameRaw)
    // ...
}
```

**Почему `Raw`, а не `Str`:** `Str` не подходит для чисел (`difficultyRaw int`). `Raw` — универсален и семантичен: «сырое значение».

**N-4. Геттеры: имя метода = имя типа.**

```go
func (e Exercise) ExerciseType() ExerciseType { return e.exerciseType }
func (p Profile) Sex() *Sex { return p.sex }
func (u User) Email() Email { return u.email }
```

**Почему так:** единообразие. `GetX`-префикс не пишем — в Go это конвенция.

### Code style

**CS-1. Проверка ошибок — вложенный стиль.**

**Стандарт:**
```go
if err != nil {
    if errors.Is(err, domain.ErrX) {
        return ..., domain.ErrX
    }
    return ..., fmt.Errorf("operation: %w", err)
}
```

**Не используем:** плоский стиль (два отдельных `if`).

**Почему вложенный:** явный «ошибка → какая ошибка», классический Go-стиль, не даёт забыть про `err == nil`.

**CS-2. Логирование VO и `uuid.UUID` — без `.String()`.**

`slog` сам вызывает `String()` у типов, реализующих `fmt.Stringer`. Явное `.String()` избыточно.

**Правильно:**
```go
log.Info("exercise fetched", "exercise_id", id, "name", exercise.Name())
```

**Неправильно:**
```go
log.Info("exercise fetched", "exercise_id", id.String(), "name", exercise.Name().String())
```

**CS-3. Метод `Scan(row postgres.Row) error` на модели — стандарт.**

```go
func (m *ExerciseModel) Scan(row postgres.Row) error {
    return row.Scan(&m.ID, &m.Name, &m.Description, ...)
}
```

Применяем везде, где мест скана больше одного. Inline-скан в репозитории (как в Profiles) — устаревший вариант.

**Почему:** единый порядок полей, правка при добавлении нового поля — в одном месте.

**CS-4. Обёртка ошибок `ModelToDomain` с id записи.**

При маппинге модели в домен оборачивать ошибку с id:

```go
exercise, err := ModelToDomain(m)
if err != nil {
    return domain.Exercise{}, fmt.Errorf("exercise %s: %w", m.ID, err)
}
```

**Почему:** в списках (`GetExercises`) важно видеть, **какая** запись не разобралась — иначе непонятно, где в БД искать мусор.

**CS-5. `nil` slice при ошибке, `make([]T, 0)` при успехе.**

```go
func (r *Repository) GetExercises(ctx context.Context) ([]domain.Exercise, error) {
    // ...
    rows, err := r.pool.Query(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("query exercises: %w", err)  // nil при ошибке
    }
    // ...
    exercises := make([]domain.Exercise, 0)  // make при успехе
    // ...
}
```

**Почему:** `nil` slice → JSON `null`, `make([]T, 0)` → JSON `[]`. Клиенты ждут массив, не `null`.

**CS-6. `RETURNING` + `QueryRow` при возврате сущности; `Exec` + `RowsAffected` без.**

| Операция | Возврат | Go |
|---|---|---|
| `Create` (INSERT) | сущность | `QueryRow` + `Scan` + `RETURNING` |
| `Update` (UPDATE) | сущность | `QueryRow` + `Scan` + `RETURNING` |
| `MarkDeleted` (UPDATE) | `error` | `Exec` + `RowsAffected` |
| `Delete` (DELETE) | `error` | `Exec` + `RowsAffected` |

**Почему:** `Exec` не читает `RETURNING`. Если метод возвращает сущность и SQL меняет данные — нужен `RETURNING` для получения актуального состояния из БД (`created_at`, `updated_at`).

**CS-7. `MarkDeleted` в репозитории — soft delete через `deleted_at`.**

```sql
UPDATE <table>
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
```

**Почему:**
- `is_deleted BOOL` хранит только факт, `deleted_at TIMESTAMPTZ` — ещё и время (для аудита, очистки старых записей, восстановления).
- `updated_at` обновляется, потому что сущность изменилась.
- Фильтр `deleted_at IS NULL` — защита от race condition: если параллельно кто-то удалил, `RowsAffected() == 0`.

## Последствия

**Плюсы:**
- Единый стиль во всех сервисах — легче читать, легче переключаться.
- Правила явные — не надо каждый раз решать «а как тут назвать».
- При добавлении нового сервиса — просто следуем конвенциям.

**Минусы:**
- Существующие сервисы (Auth, Profiles) не соответствуют всем правилам — нужен рефакторинг.
- Правила не автоматизированы — держатся на дисциплине и ревью.

**Когда пересмотреть:**
- Если правило окажется неудобным на практике — обсудить и изменить.
- Если появятся инструменты автопроверки (линтеры) — добавить правила в CI.

## Что делать

**Новые сервисы (Exercises, Habits, Workouts, Leaderboard):** следуют этим конвенциям с самого начала.

**Существующие (Auth, Profiles):** привести к конвенциям в отдельном PR — «refactor: apply naming conventions».

**Приоритет рефакторинга:** низкий. Косметика, не блокер.

## Связанные решения

- [ADR-007: Известные проблемы](007-known-issues.md) — список известных багов и tech debt.
- [ADR-009: Карта зависимостей](009-service-dependencies.md) — что есть в проекте.
- [ADR-010: Общие инфраструктурные пакеты](010-shared-infrastructure-packages.md) — что выносим в `pkg/`.