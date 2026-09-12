
# Orange Team Microservices

Монорепозиторий с микросервисами на Go, построенными по чистой архитектуре и DDD.

## Содержание

- [Требования](#требования)
  - [Установка Task](#установка-task)
  - [Установка grpcurl](#установка-grpcurl)
- [Быстрый старт](#быстрый-старт)
- [Структура проекта](#структура-проекта)
- [Полная архитектура микросервисного приложения](#полная-архитектура-микросервисного-приложения)
- [Инфраструктурные компоненты](#инфраструктурные-компоненты)
- [Сводная таблица портов](#сводная-таблица-портов)
  - [Логика смещения портов](#логика-смещения-портов)
- [Разработка](#разработка)
  - [Локальный запуск (без Docker)](#локальный-запуск-без-docker)
  - [Важно: команда `task docker-down-v`](#важно-команда-task-docker-down-v)
  - [Управление зависимостями (vendor)](#управление-зависимостями-vendor)
- [Сервисы](#сервисы)
  - [Auth (аутентификация)](#auth-аутентификация)
  - [Gateway (API Gateway)](#gateway-api-gateway)
- [CI/CD и деплой](#cicd-и-деплой)
- [Мониторинг и логирование](#мониторинг-и-логирование)
- [Тестирование](#тестирование)
- [Управление миграциями](#управление-миграциями)
- [Переменные окружения](#переменные-окружения)

## Требования

- **Go** 1.26.7 или выше
- **Docker** 24.0 или выше
- **Docker Compose** v2.20 или выше (плагин `docker compose`, а не `docker-compose`)
- **Task** (для управления задачами)
- **grpcurl** (для тестирования gRPC)

### Установка Task
```
go install github.com/go-task/task/v3/cmd/task@latest
```
### Установка grpcurl

**Через Go:**
```
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

**Через Homebrew (macOS):**
```
brew install grpcurl
```

**Через snap (Ubuntu):**
```
sudo snap install grpcurl
```

**Бинарник с GitHub:**
```
https://github.com/fullstorydev/grpcurl/releases
```


## Быстрый старт

1. Склонируйте репозиторий
```
git clone https://github.com/BladeRunner322/orange-team-microservices
cd orange-team-microservices
```
2. Настройте переменные окружения

В корне проекта лежит шаблон `.env.example` со всеми переменными. Скопируйте его в `.env` и заполните секреты:
```
cp .env.example .env
```

Обязательно укажите:

- `JWT_SECRET` — секретный ключ для JWT (минимум 32 байта)
- `POSTGRES_PASSWORD` — пароль для БД
- `GRAFANA_PASSWORD` — пароль администратора Grafana

> ⚠️ Файл `.env` добавлен в `.gitignore` и **не коммитится** в репозиторий. Секреты хранятся только локально.

3. Сгенерируйте TLS-сертификаты для разработки

Auth Service использует gRPC с TLS. Для локальной разработки создайте самоподписанный сертификат одним из способов:

**Через Taskfile (рекомендуется):**
```
task gen-certs
```

**Вручную через openssl:**
```
mkdir -p certs
openssl req -x509 -newkey rsa:4096 -keyout certs/server.key -out certs/server.crt -days 365 -nodes -subj "/CN=localhost"
```

4. Запустите всё окружение (PostgreSQL + миграции + сервисы)
```
task docker-up
```

Это поднимет:

- PostgreSQL (порт 5432)

- Миграции (создание таблиц)

- Auth-сервис (gRPC, порт 50051)

- Gateway-сервис (HTTP, порт 8081)

- Мониторинг (Prometheus, Grafana, Loki, Promtail)


5. Проверьте, что сервисы работают

**Auth (gRPC):**
```
grpcurl -insecure localhost:50051 list
```


Ожидаемый ответ Auth:
```
auth.AuthService
grpc.reflection.v1.ServerReflection
grpc.reflection.v1alpha.ServerReflection
```

**Gateway (HTTP):**
```
curl http://localhost:8081/health
```

Ожидаемый ответ Gateway: `{"service":"gateway","status":"ok"}`.

6. Протестируйте регистрацию и логин

**Через gRPC (напрямую в Auth):**
```
grpcurl -insecure -d '{"email":"test@example.com","password":"password123","full_name":"Test User"}' localhost:50051 auth.AuthService/Register
```
```
grpcurl -insecure -d '{"email":"test@example.com","password":"password123"}' localhost:50051 auth.AuthService/Login
```

**Через Gateway (HTTP):**
```
curl -X POST http://localhost:8081/register -H "Content-Type: application/json" -d '{"email":"test@example.com","password":"password123","full_name":"Test User"}'
```
```
curl -X POST http://localhost:8081/login -H "Content-Type: application/json" -d '{"email":"test@example.com","password":"password123"}'
```

## Структура проекта
```
orange-team-microservices/
├── api/                              # gRPC-контракты
│   ├── auth/
│   │   └── auth.proto
│   ├── users/
│   │   └── users.proto
│   ├── exercises/
│   │   └── exercises.proto
│   ├── habits/
│   │   └── habits.proto
│   ├── workouts/
│   │   └── workouts.proto
│   ├── leaderboard/
│   │   └── leaderboard.proto
│   └── postman/                      # коллекция Postman для тестирования HTTP API
│       └── gateway_collection.json
│
├── internal/
│   └── gen/                          # сгенерированный код из proto
│       └── api/
│           ├── auth/                 # auth.pb.go, auth_grpc.pb.go
│           ├── users/                # users.pb.go, users_grpc.pb.go
│           ├── exercises/
│           ├── habits/
│           ├── workouts/
│           └── leaderboard/
│
├── pkg/                              # общие пакеты
│   ├── grpc/
│   │   └── interceptors/             # gRPC-интерсепторы (логирование, метрики, recovery)
│   ├── logger/                       # структурированное логирование (slog)
│   ├── metrics/                      # метрики Prometheus
│   └── postgres/                     # пул соединений pgx, конфиг, адаптеры, ошибки
│
├── services/
│   ├── gateway/                      # ✅ ГОТОВ — API Gateway (HTTP → gRPC прокси)
│   │   ├── cmd/                      # точка входа (main.go)
│   │   ├── config/                   # конфигурация (envconfig)
│   │   ├── internal/
│   │   │   ├── bootstrap/            # сборка зависимостей приложения
│   │   │   ├── application/
│   │   │   │   └── ports/            # интерфейсы клиентов (AuthClientInterface)
│   │   │   ├── infrastructure/
│   │   │   │   └── clients/          # реализация gRPC-клиента Auth
│   │   │   └── interfaces/
│   │   │       └── http/             # HTTP-слой
│   │   │           ├── handlers/     # хендлеры (/register, /login, /users/me, /health)
│   │   │           ├── middleware/   # аутентификация, логирование, request_id
│   │   │           └── httputil/     # утилиты (SendJSON, SendError, GrpcErrorToHTTP)
│   │   └── Dockerfile
│   │
│   ├── auth/                         # ✅ ГОТОВ — Auth Service (регистрация, логин, валидация JWT)
│   │   ├── cmd/                      # точка входа (main.go)
│   │   ├── config/                   # конфигурация (envconfig)
│   │   ├── internal/
│   │   │   ├── bootstrap/            # сборка зависимостей приложения
│   │   │   ├── domain/               # сущности, value objects, ошибки
│   │   │   ├── application/
│   │   │   │   ├── ports/            # интерфейсы (Repository, TokenManager)
│   │   │   │   └── usecases/         # бизнес-логика (Register, Login, ValidateToken)
│   │   │   ├── infrastructure/
│   │   │   │   ├── jwt/              # JWT-менеджер (генерация и валидация токенов)
│   │   │   │   └── postgres_repo/    # реализация репозитория для PostgreSQL
│   │   │   └── interfaces/
│   │   │       ├── authgrpc/         # gRPC-сервер (обработчики AuthService)
│   │   │       └── http/
│   │   │           └── health/       # HTTP-эндпоинты /health и /metrics
│   │   ├── migrations/               # SQL-миграции для auth_db
│   │   └── Dockerfile
│   │
│   ├── users/                        # НОВЫЙ
│   │   ├── cmd/
│   │   ├── config/
│   │   ├── internal/
│   │   │   ├── bootstrap/
│   │   │   ├── domain/
│   │   │   ├── application/
│   │   │   ├── infrastructure/
│   │   │   └── interfaces/
│   │   ├── migrations/               # users_db
│   │   └── Dockerfile
│   │
│   ├── exercises/                    # НОВЫЙ
│   │   └── ... (аналогично)
│   │
│   ├── habits/                       # НОВЫЙ
│   │   └── ... (аналогично)
│   │
│   ├── workouts/                     # НОВЫЙ
│   │   └── ... (аналогично)
│   │
│   └── leaderboard/                  # НОВЫЙ
│       └── ... (аналогично)
│
├── .env.example                      # шаблон переменных окружения
├── .dockerignore                     # исключения для Docker-контекста
├── .gitignore                        # исключения для Git
├── .github/                          # GitHub Actions
│   └── workflows/
│       └── ci-cd.yml                 # CI/CD пайплайн
│
├── certs/                            # TLS-сертификаты (в .gitignore)
│   ├── server.crt
│   └── server.key
│
├── vendor/                           # зафиксированные зависимости (go mod vendor)
│
├── coverage/                         # отчёты о покрытии (в .gitignore)
│   ├── coverage.out
│   └── coverage.html
│
├── logs/                             # файлы логов (в .gitignore)
│
├── docker-compose.yml                # все контейнеры
├── Taskfile.yml                      # задачи для разработки
├── prometheus.yml                    # конфигурация Prometheus
├── promtail-config.yml               # конфигурация Promtail
├── go.mod
├── go.sum
└── README.md
```
## Полная архитектура микросервисного приложения
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          КЛИЕНТЫ (Внешние)                                  │
│                    Браузер / Мобильное приложение                           │
└────────────────────────────────┬────────────────────────────────────────────┘
                                 │ HTTP (REST API)
                                 ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                         API GATEWAY                                        │
│                    (HTTP → gRPC прокси)                                    │
│                                                                            │
│  📋 Функции:                                                               │
│  • Принимает HTTP-запросы от клиентов                                      │
│  • Для публичных эндпоинтов (/register, /login) → проксирует в Auth        │
│  • Для защищённых эндпоинтов:                                              │
│    1. Вызывает Auth.ValidateToken() → получает user_id                     │
│    2. Добавляет user_id в gRPC-метаданные                                  │
│    3. Проксирует запрос в нужный сервис                                    │
│  • Преобразует gRPC-ответы в HTTP-ответы                                   │
│  • Собирает метрики (Prometheus)                                           │
│  • Логирует запросы (Loki)                                                 │
│                                                                            │
│  🔀 Маршруты:                                                              │
│  POST   /register               → Auth.Register                            │
│  POST   /login                  → Auth.Login                               │
│  GET    /users/me               → Users.GetUser                            │
│  PATCH  /users/me               → Users.PatchUser                          │
│  DELETE /users/me               → Users.DeleteUser                         │
│  GET    /exercises              → Exercises.GetExercises                   │
│  POST   /exercises              → Exercises.CreateExercise (admin)         │
│  GET    /habits                 → Habits.GetHabits                         │
│  POST   /habits                 → Habits.CreateHabit                       │
│  POST   /habits/{id}/complete   → Habits.CompleteHabit                     │
│  DELETE /habits/{id}            → Habits.DeleteHabit                       │
│  GET    /workouts               → Workouts.GetWorkouts                     │
│  GET    /workouts/{id}          → Workouts.GetWorkout                      │
│  POST   /workouts               → Workouts.CreateWorkout                   │
│  PATCH  /workouts/{id}          → Workouts.PatchWorkout                    │
│  DELETE /workouts/{id}          → Workouts.DeleteWorkout                   │
│  POST   /workouts/{id}/exercises → Workouts.CreateWorkoutExercise          │
│  GET    /workouts/{id}/exercises → Workouts.GetWorkoutExercises            │
│  PATCH  /workouts/{id}/exercises/{eid} → Workouts.PatchWorkoutExercise     │
│  DELETE /workouts/{id}/exercises/{eid} → Workouts.DeleteWorkoutExercise    │
│  GET    /leaderboard/daily     → Leaderboard.GetDaily                      │
│  GET    /leaderboard/weekly    → Leaderboard.GetWeekly                     │
│  GET    /leaderboard/monthly   → Leaderboard.GetMonthly                    │
└──────────────┬─────────────────┬─────────────────┬─────────────────────────┘
               │                 │                 │
               │ gRPC            │ gRPC            │ gRPC
               ▼                 ▼                 ▼
┌──────────────────────┐ ┌──────────────────┐ ┌───────────────────┐
│   AUTH SERVICE       │ │   USERS SERVICE  │ │  EXERCISES        │
│   (✅ ГОТОВ)         │ │   (НОВЫЙ)        │ │  SERVICE          │
│                      │ │                  │ │  (НОВЫЙ)          │
│  gRPC-методы:        │ │  gRPC-методы:    │ │  gRPC-методы:     │
│  • Register          │ │  • GetUser       │ │  • GetExercises   │
│  • Login             │ │  • PatchUser     │ │  • CreateExercise │
│  • ValidateToken     │ │  • DeleteUser    │ │                   │
│                      │ │                  │ │                   │
│  БД: PostgreSQL      │ │  БД: PostgreSQL  │ │  БД: PostgreSQL   │
│  └── auth_db         │ │  └── users_db    │ │  └── exercises_db │
│      └── users       │ │      └── users   │ │      └── exercises│
│                      │ │                  │ │                   │
│  Порт: :50051        │ │  Порт: :50052    │ │  Порт: :50053     │
└──────────────────────┘ └──────────────────┘ └───────────────────┘
               │                 │                 │
               │ gRPC            │ gRPC            │ gRPC
               ▼                 ▼                 ▼
┌──────────────────────┐ ┌───────────────────┐ ┌─────────────────────┐
│   HABITS SERVICE     │ │  WORKOUTS         │ │  LEADERBOARD        │
│   (НОВЫЙ)            │ │  SERVICE          │ │  SERVICE            │
│                      │ │  (НОВЫЙ)          │ │  (НОВЫЙ)            │
│  gRPC-методы:        │ │                   │ │                     │
│  • GetHabits         │ │  gRPC-методы:     │ │  gRPC-методы:       │
│  • CreateHabit       │ │  • GetWorkouts    │ │  • GetDaily         │
│  • CompleteHabit     │ │  • GetWorkout     │ │  • GetWeekly        │
│  • DeleteHabit       │ │  • CreateWorkout  │ │  • GetMonthly       │
│                      │ │  • PatchWorkout   │ │                     │
│                      │ │  • DeleteWorkout  │ │  БД: PostgreSQL     │
│  БД: PostgreSQL      │ │  • CreateWorkout  │ │  └── leaderboard_db │
│  └── habits_db       │ │    Exercise       │ │      └── snapshots  │
│      └── habits      │ │  • GetWorkout     │ │      └── entries    │
│                      │ │    Exercises      │ │                     │
│                      │ │  • PatchWorkout   │ │                     │
│  Порт: :50054        │ │    Exercise       │ │                     │
│                      │ │  • DeleteWorkout  │ │  Порт: :50056       │
│                      │ │    Exercise       │ │                     │
│                      │ │                   │ │                     │
│                      │ │  БД: PostgreSQL   │ │                     │
│                      │ │  └── workouts_db  │ │                     │
│                      │ │      └── workouts │ │                     │
│                      │ │      └── workout_ │ │                     │
│                      │ │          exercises│ │                     │
│                      │ │                   │ │                     │
│                      │ │  Порт: :50055     │ │                     │
└──────────────────────┘ └───────────────────┘ └─────────────────────┘
```

## Инфраструктурные компоненты
```
┌─────────────────────────────────────────────────────────────────────┐
│                    ИНФРАСТРУКТУРА (Docker Compose)                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  🔷 PostgreSQL Контейнеры:                                          │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  postgres-auth     (порт 5432)  → auth_db                    │   │
│  │  postgres-users    (порт 5433)  → users_db                   │   │
│  │  postgres-exercises (порт 5434) → exercises_db               │   │
│  │  postgres-habits   (порт 5435)  → habits_db                  │   │
│  │  postgres-workouts (порт 5436)  → workouts_db                │   │
│  │  postgres-leader   (порт 5437)  → leaderboard_db             │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  🔷 Мониторинг и логирование:                                       │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  Prometheus (порт 9090)  → сбор метрик                       │   │
│  │  Grafana    (порт 3000)  → визуализация                      │   │
│  │  Loki       (порт 3100)  → хранение логов                    │   │
│  │  Promtail   (порт 9080)  → сбор логов из контейнеров         │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  🔷 Миграции:                                                       │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  migrate-auth     → для auth_db                              │   │
│  │  migrate-users    → для users_db                             │   │
│  │  migrate-exercises → для exercises_db                        │   │
│  │  migrate-habits   → для habits_db                            │   │
│  │  migrate-workouts → для workouts_db                          │   │
│  │  migrate-leader   → для leaderboard_db                       │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

## Сводная таблица портов
| Сервис | Назначение | Протокол | Внутри контейнера (Docker) | Хост (Docker) | Локальная разработка | Примечание |
|--------|------------|----------|----------------------------|---------------|----------------------|------------|
| **Auth** | gRPC | gRPC | `50051` | `50051` | `50061` | Смещение +10 |
| **Auth** | Health / Metrics | HTTP | `8080` | `8080` | `8090` | Смещение +10, переопределяется в `docker-compose.yml` |
| **Gateway** | HTTP API | HTTP | `8081` | `8081` | `8091` | Смещение +10 |
| **PostgreSQL** | База данных | TCP | `5432` | `5432` | — | Используется через Docker, проброс на хост |
| **Prometheus** | Метрики | HTTP | `9090` | `9090` | — | Только в Docker |
| **Grafana** | Визуализация | HTTP | `3000` | `3000` | — | Только в Docker |
| **Loki** | Логи | HTTP | `3100` | `3100` | — | Только в Docker |

### Логика смещения портов

- **Docker** — используются стандартные порты:  
  `50051` (Auth gRPC), `8080` (Auth HTTP), `8081` (Gateway HTTP).

- **Локальная разработка** — все порты сдвинуты на **+10**:  
  `50061`, `8090`, `8091` — чтобы не конфликтовать с запущенными Docker-контейнерами.

- **Gateway** внутри Docker слушает на `8081` и пробрасывается на хост на `8081` — это сделано намеренно, чтобы не конфликтовать с Auth на `8080`.

## Разработка

### Локальный запуск (без Docker)

Для разработки можно запускать сервисы локально. Порты сдвинуты на **+10** относительно Docker, чтобы не конфликтовать с контейнерами (подробнее — в разделе «Сводная таблица портов»).

**Auth Service:**
```
task auth:run
```

- gRPC: `localhost:50061`
- HTTP (health/metrics): `localhost:8090`

**Gateway Service (требует запущенного Auth):**
```
task gateway:run
```

- HTTP: `localhost:8091`

**Требования для локального запуска:**
- PostgreSQL запущен через Docker: `task auth:postgres-up`
- Миграции применены: `task auth:migrate-up`
- TLS-сертификаты созданы: `task gen-certs`

### Важно: команда `task docker-down-v`

Команда `task docker-down-v` **останавливает все контейнеры и удаляет volumes**, включая:

- `pgdata` — **все данные PostgreSQL** (пользователи, тренировки, привычки и т.д.)
- `grafana-storage` — дашборды и настройки Grafana

После её выполнения база данных будет пустой, и миграции придётся применять заново (`task docker-up` сделает это автоматически).

**Когда использовать:**
- Нужно чистое состояние (например, после смены схемы БД в миграциях).
- Отладка проблем с БД.

**Когда НЕ использовать:**
- Если в БД есть важные данные (они будут безвозвратно удалены).
- Если нужно просто перезапустить сервис — используй `task docker-down` (без `-v`) или `task auth:restart`.

### Управление зависимостями (vendor)

Проект использует `vendor` для ускорения сборки Docker. Все зависимости зафиксированы в папке `vendor`.

Если вы добавили новую зависимость в `go.mod`, обновите `vendor`:

```
go mod vendor
```

или
 
```
task vendor
```

## Сервисы

### Auth (аутентификация)

- Назначение: регистрация, логин, валидация JWT.

- Протокол: gRPC

- Порт: 50051

- БД: PostgreSQL (схема auth, таблица users)

- Команды:

  - Запуск локально:
    ```
    task auth:run
    ```
  - Сборка Docker-образа:
    ```
    task auth:build
    ```
  - Пересборка без кеша (с обновлением vendor):
    ```
    task auth:rebuild
    ```
  - Запуск в Docker Compose:
    ```
    task auth:up
    ```
  - Перезапуск (пересборка + запуск):
    ```
    task auth:restart
    ```
  - Просмотр логов:
    ```
    task auth:logs
    ```

#### Healthcheck

Auth-сервис предоставляет HTTP-эндпоинт для проверки состояния:

- **Порт:** `8080` (Docker) / `8090` (локально)
- **Эндпоинт:** `/health`
- **Ответ:** `{"status":"ok","service":"auth"}`

Проверка:

**Docker:**
```
curl http://localhost:8080/health
```

**Локально (после `task auth:run`):**
```
curl http://localhost:8090/health
```

> 💡 Этот эндпоинт используется Docker healthcheck'ом — если он не отвечает, контейнер помечается как `unhealthy` и может быть перезапущен.

### Gateway (API Gateway)

- Назначение: HTTP → gRPC прокси. Единая точка входа для клиентов, централизованная проверка JWT через Auth.

- Протокол: HTTP (REST)

- Порт: 8081

- БД: нет (Gateway не хранит данные)

- Зависимости: требует запущенного Auth Service для валидации токенов.

- Команды:

  - Запуск локально (требует запущенного `task auth:run`):
    ```
    task gateway:run
    ```
  - Сборка Docker-образа:
    ```
    task gateway:build
    ```
  - Пересборка без кеша (с обновлением vendor):
    ```
    task gateway:rebuild
    ```
  - Запуск в Docker Compose:
    ```
    task gateway:up
    ```
  - Перезапуск (пересборка + запуск):
    ```
    task gateway:restart
    ```
  - Просмотр логов:
    ```
    task gateway:logs
    ```

#### Healthcheck

Gateway предоставляет HTTP-эндпоинт для проверки состояния:

- **Порт:** `8081` (Docker) / `8091` (локально)
- **Эндпоинт:** `/health`
- **Ответ:** `{"status":"ok","service":"gateway"}`

Проверка:

**Docker:**
```
curl http://localhost:8081/health
```

**Локально (после `task gateway:run`):**
```
curl http://localhost:8091/health
```

> 💡 Этот эндпоинт используется Docker healthcheck'ом — если он не отвечает, контейнер помечается как `unhealthy` и может быть перезапущен.

#### Маршруты

| Метод | Путь | Назначение |
|-------|------|------------|
| POST | `/register` | Регистрация (прокси в Auth) |
| POST | `/login` | Логин (прокси в Auth) |
| GET | `/users/me` | Профиль пользователя (пока заглушка) |
| PATCH | `/users/me` | Обновление профиля (заглушка) |
| DELETE | `/users/me` | Удаление профиля (заглушка) |
| GET | `/exercises` | Список упражнений (заглушка) |
| GET | `/habits` | Привычки (заглушка) |
| GET | `/workouts` | Тренировки (заглушка) |
| GET | `/leaderboard/daily` | Лидерборд (заглушка) |

> Все защищённые маршруты требуют заголовок `Authorization: Bearer <token>`, который валидируется через Auth Service.

## CI/CD и деплой

Проект использует **GitHub Actions** для автоматической проверки, сборки и деплоя сервисов.

### Триггеры

- **Pull Request в `main`** → запускается только **CI** (тесты, сборка, smoke-тест).
- **Push в `main`** (после вливания PR) → запускается **CD** (сборка образа, публикация в GHCR, деплой на сервер).

### CI (Continuous Integration)

При каждом PR выполняются:

- Установка Go.
- Кеширование модулей.
- `go test -v ./... -short` — юнит-тесты.
- `go build` для `auth` и `gateway` — smoke-тест сборки.

Если CI зелёный — PR готов к слиянию.

### CD (Continuous Deployment)

При пуше в `main`:

1. Определяются изменённые сервисы (`services/auth/**`, `services/gateway/**`, `pkg/**`).
2. Собираются Docker-образы только для изменённых сервисов.
3. Образы публикуются в **GitHub Container Registry**:
   - `ghcr.io/bladerunner322/orange-team-microservices/auth:latest`
   - `ghcr.io/bladerunner322/orange-team-microservices/auth:<git-sha>`
   - `ghcr.io/bladerunner322/orange-team-microservices/gateway:latest`
   - `ghcr.io/bladerunner322/orange-team-microservices/gateway:<git-sha>`
4. По SSH выполняется деплой на продакшен-сервер:
   - Обновление кода (`git pull origin main`).
   - Логин в GHCR.
   - `docker compose pull auth gateway` — скачивание свежих образов.
   - `docker compose up -d auth gateway` — перезапуск контейнеров.

### Секреты GitHub Actions

Для деплоя используются секреты репозитория (Settings → Secrets and variables → Actions):

| Секрет | Назначение |
|--------|------------|
| `SERVER_HOST` | IP или домен продакшен-сервера |
| `SERVER_USER` | SSH-пользователь (обычно `root`) |
| `SERVER_SSH_KEY` | Приватный SSH-ключ для доступа к серверу |

### Проверка после деплоя

После успешного CD проверь на сервере:
```
curl http://<SERVER_HOST>:8081/health # Gateway
```
```
curl http://<SERVER_HOST>:8080/health # Auth
```

Оба должны вернуть `{"status":"ok"}`.

## Мониторинг и логирование

В проекте настроен полный стек для мониторинга и логирования:

- **Prometheus** — сбор метрик
- **Loki** — агрегация логов
- **Grafana** — визуализация

### Метрики (Prometheus)

**Auth-сервис** предоставляет эндпоинт с метриками:

- **Порт:** `8080` (Docker) / `8090` (локально)
- **Эндпоинт:** `/metrics`

Проверка:

**Docker:**
```
curl http://localhost:8080/metrics
```

**Локально:**
```
curl http://localhost:8090/metrics
```

**Gateway** пока не отдаёт метрики через HTTP. В будущем планируется добавить `/metrics` эндпоинт с HTTP-метриками (количество запросов, длительность, статусы).

### Логи (Loki + Promtail)

Логи собираются и хранятся в Loki, доставляются через Promtail.

Просмотр логов в Grafana:

1. Открой `http://localhost:3000`
2. Перейди в **Explore** → выбери **Loki**
3. Введи запрос: `{service="auth"}`

Доступные лейблы для фильтрации:

- `service` — имя сервиса (`auth`)
- `container` — имя контейнера (`auth-service`)
- `level` — уровень логирования (`INFO`, `WARN`, `ERROR`)

### Grafana

Grafana доступна по адресу: `http://localhost:3000`

- Логин: `admin`
- Пароль: `admin` (измените при первом входе)

Источники данных (Prometheus и Loki) подключаются автоматически через provisioning.

Дашборд нужно создать вручную: **+ → Create dashboard** → добавить панели с запросами к Prometheus (метрики) и Loki (логи). Пример запросов:

- `grpc_requests_total` — общее количество gRPC-запросов
- `grpc_request_duration_ms_bucket` — длительность запросов
- `{service="auth"}` — логи Auth в Loki
- `{service="gateway"}` — логи Gateway в Loki

> ⚠️ Дашборды Grafana не сохраняются при `task docker-down-v` (удаление volume `grafana-storage`). Для постоянного хранения настрой **provisioning** (папка `grafana/provisioning/`) или экспортируй дашборд в JSON и положи его в репозиторий.

### Запуск мониторинга

Все команды доступны через Taskfile:

```
task docker-up        # поднимает всё (приложение + мониторинг + логи)
task monitoring-up    # только Prometheus + Grafana
task logging-up       # только Loki + Promtail
```

### Проверка работы

**Метрики (Docker):**
```
curl http://localhost:8080/metrics
```

**Метрики (локально после `task auth:run`):**
```
curl http://localhost:8090/metrics
```

**Loki готов:**
```
curl http://localhost:3100/ready
```

## Тестирование

### Юнит-тесты
```
task test
```

### Интеграционные тесты (с Testcontainers)
```
task test-integration
```

### Покрытие
```
task test-cover
```

Отчёт будет в coverage/coverage.html.

### Postman-коллекция

Для тестирования HTTP API Gateway через Postman подготовлена готовая коллекция:

- **Расположение:** `api/postman/gateway_collection.json`
- **Покрытие:** 30+ тестов (health, register, login, protected endpoints, error cases)
- **Автоматизация:** pre-request скрипт генерирует уникальный email, тест после login сохраняет `access_token` в переменную

**Как использовать:**

1. Импортируй файл `api/postman/gateway_collection.json` в Postman.
2. Убедись, что переменная `base_url` = `http://localhost:8081` (Docker) или `http://localhost:8091` (локально).
3. Запусти коллекцию через **Run collection** — все тесты должны пройти.

**Переменные коллекции:**

| Переменная | Назначение |
|------------|------------|
| `base_url` | Адрес Gateway |
| `accessToken` | JWT-токен (заполняется автоматически после login) |
| `userId` | ID пользователя (заполняется после register) |
| `email` | Уникальный email (генерируется pre-request скриптом) |
| `password` | Пароль по умолчанию |
| `fullName` | Полное имя пользователя |

## Управление миграциями

> 💡 При запуске `task docker-up` миграции применяются **автоматически** (сервис `migrate-auth` в `docker-compose.yml`). Ручные команды ниже нужны только для случаев, когда миграции запускаются отдельно (например, локальная разработка или откат).

> ⚠️ Миграции **не создают базу данных** — они только создают таблицы и схему внутри существующей БД. Если база `auth_db` отсутствует, `task auth:migrate-up` упадёт с ошибкой `database "auth_db" does not exist`.

Для создания БД вручную:
```
docker exec -it auth-postgres psql -U test -c "CREATE DATABASE auth_db;"
```

Или просто удали volume и подними заново — `task docker-up` создаст БД автоматически из переменной `POSTGRES_DB` в `.env`:
```
task docker-down-v
task docker-up
```

### Создать новую миграцию
```
task <service-name>:migrate-create -- create_users_table
```

### Применить миграции
```
task <service-name>:migrate-up
```

### Откатить последнюю
```
task <service-name>:migrate-down -- 1
```

### Показать текущую версию
```
task <service-name>:migrate-version
```

## Переменные окружения

Все переменные хранятся в едином файле `.env` в корне проекта (шаблон — `.env.example`).

### Обязательные

| Переменная | Назначение |
|------------|------------|
| `JWT_SECRET` | Секрет для подписи JWT (минимум 32 байта) |
| `POSTGRES_USER` | Имя пользователя БД |
| `POSTGRES_PASSWORD` | Пароль для PostgreSQL |
| `POSTGRES_DB` | Имя базы данных Auth Service |
| `GRAFANA_PASSWORD` | Пароль администратора Grafana |

### Auth Service

| Переменная | Значение по умолчанию | Назначение |
|------------|----------------------|------------|
| `GRPC_PORT` | `:50051` | Порт gRPC-сервера (внутри контейнера) |
| `HTTP_PORT` | `:8080` | Порт HTTP-сервера для /health и /metrics (внутри контейнера) |
| `JWT_ISSUER` | `auth-service` | Издатель токена |
| `JWT_AUDIENCE` | `orange-team` | Аудитория токена |
| `JWT_EXPIRATION` | `24h` | Время жизни токена |
| `ENABLE_REFLECTION` | `true` | gRPC reflection (для grpcurl). В проде — `false` |
| `POSTGRES_HOST` | `postgres-auth` | Хост PostgreSQL внутри Docker-сети |
| `POSTGRES_PORT` | `5432` | Порт PostgreSQL |
| `POSTGRES_TIMEOUT` | `30s` | Таймаут операций с БД |
| `ENABLE_TLS` | `true` | Использовать TLS для gRPC |
| `TLS_CERT_FILE` | `/app/certs/server.crt` | Путь к сертификату (внутри контейнера) |
| `TLS_KEY_FILE` | `/app/certs/server.key` | Путь к приватному ключу |

### Gateway Service

| Переменная | Значение по умолчанию | Назначение |
|------------|----------------------|------------|
| `GATEWAY_HTTP_PORT` | `:8081` | HTTP-порт Gateway (внутри контейнера) |
| `AUTH_GRPC_ADDR` | `auth-service:50051` | Адрес Auth Service для gRPC-вызовов |
| `GATEWAY_TIMEOUT` | `10s` | Таймаут gRPC-запросов к Auth |

### Общие

| Переменная | Значение по умолчанию | Назначение |
|------------|----------------------|------------|
| `LOGGER_LEVEL` | `DEBUG` | Уровень логирования (`DEBUG`, `INFO`, `WARN`, `ERROR`) |
| `LOGGER_FORMAT` | `json` | Формат логов (`text` или `json`) |
| `LOGGER_FOLDER` | `logs` | Папка для файлов логов |

> ⚠️ Реальный `.env` **не коммитится** в репозиторий (добавлен в `.gitignore`). Для запуска скопируйте `.env.example` в `.env` и заполните секреты.
