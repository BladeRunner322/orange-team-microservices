
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
```bash
go install github.com/go-task/task/v3/cmd/task@latest
```
### Установка grpcurl

**Через Go:**
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

**Через Homebrew (macOS):**
```bash
brew install grpcurl
```

**Через snap (Ubuntu):**
```bash
sudo snap install grpcurl
```

**Бинарник с GitHub:**
```bash
https://github.com/fullstorydev/grpcurl/releases
```


## Быстрый старт

1. Склонируйте репозиторий
```bash
git clone https://github.com/BladeRunner322/orange-team-microservices
cd orange-team-microservices
```
2. Настройте переменные окружения

В корне проекта лежит шаблон `.env.example` со всеми переменными. Скопируйте его в `.env` и заполните секреты:
```bash
cp .env.example .env
```

Обязательно укажите:

- `JWT_SECRET` — секретный ключ для JWT (минимум 32 байта)
- `POSTGRES_PASSWORD` — пароль для БД
- `REDIS_PASSWORD` — пароль для Redis (refresh-токены)
- `GRAFANA_PASSWORD` — пароль администратора Grafana

> ⚠️ Файл `.env` добавлен в `.gitignore` и **не коммитится** в репозиторий. Секреты хранятся только локально.

3. Сгенерируйте TLS-сертификаты для разработки

Auth Service использует gRPC с TLS. Для локальной разработки создайте самоподписанный сертификат одним из способов:

**Через Taskfile (рекомендуется):**
```bash
task gen-certs
```

**Вручную через openssl:**
```bash
mkdir -p certs
openssl req -x509 -newkey rsa:4096 -keyout certs/server.key -out certs/server.crt -days 365 -nodes -subj "/CN=localhost"
```

4. Запустите всё окружение (PostgreSQL + Redis + миграции + сервисы)
```bash
task docker-up
```

Это поднимет:

- PostgreSQL (порт 5432)

- Redis (порт 6379) — для refresh-токенов

- Миграции (создание таблиц)

- Auth-сервис (gRPC, порт 50051)

- Gateway-сервис (HTTP, порт 8081)

- Мониторинг (Prometheus, Grafana, Loki, Promtail)


5. Проверьте, что сервисы работают

**Auth (gRPC):**
```bash
grpcurl -insecure localhost:50051 list
```


Ожидаемый ответ Auth:
```bash
auth.AuthService
grpc.reflection.v1.ServerReflection
grpc.reflection.v1alpha.ServerReflection
```

**Gateway (HTTP):**
```bash
curl http://localhost:8081/health
```

Ожидаемый ответ Gateway: `{"service":"gateway","status":"ok"}`.

6. Протестируйте регистрацию и логин

**Через gRPC (напрямую в Auth):**
```bash
grpcurl -insecure -d '{"email":"test@example.com","password":"password123","full_name":"Test User"}' localhost:50051 auth.AuthService/Register
```
```bash
grpcurl -insecure -d '{"email":"test@example.com","password":"password123"}' localhost:50051 auth.AuthService/Login
```

**Через Gateway (HTTP):**
```bash
curl -X POST http://localhost:8081/register -H "Content-Type: application/json" -d '{"email":"test@example.com","password":"password123","full_name":"Test User"}'
```
```bash
curl -X POST http://localhost:8081/login -H "Content-Type: application/json" -d '{"email":"test@example.com","password":"password123"}'
```

7. Протестируйте обновление токена и logout

После логина ты получил `refresh_token`. Используй его для обновления пары токенов.

**Refresh (rotation):**

```bash
curl -X POST http://localhost:8081/refresh -H "Content-Type: application/json" -d '{"refresh_token": "<refresh_token_из_логина>"}'
```

Ожидаемый ответ: новая пара `access_token` + `refresh_token`. Старый refresh становится невалидным.

**Logout:**

```bash
curl -X POST http://localhost:8081/logout -H "Content-Type: application/json" -d '{"refresh_token": "<refresh_token>"}'
```

Ожидаемый ответ: `{"status":"logged out"}`. После этого refresh-токен недействителен.

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
│   ├── postgres/                     # пул соединений pgx, конфиг, адаптеры, ошибки
│   └── redis/                        # клиент Redis (refresh tokens), конфиг
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
│   │   │           ├── handlers/     # хендлеры (auth, user, token, health, proxy)
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
│   │   │   │   ├── ports/            # интерфейсы (Repository, TokenManager, RefreshTokenRepository)
│   │   │   │   └── usecases/         # бизнес-логика (Register, Login, ValidateToken, RefreshToken, Logout)
│   │   │   ├── infrastructure/
│   │   │   │   ├── jwt/              # JWT-менеджер (генерация и валидация токенов)
│   │   │   │   ├── postgres_repo/    # реализация репозитория для PostgreSQL
│   │   │   │   └── redis_repo/       # реализация RefreshTokenRepository на Redis
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
│  • Для публичных эндпоинтов (/register, /login, /refresh, /logout)         │
│    → проксирует в Auth                                                     │
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
│  POST   /refresh                → Auth.RefreshToken                        │
│  POST   /logout                 → Auth.Logout                              │
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
│  • RefreshToken      │ │                  │ │                   │
│  • Logout            │ │                  │ │                   │
│                      │ │                  │ │                   │
│  БД: auth_db + Redis │ │  БД: users_db    │ │  БД: exercises_db │
│  (PG users, Redis    │ │                  │ │                   │
│   refresh tokens)    │ │                  │ │                   │
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
│  🔷 Redis (для refresh-токенов):                                    │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  redis-auth        (порт 6379)  → refresh tokens             │   │
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
| **Redis** | Refresh-токены | TCP | `6379` | `6379` | — | Используется через Docker, проброс на хост |
| **Prometheus** | Метрики | HTTP | `9090` | `9090` | — | Только в Docker |
| **Grafana** | Визуализация | HTTP | `3000` | `3000` | — | Только в Docker |
| **Loki** | Логи | HTTP | `3100` | `3100` | — | Только в Docker |

### Логика смещения портов

- **Docker** — используются стандартные порты:  
  `50051` (Auth gRPC), `8080` (Auth HTTP), `8081` (Gateway HTTP).  
  Redis и PostgreSQL пробрасываются на стандартные порты (`6379`, `5432`) без смещения.

- **Локальная разработка** — порты приложений сдвинуты на **+10**:  
  `50061`, `8090`, `8091` — чтобы не конфликтовать с запущенными Docker-контейнерами.  
  Redis и PostgreSQL для локальной разработки используются из Docker через проброс на `localhost`.

- **Gateway** внутри Docker слушает на `8081` и пробрасывается на хост на `8081` — это сделано намеренно, чтобы не конфликтовать с Auth на `8080`.

## Разработка

### Локальный запуск (без Docker)

Для разработки можно запускать сервисы локально. Порты сдвинуты на **+10** относительно Docker, чтобы не конфликтовать с контейнерами (подробнее — в разделе «Сводная таблица портов»).

**Auth Service:**
```bash
task auth:run
```

- gRPC: `localhost:50061`
- HTTP (health/metrics): `localhost:8090`

**Gateway Service (требует запущенного Auth):**
```bash
task gateway:run
```

- HTTP: `localhost:8091`

**Требования для локального запуска:**
- PostgreSQL запущен через Docker: `task auth:postgres-up`
- Redis запущен через Docker: `task auth:redis-up`
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

```bash
go mod vendor
```

или
 
```bash
task vendor
```

## Сервисы

### Auth (аутентификация)

- Назначение: регистрация, логин, валидация JWT, refresh-токены (с rotation), logout.

- Протокол: gRPC

- gRPC-методы: `Register`, `Login`, `ValidateToken`, `RefreshToken`, `Logout`

- Порт: 50051

- БД: PostgreSQL (схема auth, таблица users) + Redis (refresh-токены с TTL)

- Команды:

  | Команда | Назначение |
  |---------|------------|
  | `task auth:run` | Запуск локально (gRPC `50061`, HTTP `8090`) |
  | `task auth:build` | Сборка Docker-образа |
  | `task auth:rebuild` | Пересборка без кеша (с обновлением vendor) |
  | `task auth:up` | Запуск в Docker Compose |
  | `task auth:restart` | Перезапуск (rebuild + up) |
  | `task auth:logs` | Просмотр логов Auth |

- Команды для PostgreSQL:

  | Команда | Назначение |
  |---------|------------|
  | `task auth:postgres-up` | Запуск PostgreSQL |
  | `task auth:postgres-down` | Остановка PostgreSQL |
  | `task auth:postgres-logs` | Просмотр логов PostgreSQL |
  | `task auth:postgres-psql` | Консоль psql |

- Команды для Redis:

  | Команда | Назначение |
  |---------|------------|
  | `task auth:redis-up` | Запуск Redis |
  | `task auth:redis-down` | Остановка Redis |
  | `task auth:redis-logs` | Просмотр логов Redis |
  | `task auth:redis-cli` | Консоль redis-cli |

#### Healthcheck

Auth-сервис предоставляет HTTP-эндпоинт для проверки состояния:

- **Порт:** `8080` (Docker) / `8090` (локально)
- **Эндпоинт:** `/health`
- **Ответ:** `{"status":"ok","service":"auth"}`

Проверка:

**Docker:**
```bash
curl http://localhost:8080/health
```

**Локально (после `task auth:run`):**
```bash
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

  | Команда | Назначение |
  |---------|------------|
  | `task gateway:run` | Запуск локально (HTTP `8091`, требует запущенного Auth) |
  | `task gateway:build` | Сборка Docker-образа |
  | `task gateway:rebuild` | Пересборка без кеша (с обновлением vendor) |
  | `task gateway:up` | Запуск в Docker Compose |
  | `task gateway:restart` | Перезапуск (rebuild + up) |
  | `task gateway:logs` | Просмотр логов Gateway |

#### Healthcheck

Gateway предоставляет HTTP-эндпоинт для проверки состояния:

- **Порт:** `8081` (Docker) / `8091` (локально)
- **Эндпоинт:** `/health`
- **Ответ:** `{"status":"ok","service":"gateway"}`

Проверка:

**Docker:**
```bash
curl http://localhost:8081/health
```

**Локально (после `task gateway:run`):**
```bash
curl http://localhost:8091/health
```

> 💡 Этот эндпоинт используется Docker healthcheck'ом — если он не отвечает, контейнер помечается как `unhealthy` и может быть перезапущен.

#### Маршруты

| Метод | Путь | Назначение | Требует токен |
|-------|------|------------|----------------|
| POST | `/register` | Регистрация (прокси в Auth) | ❌ |
| POST | `/login` | Логин (прокси в Auth), возвращает access + refresh | ❌ |
| POST | `/refresh` | Обновление пары токенов по refresh (rotation) | ❌ |
| POST | `/logout` | Отзыв refresh-токена | ❌ |
| GET | `/users/me` | Профиль пользователя (пока заглушка) | ✅ |
| PATCH | `/users/me` | Обновление профиля (заглушка) | ✅ |
| DELETE | `/users/me` | Удаление профиля (заглушка) | ✅ |
| GET | `/exercises` | Список упражнений (заглушка) | ✅ |
| GET | `/habits` | Привычки (заглушка) | ✅ |
| GET | `/workouts` | Тренировки (заглушка) | ✅ |
| GET | `/leaderboard/daily` | Лидерборд (заглушка) | ✅ |

> **Защищённые маршруты** требуют заголовок `Authorization: Bearer <access_token>`, который валидируется через Auth Service.  
> **`/refresh` и `/logout`** принимают `refresh_token` в теле запроса (не требуют access-токен, потому что access мог истечь).

**Пример `/refresh`:**

```bash
curl -X POST http://localhost:8081/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "<refresh_token>"}'
```

**Пример `/logout`:**

```bash
curl -X POST http://localhost:8081/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "<refresh_token>"}'
```

### Refresh tokens flow

Auth Service выдаёт **пару токенов** при логине:

- **Access token** (JWT, TTL 15 минут) — используется для защищённых запросов.
- **Refresh token** (случайная строка, TTL 30 дней) — хранится в Redis, используется только для обновления пары.

#### Зачем два токена

Короткий access-токен уменьшает окно атаки: если его украдут, злоумышленник сможет пользоваться им максимум 15 минут. Refresh-токен живёт дольше, но используется редко и всегда **ротируется** (см. ниже).

#### Rotation (ротация)

Каждый раз, когда клиент вызывает `/refresh`, происходит **rotation**:

1. Клиент отправляет старый `refresh_token`.
2. Auth проверяет, что токен есть в Redis.
3. **Старый refresh удаляется** из Redis.
4. Генерируется **новая пара**: новый access + новый refresh.
5. Новый refresh сохраняется в Redis с TTL 30 дней.
6. Клиент получает новую пару.

Если кто-то попробует использовать **старый** refresh повторно (например, злоумышленник, который украл его ранее) — Auth вернёт `401`, потому что токена уже нет в Redis.

#### Logout

`/logout` удаляет refresh-токен из Redis. Access-токен **не отзывается** — он сам истечёт через 15 минут. Это компромисс: мы не храним blacklist для access-токенов (это дорого), а полагаемся на короткий TTL.

#### Когда что использовать

| Сценарий | Что вызывает клиент |
|----------|---------------------|
| Первый вход | `POST /login` |
| Истёк access, нужен новый | `POST /refresh` |
| Пользователь нажал «Выйти» | `POST /logout` |
| Обычный защищённый запрос | Заголовок `Authorization: Bearer <access_token>` |

#### Хранение refresh-токенов в Redis

Ключи в Redis:

```
refresh:<token>          → userID              (TTL = 30 дней)
user:<userID>:tokens     → SET of tokens       (обновляется при каждой ротации)
```

Второй ключ нужен для **revoke all** — когда потребуется разлогинить пользователя из всех сессий (например, при смене пароля). Метод `DeleteAllForUser` уже реализован в репозитории, но пока не вызывается из use case.

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
```bash
curl http://<SERVER_HOST>:8081/health # Gateway
```
```bash
curl http://<SERVER_HOST>:8080/health # Auth
```

Оба должны вернуть `{"status":"ok"}`.

### Обновление `.env` на сервере

Файл `.env` **не хранится в git** (добавлен в `.gitignore`) и **не подтягивается** при `git pull` на сервере. Это значит, что при добавлении новых переменных окружения (например, `REDIS_ADDR`, `REDIS_PASSWORD`, `ACCESS_TOKEN_TTL`, `REFRESH_TOKEN_TTL`) их нужно **обновить вручную** перед деплоем.

Порядок действий при добавлении новых переменных:

1. На **локальной машине** обнови `.env.example` (шаблон) и закоммить в репозиторий.
2. Подключись к серверу и обнови `/root/projects/orange-team-microservices/.env`:
   ```bash
   ssh root@<SERVER_HOST>
   cd /root/projects/orange-team-microservices
   nano .env
   ```
3. Добавь новые переменные, сохрани (`Ctrl+O`, Enter, `Ctrl+X`).
4. После этого — вливай PR в `main`. CD задеплоит сервисы, и они корректно подхватят новые переменные.

> ⚠️ Если забыть обновить `.env` на сервере, Auth/Gateway упадут при старте с ошибкой вида `envconfig: required env var REDIS_ADDR not set`.

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
```bash
curl http://localhost:8080/metrics
```

**Локально:**
```bash
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

```bash
task docker-up        # поднимает всё (приложение + мониторинг + логи)
task monitoring-up    # только Prometheus + Grafana
task logging-up       # только Loki + Promtail
```

### Проверка работы

**Метрики (Docker):**
```bash
curl http://localhost:8080/metrics
```

**Метрики (локально после `task auth:run`):**
```bash
curl http://localhost:8090/metrics
```

**Loki готов:**
```bash
curl http://localhost:3100/ready
```

## Тестирование

### Юнит-тесты
```bash
task test
```

### Интеграционные тесты (с Testcontainers)
```bash
task test-integration
```

### Все тесты (юнит + интеграционные)
```bash
task test-all
```

### Покрытие
```bash
task test-cover
```

Отчёт будет в coverage/coverage.html.

### Postman-коллекция

Для тестирования HTTP API Gateway через Postman подготовлена готовая коллекция:

- **Расположение:** `api/postman/gateway_collection.json`
- **Покрытие:** 47 проверок (34 запроса: health, register, login, refresh, logout, protected endpoints, error cases)
- **Автоматизация:** pre-request скрипт генерирует уникальный email, тест после login сохраняет `access_token` и `refresh_token` в переменные

**Что покрыто:**

- Healthcheck
- Регистрация (успех, дубликат, невалидный email, слабый пароль)
- Логин (успех, неверные учётные данные)
- Обновление токенов через `/refresh` (успех, невалидный refresh)
- Logout и проверка, что refresh после logout не работает
- Защищённые эндпоинты (с токеном, без токена, с невалидным токеном)
- Заглушки для будущих сервисов (Users, Exercises, Habits, Workouts, Leaderboard)

**Как использовать:**

1. Импортируй файл `api/postman/gateway_collection.json` в Postman.
2. Убедись, что переменная `base_url` = `http://localhost:8081` (Docker) или `http://localhost:8091` (локально).
3. Запусти коллекцию через **Run collection** — все тесты должны пройти.

**Переменные коллекции:**

| Переменная | Назначение |
|------------|------------|
| `base_url` | Адрес Gateway |
| `accessToken` | JWT access-токен (заполняется после login и обновляется при refresh) |
| `refreshToken` | Refresh-токен (заполняется после login, обновляется при refresh, удаляется при logout) |
| `userId` | ID пользователя (заполняется после register) |
| `email` | Уникальный email (генерируется pre-request скриптом) |
| `password` | Пароль по умолчанию |
| `fullName` | Полное имя пользователя |

## Управление миграциями

> 💡 При запуске `task docker-up` миграции применяются **автоматически** (сервис `migrate-auth` в `docker-compose.yml`). Ручные команды ниже нужны только для случаев, когда миграции запускаются отдельно (например, локальная разработка или откат).

> ⚠️ Миграции **не создают базу данных** — они только создают таблицы и схему внутри существующей БД. Если база `auth_db` отсутствует, `task auth:migrate-up` упадёт с ошибкой `database "auth_db" does not exist`.

Для создания БД вручную:
```bash
docker exec -it auth-postgres psql -U test -c "CREATE DATABASE auth_db;"
```

Или просто удали volume и подними заново — `task docker-up` создаст БД автоматически из переменной `POSTGRES_DB` в `.env`:
```bash
task docker-down-v
task docker-up
```

### Создать новую миграцию
```bash
task <service-name>:migrate-create -- create_users_table
```

### Применить миграции
```bash
task <service-name>:migrate-up
```

### Откатить последнюю
```bash
task <service-name>:migrate-down -- 1
```

### Показать текущую версию
```bash
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
| `REDIS_PASSWORD` | Пароль для Redis (refresh-токены) |

### Auth Service

| Переменная | Значение по умолчанию | Назначение |
|------------|----------------------|------------|
| `GRPC_PORT` | `:50051` | Порт gRPC-сервера (внутри контейнера) |
| `HTTP_PORT` | `:8080` | Порт HTTP-сервера для /health и /metrics (внутри контейнера) |
| `JWT_ISSUER` | `auth-service` | Издатель токена |
| `JWT_AUDIENCE` | `orange-team` | Аудитория токена |
| `ACCESS_TOKEN_TTL` | `15m` | Время жизни access-токена |
| `REFRESH_TOKEN_TTL` | `720h` | Время жизни refresh-токена (30 дней) |
| `ENABLE_REFLECTION` | `true` | gRPC reflection (для grpcurl). В проде — `false` |
| `POSTGRES_HOST` | `postgres-auth` | Хост PostgreSQL внутри Docker-сети |
| `POSTGRES_PORT` | `5432` | Порт PostgreSQL |
| `POSTGRES_TIMEOUT` | `30s` | Таймаут операций с БД |
| `REDIS_ADDR` | `redis-auth:6379` | Адрес Redis внутри Docker-сети |
| `REDIS_PASSWORD` | — | Пароль Redis (для локали можно простой, для прода — `openssl rand -hex 32`) |
| `REDIS_DB` | `0` | Номер логической БД Redis |
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
