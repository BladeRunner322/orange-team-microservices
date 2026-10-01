
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
  - [Profiles (профили пользователей)](#profiles-профили-пользователей)
  - [Gateway (API Gateway)](#gateway-api-gateway)
  - [Refresh tokens flow](#refresh-tokens-flow)
  - [Rate Limiting flow](#rate-limiting-flow)
  - [RBAC flow](#rbac-flow)
- [CI/CD и деплой](#cicd-и-деплой)
- [Мониторинг и логирование](#мониторинг-и-логирование)
- [Тестирование](#тестирование)
- [Управление миграциями](#управление-миграциями)
- [Переменные окружения](#переменные-окружения)
- [Architecture Decision Records](#architecture-decision-records)

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

В проекте **четыре файла `.env`** — по одному на каждый контекст:

- **Корневой `.env`** — порты для `docker-compose` и Grafana.
- **`services/auth/.env`** — переменные Auth Service.
- **`services/gateway/.env`** — переменные Gateway Service.
- **`services/profiles/.env`** — переменные Profiles Service.

У каждого есть шаблон `.env.example`. Скопируйте все четыре:

```bash
cp .env.example .env
cp services/auth/.env.example services/auth/.env
cp services/gateway/.env.example services/gateway/.env
cp services/profiles/.env.example services/profiles/.env
```

Затем заполните секреты в каждом из них.

**В корневом `.env`:**

- `GRAFANA_PASSWORD` — пароль администратора Grafana (по умолчанию `admin`)

**В `services/auth/.env`:**

- `JWT_SECRET` — секретный ключ для JWT (минимум 32 байта). Сгенерировать: `openssl rand -hex 32`
- `POSTGRES_PASSWORD` — пароль для PostgreSQL
- `REDIS_PASSWORD` — пароль для Redis (refresh-токены)

**В `services/gateway/.env`:**

- `REDIS_PASSWORD` — пароль для Redis (rate limiting, можно тот же или отдельный)

**В `services/profiles/.env`:**

- `POSTGRES_PASSWORD` — пароль для PostgreSQL

> ⚠️ Все четыре `.env` добавлены в `.gitignore` и **не коммитятся** в репозиторий. Секреты хранятся только локально.

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

- PostgreSQL Auth (порт 5432) — для Auth Service
- PostgreSQL Profiles (порт 5433) — для Profiles Service
- Redis Auth (порт 6379) — для refresh-токенов
- Redis Gateway (порт 6380) — для rate limiting
- Миграции Auth и Profiles (создание таблиц)
- Auth-сервис (gRPC, порт 50051)
- Profiles-сервис (gRPC, порт 50052)
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

**Profiles (gRPC):**
```bash
grpcurl -insecure localhost:50052 list
```

Ожидаемый ответ Profiles:
```bash
profiles.ProfilesService
grpc.reflection.v1.ServerReflection
grpc.reflection.v1alpha.ServerReflection
```

**Gateway (HTTP):**
```bash
curl http://localhost:8081/health
```

Ожидаемый ответ Gateway: `{"service":"gateway","status":"ok"}`.

**Profiles (HTTP):**
```bash
curl http://localhost:8082/health
curl http://localhost:8082/ready
```

Ожидаемые ответы: `{"service":"profiles","status":"ok"}` и `{"status":"ok","checks":{"postgres":"ok"}}`.

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
│   ├── profiles/
│   │   └── profiles.proto
│   ├── exercises/
│   │   └── exercises.proto
│   ├── habits/
│   │   └── habits.proto
│   ├── workouts/
│   │   └── workouts.proto
│   ├── leaderboard/
│   │   └── leaderboard.proto
│   └── postman/                      # коллекции Postman
│       ├── gateway_collection.json
│       └── rate_limit_collection.json
│
├── internal/
│   └── gen/                          # сгенерированный код из proto
│       └── api/
│           ├── auth/                 # auth.pb.go, auth_grpc.pb.go
│           ├── profiles/             # profiles.pb.go, profiles_grpc.pb.go
│           ├── exercises/
│           ├── habits/
│           ├── workouts/
│           └── leaderboard/
│
├── pkg/                              # общие пакеты
│   ├── grpc/
│   │   ├── authctx/                  # user_id/role в context + gRPC metadata
│   │   ├── client/                   # конструктор gRPC-клиентов (TLS + interceptor)
│   │   └── interceptors/             # gRPC-интерсепторы (логирование, метрики, recovery, user_id)
│   ├── logger/                       # структурированное логирование (slog)
│   ├── metrics/                      # метрики Prometheus (grpc.go, http.go)
│   ├── nullable/                     # Nullable[T] для patch-полей (Set / Value)
│   ├── postgres/                     # пул соединений pgx, конфиг, адаптеры, ошибки
│   ├── ratelimit/                    # Token Bucket на Redis (rate limiting)
│   └── redis/                        # клиент Redis (refresh tokens, rate limiting)
│
├── services/
│   ├── gateway/                      # ✅ ГОТОВ — API Gateway (HTTP → gRPC прокси)
│   │   ├── cmd/                      # точка входа (main.go)
│   │   ├── config/                   # конфигурация (envconfig)
│   │   ├── internal/
│   │   │   ├── bootstrap/            # сборка зависимостей приложения
│   │   │   ├── application/
│   │   │   │   └── ports/            # интерфейсы клиентов (AuthClientInterface, ProfilesClientInterface)
│   │   │   ├── infrastructure/
│   │   │   │   └── clients/          # реализация gRPC-клиентов Auth и Profiles
│   │   │   └── interfaces/
│   │   │       └── http/             # HTTP-слой
│   │   │           ├── handlers/     # хендлеры (auth, user, token, health, proxy, mapper)
│   │   │           ├── middleware/   # аутентификация, RBAC, rate limit, логирование, request_id
│   │   │           └── httputil/     # утилиты (SendJSON, SendError, GrpcErrorToHTTP)
│   │   ├── .env.example              # шаблон переменных Gateway Service
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
│   │   │           └── health/       # HTTP-эндпоинты /health, /ready, /metrics
│   │   ├── migrations/               # SQL-миграции для auth_db
│   │   ├── .env.example              # шаблон переменных Auth Service
│   │   └── Dockerfile
│   │
│   ├── profiles/                     # ✅ ГОТОВ — Profiles Service (профили пользователей)
│   │   ├── cmd/                      # точка входа (main.go)
│   │   ├── config/                   # конфигурация (envconfig)
│   │   ├── internal/
│   │   │   ├── bootstrap/            # сборка зависимостей приложения
│   │   │   ├── domain/               # Profile, ProfilePatch, VO (sex, weight, height, birth_date)
│   │   │   ├── application/
│   │   │   │   ├── ports/            # интерфейс Repository
│   │   │   │   └── usecases/         # GetMyProfile, GetProfile, PatchMyProfile, DeleteMyProfile
│   │   │   ├── infrastructure/
│   │   │   │   └── postgres_repo/    # реализация Repository для PostgreSQL
│   │   │   └── interfaces/
│   │   │       ├── profilesgrpc/     # gRPC-сервер (обработчики ProfilesService)
│   │   │       └── http/
│   │   │           └── health/       # HTTP-эндпоинты /health, /ready, /metrics
│   │   ├── migrations/               # SQL-миграции для profiles_db
│   │   ├── .env.example              # шаблон переменных Profiles Service
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
├── adr/                              # Architecture Decision Records
│   ├── README.md
│   ├── 001-profile.md
│   ├── 002-user-workout-score.md
│   ├── 003-transactions.md
│   ├── 004-microservices-patterns.md
│   ├── 005-port-allocation.md
│   ├── 006-secrets-management.md
│   └── 007-known-issues.md
│
├── .env.example                      # шаблон корневого .env (порты, Grafana)
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

### Что реализовано сейчас

Актуально на текущий момент работают три сервиса: **Auth**, **Gateway** и **Profiles**. Остальные (Exercises, Habits, Workouts, Leaderboard) — в плане.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          КЛИЕНТЫ (Внешние)                                  │
│                    Браузер / Мобильное приложение                           │
└────────────────────────────────┬────────────────────────────────────────────┘
                                 │ HTTP (REST API)
                                 ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                         API GATEWAY  (✅ ГОТОВ)                            │
│                    (HTTP → gRPC прокси)                                    │
│                                                                            │
│  • Принимает HTTP-запросы от клиентов                                      │
│  • Проверяет JWT через Auth.ValidateToken                                  │
│  • RBAC: читает role из токена, RequireRole("admin") на POST /exercises    │
│  • Rate limiting (Token Bucket на Redis)                                   │
│  • Проксирует публичные запросы в Auth (register/login/refresh/logout)     │
│  • Проксирует /users/me (GET/PATCH/DELETE) в Profiles                      │
│  • Остальные защищённые маршруты бизнес-сервисов — пока заглушки 501       │
│  • Логирует запросы, собирает метрики                                      │
└──────────────┬─────────────────────────────────────────────────────────────┘
               │ gRPC
               ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                       AUTH SERVICE  (✅ ГОТОВ)                             │
│                                                                            │
│  gRPC-методы:                                                              │
│  • Register, Login, ValidateToken, RefreshToken, Logout                    │
│                                                                            │
│  • JWT (HS256) с claims: user_id, role (user|admin)                        │
│  • Refresh-токены в Redis с rotation                                       │
│  • RBAC: роль хранится в auth.users, кладётся в токен при логине/refresh   │
│  • БД: auth_db (PostgreSQL) + Redis (refresh)                              │
│  • Порт: :50051                                                            │
│  • HTTP: /health, /ready, /metrics (:8080)                                 │
└────────────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────────────┐
│                      PROFILES SERVICE  (✅ ГОТОВ)                          │
│                                                                            │
│  gRPC-методы:                                                              │
│  • GetMyProfile, GetProfile, PatchMyProfile, DeleteMyProfile               │
│                                                                            │
│  • Владеет профилем пользователя (sex, weight, height, birth_date)         │
│  • Lazy-create: пустая запись создаётся при первом чтении (см. ADR-001)    │
│  • profile_completed = все 4 поля заполнены                                │
│  • user_id из gRPC metadata (ставит Gateway) или из тела запроса           │
│    для GetProfile (внутренний вызов других сервисов)                       │
│  • БД: profiles_db (PostgreSQL)                                            │
│  • Порт: :50052                                                            │
│  • HTTP: /health, /ready, /metrics (:8080)                                 │
└────────────────────────────────────────────────────────────────────────────┘
```

### План (ещё не реализовано)

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
│  GET    /users/me               → Profiles.GetMyProfile                    │
│  PATCH  /users/me               → Profiles.PatchMyProfile                  │
│  DELETE /users/me               → Profiles.DeleteMyProfile                 │
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
┌──────────────────────┐ ┌──────────────────────┐ ┌──────────────────────┐
│   AUTH SERVICE       │ │  PROFILES SERVICE    │ │  EXERCISES SERVICE   │
│   (✅ ГОТОВ)         │ │   (✅ ГОТОВ)         │ │  (НОВЫЙ)             │
│                      │ │                      │ │                      │
│  gRPC-методы:        │ │  gRPC-методы:        │ │  gRPC-методы:        │
│  • Register          │ │  • GetMyProfile      │ │  • GetExercises      │
│  • Login             │ │  • GetProfile        │ │  • CreateExercise    │
│  • ValidateToken     │ │  • PatchMyProfile    │ │                      │
│  • RefreshToken      │ │  • DeleteMyProfile   │ │                      │
│  • Logout            │ │                      │ │                      │
│                      │ │                      │ │                      │
│  БД: auth_db + Redis │ │  БД: profiles_db     │ │  БД: exercises_db    │
│  (PG users, Redis    │ │                      │ │                      │
│   refresh tokens)    │ │                      │ │                      │
│                      │ │                      │ │                      │
│  Порт: :50051        │ │  Порт: :50052        │ │  Порт: :50053        │
└──────────────────────┘ └──────────────────────┘ └──────────────────────┘
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

### Что реализовано сейчас

- **PostgreSQL:** `postgres-auth` (порт 5432) → `auth_db`; `postgres-profiles` (порт 5433) → `profiles_db`
- **Redis:** `redis-auth` (порт 6379) — refresh-токены; `redis-gateway` (порт 6380) — rate limiting
- **Миграции:** `migrate-auth` → для `auth_db`; `migrate-profiles` → для `profiles_db`
- **Мониторинг:** Prometheus (9090), Grafana (3000)
- **Логи:** Loki (3100), Promtail (9080)

### План (после реализации остальных сервисов)

```
┌─────────────────────────────────────────────────────────────────────┐
│                    ИНФРАСТРУКТУРА (Docker Compose)                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  🔷 PostgreSQL Контейнеры:                                          │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  postgres-auth     (порт 5432)  → auth_db                    │   │
│  │  postgres-profiles (порт 5433)  → profiles_db                │   │
│  │  postgres-exercises (порт 5434) → exercises_db               │   │
│  │  postgres-habits   (порт 5435)  → habits_db                  │   │
│  │  postgres-workouts (порт 5436)  → workouts_db                │   │
│  │  postgres-leader   (порт 5437)  → leaderboard_db             │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  🔷 Redis Контейнеры:                                               │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  redis-auth        (порт 6379)  → refresh-токены (Auth)      │   │
│  │  redis-gateway     (порт 6380)  → rate limiting (Gateway)    │   │
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
│  │  migrate-profiles → для profiles_db                          │   │
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
| **Profiles** | gRPC | gRPC | `50052` | `50052` | `50062` | Смещение +10 |
| **Profiles** | Health / Ready / Metrics | HTTP | `8080` | `8082` | `8092` | Смещение +10, переопределяется в `docker-compose.yml` |
| **PostgreSQL (Auth)** | База данных | TCP | `5432` | `5432` | — | Используется через Docker, проброс на хост |
| **PostgreSQL (Profiles)** | База данных | TCP | `5432` | `5433` | — | Используется через Docker, проброс на хост |
| **Redis Auth** | Refresh-токены | TCP | `6379` | `6379` | — | Только для Auth Service |
| **Redis Gateway** | Rate limiting | TCP | `6379` | `6380` | — | Только для Gateway Service |
| **Prometheus** | Метрики | HTTP | `9090` | `9090` | — | Только в Docker |
| **Grafana** | Визуализация | HTTP | `3000` | `3000` | — | Только в Docker |
| **Loki** | Логи | HTTP | `3100` | `3100` | — | Только в Docker |

### Логика смещения портов

- **Docker** — используются стандартные порты:  
  `50051` (Auth gRPC), `8080` (Auth HTTP), `8081` (Gateway HTTP).  
  PostgreSQL пробрасывается на стандартный порт `5432`.  
  Redis: `redis-auth` → `6379`, `redis-gateway` → `6380` (разные порты на хосте, чтобы не конфликтовать).

- **Локальная разработка** — порты приложений сдвинуты на **+10**:  
  `50061`, `8090`, `8091` — чтобы не конфликтовать с запущенными Docker-контейнерами.  
  PostgreSQL и оба Redis для локальной разработки используются из Docker через проброс на `localhost` (`5432`, `6379`, `6380`).

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

**Gateway Service (требует запущенных Auth и Profiles):**
```bash
task gateway:run
```

- HTTP: `localhost:8091`

**Profiles Service:**
```bash
task profiles:run
```

- gRPC: `localhost:50062`
- HTTP (health/ready/metrics): `localhost:8092`

**Требования для локального запуска:**
- PostgreSQL запущен через Docker: `task auth:postgres-up`
- Redis запущен через Docker: `task auth:redis-up`
- Миграции применены: `task auth:migrate-up`
- PostgreSQL для Profiles запущен: `task profiles:postgres-up`
- Миграции Profiles применены: `task profiles:migrate-up`
- TLS-сертификаты созданы: `task gen-certs`

### Важно: команда `task docker-down-v`

Команда `task docker-down-v` **останавливает все контейнеры и удаляет volumes**, включая:

- `pgdata-auth`, `pgdata-profiles` — все данные PostgreSQL (пользователи, профили)
- `redisdata-auth`, `redisdata-gateway` — данные Redis (refresh-токены, rate-limit buckets)
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

- JWT (HS256) с claims: `user_id`, `role` (`user` | `admin`)

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

#### Healthcheck и readiness

Auth-сервис предоставляет HTTP-эндпоинты для проверки состояния:

| Эндпоинт | Назначение | Ответ |
|----------|------------|-------|
| `/health` | Liveness — процесс жив | `{"status":"ok","service":"auth"}` |
| `/ready` | Readiness — PostgreSQL и Redis доступны | `{"status":"ok","checks":{"postgres":"ok","redis":"ok"}}` или `503` |
| `/metrics` | Prometheus-метрики | text/plain |

**Порт:** `8080` (Docker) / `8090` (локально)

Проверка:

**Docker:**
```bash
curl http://localhost:8080/health
curl http://localhost:8080/ready
curl http://localhost:8080/metrics
```

**Локально (после `task auth:run`):**
```bash
curl http://localhost:8090/health
curl http://localhost:8090/ready
curl http://localhost:8090/metrics
```

> 💡 `/ready` используется Docker healthcheck'ом (в compose проверяется `/ready`). `/health` — liveness. `/metrics` скрейпится Prometheus.

### Profiles (профили пользователей)

- Назначение: хранение и редактирование профиля пользователя (пол, вес, рост, дата рождения).

- Протокол: gRPC

- gRPC-методы: `GetMyProfile`, `GetProfile`, `PatchMyProfile`, `DeleteMyProfile`

- Порт: 50052

- БД: PostgreSQL (схема profiles, таблица users)

- `user_id`:
  - для `GetMyProfile` / `PatchMyProfile` / `DeleteMyProfile` — из gRPC metadata (ставит Gateway)
  - для `GetProfile` — из тела запроса (внутренний вызов других сервисов)

- **Lazy-create**: пустая запись создаётся при первом чтении профиля (см. [ADR-001](adr/001-profile.md)).

- `profile_completed = true`, если заполнены все 4 поля: `sex`, `weight_kg`, `birth_date`, `height_cm`.

- Команды:

  | Команда | Назначение |
  |---------|------------|
  | `task profiles:run` | Запуск локально (gRPC `50062`, HTTP `8092`) |
  | `task profiles:build` | Сборка Docker-образа |
  | `task profiles:rebuild` | Пересборка без кеша (с обновлением vendor) |
  | `task profiles:up` | Запуск в Docker Compose |
  | `task profiles:restart` | Перезапуск (rebuild + up) |
  | `task profiles:logs` | Просмотр логов Profiles |

- Команды для PostgreSQL:

  | Команда | Назначение |
  |---------|------------|
  | `task profiles:postgres-up` | Запуск PostgreSQL |
  | `task profiles:postgres-down` | Остановка PostgreSQL |
  | `task profiles:postgres-logs` | Просмотр логов PostgreSQL |
  | `task profiles:postgres-psql` | Консоль psql |

- Команды для миграций:

  | Команда | Назначение |
  |---------|------------|
  | `task profiles:migrate-create -- <name>` | Создать новую миграцию |
  | `task profiles:migrate-up` | Применить миграции |
  | `task profiles:migrate-down -- 1` | Откатить последнюю |
  | `task profiles:migrate-version` | Показать текущую версию |

#### Healthcheck и readiness

Profiles предоставляет HTTP-эндпоинты для проверки состояния:

| Эндпоинт | Назначение | Ответ |
|----------|------------|-------|
| `/health` | Liveness — процесс жив | `{"status":"ok","service":"profiles"}` |
| `/ready` | Readiness — PostgreSQL доступен | `{"status":"ok","checks":{"postgres":"ok"}}` или `503` |
| `/metrics` | Prometheus-метрики | text/plain |

**Порт:** `8080` (Docker) / `8092` (локально)

Проверка:

```bash
curl http://localhost:8082/health
curl http://localhost:8082/ready
curl http://localhost:8082/metrics
```

> 💡 `/health` и `/ready` используются Docker healthcheck'ом (в compose проверяется `/ready`). `/metrics` скрейпится Prometheus.

### Gateway (API Gateway)

- Назначение: HTTP → gRPC прокси. Единая точка входа для клиентов, централизованная проверка JWT через Auth Service и rate limiting.

- Протокол: HTTP (REST)

- Порт: 8081

- БД: нет (Gateway не хранит данные)

- Зависимости: требует запущенного Auth Service для валидации токенов и Redis (свой инстанс) для rate limiting.

- Rate limiting: Token Bucket на Redis. Лимиты:
  - `/login` — 5 попыток в минуту
  - `/register` — 3 попытки в минуту
  - `/refresh` — 20 попыток в минуту
  - защищённые маршруты — 60 запросов в минуту
  
  При превышении — `429 Too Many Requests` с заголовком `Retry-After: <секунды>`.

- Команды:

  | Команда | Назначение |
  |---------|------------|
  | `task gateway:run` | Запуск локально (HTTP `8091`, требует запущенного Auth) |
  | `task gateway:build` | Сборка Docker-образа |
  | `task gateway:rebuild` | Пересборка без кеша (с обновлением vendor) |
  | `task gateway:up` | Запуск в Docker Compose |
  | `task gateway:restart` | Перезапуск (rebuild + up) |
  | `task gateway:logs` | Просмотр логов Gateway |

#### Healthcheck и readiness

Gateway предоставляет три служебных HTTP-эндпоинта:

| Эндпоинт | Назначение | Ответ |
|----------|------------|-------|
| `/health` | Liveness — процесс жив | `{"status":"ok","service":"gateway"}` |
| `/ready` | Readiness — зависимости доступны (Redis) | `{"status":"ok","checks":{"redis":"ok"}}` или `503` |
| `/metrics` | Prometheus-метрики | text/plain |

**Проверка:**

```bash
curl http://localhost:8081/health
curl http://localhost:8081/ready
curl http://localhost:8081/metrics
```

Если Redis недоступен, `/ready` вернёт `503` с описанием:

```json
{"status":"not_ready","checks":{"redis":"fail: dial tcp ..."}}
```

> 💡 `/health` используется Docker healthcheck'ом. `/ready` — для внешнего балансировщика (не направлять трафик в инстанс, пока он не готов).

#### Маршруты

| Метод | Путь | Назначение | Требует токен | Роль |
|-------|------|------------|----------------|------|
| POST | `/register` | Регистрация | ❌ | — |
| POST | `/login` | Логин (access + refresh) | ❌ | — |
| POST | `/refresh` | Обновление пары токенов (rotation) | ❌ | — |
| POST | `/logout` | Отзыв refresh-токена | ❌ | — |
| GET | `/users/me` | Профиль пользователя → Profiles.GetMyProfile | ✅ | — |
| PATCH | `/users/me` | Обновление профиля → Profiles.PatchMyProfile | ✅ | — |
| DELETE | `/users/me` | Удаление профиля → Profiles.DeleteMyProfile | ✅ | — |
| GET | `/exercises` | Список упражнений (заглушка) | ✅ | — |
| POST | `/exercises` | Создать упражнение (заглушка) | ✅ | **admin** |
| GET | `/habits` | Список привычек (заглушка) | ✅ | — |
| POST | `/habits` | Создать привычку (заглушка) | ✅ | — |
| POST | `/habits/{habitId}/complete` | Отметить выполнение (заглушка) | ✅ | — |
| DELETE | `/habits/{habitId}` | Удалить привычку (заглушка) | ✅ | — |
| GET | `/workouts` | Список тренировок (заглушка) | ✅ | — |
| POST | `/workouts` | Создать тренировку (заглушка) | ✅ | — |
| GET | `/workouts/{workoutId}` | Тренировка по ID (заглушка) | ✅ | — |
| PATCH | `/workouts/{workoutId}` | Обновить тренировку (заглушка) | ✅ | — |
| DELETE | `/workouts/{workoutId}` | Удалить тренировку (заглушка) | ✅ | — |
| POST | `/workouts/{workoutId}/exercises` | Добавить упражнение в тренировку (заглушка) | ✅ | — |
| GET | `/workouts/{workoutId}/exercises` | Упражнения тренировки (заглушка) | ✅ | — |
| PATCH | `/workouts/{workoutId}/exercises/{exerciseId}` | Обновить упражнение (заглушка) | ✅ | — |
| DELETE | `/workouts/{workoutId}/exercises/{exerciseId}` | Удалить упражнение (заглушка) | ✅ | — |
| GET | `/leaderboard/daily` | Дневной лидерборд (заглушка) | ✅ | — |
| GET | `/leaderboard/weekly` | Недельный лидерборд (заглушка) | ✅ | — |
| GET | `/leaderboard/monthly` | Месячный лидерборд (заглушка) | ✅ | — |

> **Защищённые маршруты** требуют заголовок `Authorization: Bearer <access_token>`, который валидируется через Auth Service.  
> **`/refresh` и `/logout`** принимают `refresh_token` в теле запроса (не требуют access-токен, потому что access мог истечь).  
> **`/users/me`** проксируется в Profiles Service. `user_id` берётся из JWT и передаётся через gRPC metadata (`x-user-id`).

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

### Rate Limiting flow

Gateway защищает публичные эндпоинты от брутфорса и абуза с помощью **rate limiting**. Реализация — **Token Bucket на Redis** с атомарным Lua-скриптом.

#### Что такое Token Bucket

Абстрактное «ведро» с токенами:

- У каждого клиента есть **своё ведро** (определяется ключом).
- В ведре максимум `Burst` токенов.
- За период `Interval` восстанавливается `Rate` токенов (плавно).
- Каждый запрос **списывает 1 токен**.
- Если токенов нет — запрос отклоняется с `429`.

#### Лимиты по эндпоинтам

| Эндпоинт | Rate | Burst | Interval |
|----------|------|-------|----------|
| `POST /login` | 5 | 5 | 1 минута |
| `POST /register` | 3 | 3 | 1 минута |
| `POST /refresh` | 20 | 20 | 1 минута |
| Защищённые маршруты | 60 | 60 | 1 минута |

#### Ключи для лимитов

Чтобы лимит был «умным», ключ формируется по-разному для разных сценариев:

| Эндпоинт | Ключ в Redis | Почему |
|----------|--------------|--------|
| `POST /login` | `rl:login:<ip>:<email>` | Защищаем конкретный аккаунт от брутфорса с конкретного IP |
| `POST /register` | `rl:register:<ip>:<email>` | Не даём спамить регистрациями с одного IP |
| `POST /refresh` | `rl:refresh:<ip>` | Ограничиваем по IP |
| Защищённые | `rl:api:<user_id>` | Ограничиваем конкретного пользователя |

**Email** извлекается из тела запроса — middleware читает body и **восстанавливает его** обратно, чтобы handler тоже мог его прочитать.

#### Что возвращаем при превышении

```
HTTP/1.1 429 Too Many Requests
Retry-After: 11
Content-Type: application/json

{"error":"rate limit exceeded"}
```

`Retry-After` — стандартный HTTP-заголовок: через сколько секунд восстановится 1 токен.

#### Почему Lua, а не просто Redis-команды

Операция «прочитать токены → пополнить → списать 1 → записать обратно» должна быть **атомарной**. Если делать её через несколько команд `GET`/`SET` — при параллельных запросах возникнет **race condition** (двое одновременно прочитают `tokens = 1` и оба спишут).

Lua-скрипт в Redis выполняется **как единое целое** — Redis не отвлекается на другие команды, пока скрипт не закончится. Это гарантирует корректность при любом количестве инстансов Gateway.

#### Хранение в Redis

Ключи в Redis (`redis-gateway`, отдельный от `redis-auth`):

```
rl:login:<ip>:<email>      → { tokens, last_refill }   (TTL ≈ Interval + 1s)
rl:register:<ip>:<email>   → { tokens, last_refill }
rl:refresh:<ip>            → { tokens, last_refill }
rl:api:<user_id>           → { tokens, last_refill }
```

TTL устанавливается автоматически: если клиент больше не приходит — ключ удаляется, Redis не забивается.

#### Fail-open

Если Redis **недоступен** — middleware **пропускает запрос** и логирует ошибку. Это сделано намеренно: лучше пропустить запрос, чем уронить весь сервис, если Redis упадёт. Такой подход называется **fail-open**.

### RBAC flow

Auth кладёт роль пользователя в JWT при логине и refresh. Gateway читает роль из уже провалидированного токена и применяет middleware `RequireRole`.

#### Что такое роль

- Хранится в `auth.users.role` — значения `user` (по умолчанию) и `admin`.
- Валидируется CHECK-констрейнтом на уровне БД.
- Кладётся в JWT claim `role` (помимо `user_id`).

#### Почему роль в токене, а не в БД

- Gateway не ходит в БД на каждый запрос — быстрее.
- Auth не зависит от других сервисов при логине — устойчивее.
- Любой downstream-сервис может локально проверить роль по JWT.
- Это стандартный подход в проде (Auth0, Keycloak, Google IAP).

#### Проверка прав

Gateway применяет middleware `RequireRole("admin")` на админских маршрутах.

| Маршрут | Требует роль |
|---------|--------------|
| `POST /exercises` | `admin` |

Если роль не подходит — Gateway возвращает `403 Forbidden` **без** проксирования вниз.

Остальные защищённые маршруты требуют просто валидный JWT — роль любая.

#### Смена роли

Роль применяется:

1. **Сразу** при следующем логине.
2. **При следующем refresh** — `RefreshToken.Execute` перечитывает пользователя из БД и берёт актуальную роль.
3. **Максимум через 15 минут** — TTL access-токена (после этого клиент всё равно пойдёт за новым).

Это осознанный компромисс: не ходим в БД на каждый запрос, но роль не «застревает» надолго.

#### Пример

```bash
# 1. Логин — получаем access-токен с role=user
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# 2. Пробуем создать упражнение (admin-only)
curl -X POST http://localhost:8081/exercises \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{}'
# → 403 Forbidden {"error":"forbidden"}

# 3. Меняем роль в БД (вручную, для теста)
# docker compose exec postgres-auth psql -U test -d auth_db
#   UPDATE auth.users SET role='admin' WHERE email='user@example.com';

# 4. Перелогиниваемся — получаем новый токен с role=admin
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# 5. Повторяем запрос
curl -X POST http://localhost:8081/exercises \
  -H "Authorization: Bearer <новый_access_token>" \
  -H "Content-Type: application/json" \
  -d '{}'
# → 501 Not Implemented (RBAC пропустил, дошло до заглушки)
```

#### Что дальше

- В проде подпись JWT стоит заменить с HS256 на **RS256** + JWKS, чтобы не раздавать общий секрет по сервисам.
- Роль `admin` назначается вручную через БД (отдельной ручки смены роли пока нет).
- Когда появится больше ролей — расширить `RequireRole("admin", "moderator", ...)`.

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
- `go build` для `auth`, `gateway` и `profiles` — smoke-тест сборки.

Если CI зелёный — PR готов к слиянию.

### CD (Continuous Deployment)

При пуше в `main`:

1. Определяются изменённые сервисы (`services/auth/**`, `services/gateway/**`, `services/profiles/**`, `pkg/**`).
2. Собираются Docker-образы только для изменённых сервисов.
3. Образы публикуются в **GitHub Container Registry**:
   - `ghcr.io/bladerunner322/orange-team-microservices/auth:latest`
   - `ghcr.io/bladerunner322/orange-team-microservices/auth:<git-sha>`
   - `ghcr.io/bladerunner322/orange-team-microservices/gateway:latest`
   - `ghcr.io/bladerunner322/orange-team-microservices/gateway:<git-sha>`
   - `ghcr.io/bladerunner322/orange-team-microservices/profiles:latest`
   - `ghcr.io/bladerunner322/orange-team-microservices/profiles:<git-sha>`
4. По SSH выполняется деплой на продакшен-сервер:
   - Обновление кода (`git pull origin main`).
   - Логин в GHCR.
   - `docker compose pull auth gateway profiles` — скачивание свежих образов.
   - `docker compose up -d --no-build auth gateway profiles` — перезапуск контейнеров из скачанных образов (без локальной сборки).

### Секреты GitHub Actions

Для деплоя используются секреты репозитория (Settings → Secrets and variables → Actions):

| Секрет | Назначение |
|--------|------------|
| `SERVER_HOST` | IP или домен продакшен-сервера |
| `SERVER_USER` | SSH-пользователь (обычно `root`) |
| `SERVER_SSH_KEY` | Приватный SSH-ключ для доступа к серверу |

### Проверка после деплоя

После успешного CD проверь на сервере:

**Health:**

```bash
curl http://<SERVER_HOST>:8081/health   # Gateway
curl http://<SERVER_HOST>:8080/health   # Auth
curl http://<SERVER_HOST>:8082/health   # Profiles
```

Все три должны вернуть `{"status":"ok","service":"..."}`.

**Ready:**

```bash
curl http://<SERVER_HOST>:8081/ready    # Gateway → проверка Redis
curl http://<SERVER_HOST>:8080/ready    # Auth → проверка Postgres + Redis
curl http://<SERVER_HOST>:8082/ready    # Profiles → проверка Postgres
```

Должны вернуть `{"status":"ok","checks":{...}}`.

**Prometheus targets:**

Открой `http://<SERVER_HOST>:9090` → **Status → Targets**. Все три job'а должны быть в статусе **UP**:

- `auth` — target `auth:8080`
- `gateway` — target `gateway:8081`
- `profiles` — target `profiles:8080`

### Обновление `.env` на сервере

В проекте **четыре файла `.env`**, и **ни один из них не хранится в git** (все добавлены в `.gitignore`). Они **не подтягиваются** при `git pull` на сервере. Это значит, что при добавлении новых переменных окружения (например, `REDIS_ADDR`, `REDIS_PASSWORD`, `RATE_LIMIT_ENABLED`, `PROFILES_GRPC_ADDR`) их нужно **обновить вручную** в соответствующих файлах на сервере перед деплоем.

Структура `.env` на сервере:

```
/root/projects/orange-team-microservices/
├── .env                          # корневой (порты, Grafana)
└── services/
    ├── auth/.env                 # Auth Service
    ├── gateway/.env              # Gateway Service
    └── profiles/.env             # Profiles Service
```

Порядок действий при добавлении новых переменных:

1. На **локальной машине** обнови соответствующий `.env.example` (шаблон) и закоммить в репозиторий.
2. Подключись к серверу:
   ```bash
   ssh root@<SERVER_HOST>
   cd /root/projects/orange-team-microservices
   ```
3. Обнови нужные `.env` файлы через `nano`:
   ```bash
   nano .env                          # если меняются порты или Grafana
   nano services/auth/.env            # если меняются переменные Auth
   nano services/gateway/.env         # если меняются переменные Gateway
   nano services/profiles/.env        # если меняются переменные Profiles
   ```
4. Сохрани (`Ctrl+O`, Enter, `Ctrl+X`).
5. После этого — вливай PR в `main`. CD задеплоит сервисы, и они корректно подхватят новые переменные.

> ⚠️ Если забыть обновить хотя бы один из `.env` на сервере, соответствующий сервис упадёт при старте с ошибкой вида `envconfig: required env var REDIS_ADDR not set`.

## Мониторинг и логирование

В проекте настроен полный стек для мониторинга и логирования:

- **Prometheus** — сбор метрик
- **Loki** — агрегация логов
- **Grafana** — визуализация

### Метрики (Prometheus)

Все три сервиса отдают метрики в формате Prometheus.

**Auth-сервис:**

- **Порт:** `8080` (Docker) / `8090` (локально)
- **Эндпоинт:** `/metrics`

Метрики:
- `grpc_requests_total` — количество gRPC-запросов (method, status)
- `grpc_request_duration_ms` — длительность gRPC-запросов в мс
- `grpc_requests_in_flight` — gRPC-запросы в обработке

**Gateway:**

- **Порт:** `8081` (Docker) / `8091` (локально)
- **Эндпоинт:** `/metrics`

Метрики:
- `http_requests_total` — количество HTTP-запросов (method, path, status)
- `http_request_duration_seconds` — длительность HTTP-запросов в секундах
- `http_requests_in_flight` — HTTP-запросы в обработке

Плюс транзитивно подтягиваются gRPC-метрики исходящих вызовов в Auth и Profiles.

**Profiles:**

- **Порт:** `8080` (Docker) / `8092` (локально)
- **Эндпоинт:** `/metrics`

Метрики:
- `grpc_requests_total` — количество gRPC-запросов (method, status)
- `grpc_request_duration_ms` — длительность gRPC-запросов в мс
- `grpc_requests_in_flight` — gRPC-запросы в обработке

**Путь нормализуется** через chi RoutePattern, чтобы `/workouts/123` и `/workouts/456` не создавали отдельные серии (защита от взрыва кардинальности).

**Проверка:**

```bash
curl http://localhost:8080/metrics   # Auth
curl http://localhost:8081/metrics   # Gateway
curl http://localhost:8082/metrics   # Profiles
```

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
- `grpc_request_duration_ms_bucket` — длительность gRPC-запросов
- `http_requests_total` — общее количество HTTP-запросов Gateway
- `http_request_duration_seconds_bucket` — длительность HTTP-запросов
- `rate(http_requests_total[1m])` — RPS по эндпоинтам
- `{service="auth"}` — логи Auth в Loki
- `{service="gateway"}` — логи Gateway в Loki
- `{service="profiles"}` — логи Profiles в Loki

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
- Заглушки для будущих сервисов (Exercises, Habits, Workouts, Leaderboard)

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

#### Rate Limiting коллекция

Отдельная коллекция для проверки rate limiting:

- **Расположение:** `api/postman/rate_limit_collection.json`
- **Покрытие:** 9 проверок (5 успешных попыток, 6-я → 429, Retry-After, 7-я → 429)
- **Автоматизация:** pre-request скрипт генерирует уникальный email, чтобы каждый запуск имел своё ведро

**Что покрыто:**

- 5 попыток `/login` — все `401` (пароль неверный, но запрос проходит)
- 6-я попытка — `429 Too Many Requests` с заголовком `Retry-After`
- 7-я попытка сразу — снова `429` (ведро пустое)

**Как использовать:**

1. Импортируй файл `api/postman/rate_limit_collection.json` в Postman.
2. Убедись, что переменная `base_url` = `http://localhost:8081` (Docker) или `http://localhost:8091` (локально).
3. Запусти коллекцию через **Run collection** с **Delay = 0 ms** — важно, чтобы запросы шли подряд без задержек.
4. Должно быть `totalPass: 9`, `totalFail: 0`.

> ⚠️ Коллекция рассчитана на пустое ведро. Если запустить её дважды подряд — второй раз начнётся с уже исчерпанного ведра. Pre-request генерирует уникальный email, поэтому повторный запуск сработает корректно.

## Управление миграциями

> 💡 При запуске `task docker-up` миграции применяются **автоматически** (сервис `migrate-auth` в `docker-compose.yml`). Ручные команды ниже нужны только для случаев, когда миграции запускаются отдельно (например, локальная разработка или откат).

> ⚠️ Миграции **не создают базу данных** — они только создают таблицы и схему внутри существующей БД. Если база `auth_db` (или `profiles_db`) отсутствует, `task auth:migrate-up` (или `task profiles:migrate-up`) упадёт с ошибкой `database "..." does not exist`.

Для создания БД вручную:
```bash
docker exec -it auth-postgres psql -U test -c "CREATE DATABASE auth_db;"
docker exec -it profiles-postgres psql -U test -c "CREATE DATABASE profiles_db;"
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

В проекте **четыре файла `.env`** — по одному на каждый контекст:

- **Корневой `.env`** — порты для `docker-compose` и Grafana.
- **`services/auth/.env`** — переменные Auth Service.
- **`services/gateway/.env`** — переменные Gateway Service.
- **`services/profiles/.env`** — переменные Profiles Service.

У каждого есть шаблон `.env.example` (в той же папке). Реальные `.env` в `.gitignore` и **не коммитятся**.

### Корневой `.env`

| Переменная | Значение по умолчанию | Назначение |
|------------|----------------------|------------|
| `POSTGRES_PORT` | `5432` | Порт PostgreSQL на хосте |
| `REDIS_AUTH_PORT` | `6379` | Порт Redis (Auth) на хосте |
| `REDIS_GATEWAY_PORT` | `6380` | Порт Redis (Gateway) на хосте |
| `AUTH_GRPC_PORT` | `50051` | Порт gRPC Auth на хосте |
| `AUTH_HTTP_PORT` | `8080` | Порт HTTP Auth (health/metrics) на хосте |
| `GATEWAY_HTTP_PORT` | `8081` | Порт HTTP Gateway на хосте |
| `PROFILES_POSTGRES_PORT` | `5433` | Порт PostgreSQL (Profiles) на хосте |
| `PROFILES_GRPC_PORT` | `50052` | Порт gRPC Profiles на хосте |
| `PROFILES_HTTP_PORT` | `8082` | Порт HTTP Profiles (health/metrics/ready) на хосте |
| `PROMETHEUS_PORT` | `9090` | Порт Prometheus на хосте |
| `GRAFANA_PORT` | `3000` | Порт Grafana на хосте |
| `GRAFANA_PASSWORD` | `admin` | Пароль администратора Grafana |
| `LOKI_PORT` | `3100` | Порт Loki на хосте |

### `services/auth/.env`

| Переменная | Значение по умолчанию | Назначение |
|------------|----------------------|------------|
| `GRPC_PORT` | `:50051` | Порт gRPC-сервера (внутри контейнера) |
| `HTTP_PORT` | `:8080` | Порт HTTP-сервера для /health и /metrics (внутри контейнера) |
| `JWT_SECRET` | — | Секрет для подписи JWT (минимум 32 байта) |
| `JWT_ISSUER` | `auth-service` | Издатель токена |
| `JWT_AUDIENCE` | `orange-team` | Аудитория токена |
| `ACCESS_TOKEN_TTL` | `15m` | Время жизни access-токена |
| `REFRESH_TOKEN_TTL` | `720h` | Время жизни refresh-токена (30 дней) |
| `ENABLE_REFLECTION` | `true` | gRPC reflection (для grpcurl). В проде — `false` |
| `POSTGRES_HOST` | `postgres-auth` | Хост PostgreSQL внутри Docker-сети |
| `POSTGRES_PORT` | `5432` | Порт PostgreSQL |
| `POSTGRES_USER` | `test` | Пользователь БД |
| `POSTGRES_PASSWORD` | — | Пароль БД |
| `POSTGRES_DB` | `auth_db` | Имя базы данных |
| `POSTGRES_TIMEOUT` | `30s` | Таймаут операций с БД |
| `REDIS_ADDR` | `redis-auth:6379` | Адрес Redis внутри Docker-сети |
| `REDIS_PASSWORD` | — | Пароль Redis (для локали можно простой, для прода — `openssl rand -hex 32`) |
| `REDIS_DB` | `0` | Номер логической БД Redis |
| `ENABLE_TLS` | `true` | Использовать TLS для gRPC |
| `TLS_CERT_FILE` | `/app/certs/server.crt` | Путь к сертификату (внутри контейнера) |
| `TLS_KEY_FILE` | `/app/certs/server.key` | Путь к приватному ключу |

### `services/gateway/.env`

| Переменная | Значение по умолчанию | Назначение |
|------------|----------------------|------------|
| `GATEWAY_HTTP_PORT` | `:8081` | HTTP-порт Gateway (внутри контейнера) |
| `AUTH_GRPC_ADDR` | `auth-service:50051` | Адрес Auth Service для gRPC-вызовов |
| `GATEWAY_TIMEOUT` | `10s` | Таймаут gRPC-запросов к Auth |
| `REDIS_ADDR` | `redis-gateway:6379` | Адрес Redis (свой инстанс) внутри Docker-сети |
| `REDIS_PASSWORD` | — | Пароль Redis (для локали можно простой, для прода — `openssl rand -hex 32`) |
| `REDIS_DB` | `0` | Номер логической БД Redis |
| `RATE_LIMIT_ENABLED` | `true` | Включить rate limiting |
| `RATE_LIMIT_LOGIN_RATE` | `5` | Токенов в минуту для `/login` |
| `RATE_LIMIT_LOGIN_BURST` | `5` | Вместимость ведра для `/login` |
| `RATE_LIMIT_LOGIN_INTERVAL` | `1m` | Период восстановления |
| `RATE_LIMIT_REGISTER_RATE` | `3` | Токенов в минуту для `/register` |
| `RATE_LIMIT_REGISTER_BURST` | `3` | Вместимость ведра |
| `RATE_LIMIT_REGISTER_INTERVAL` | `1m` | Период восстановления |
| `RATE_LIMIT_REFRESH_RATE` | `20` | Токенов в минуту для `/refresh` |
| `RATE_LIMIT_REFRESH_BURST` | `20` | Вместимость ведра |
| `RATE_LIMIT_REFRESH_INTERVAL` | `1m` | Период восстановления |
| `RATE_LIMIT_DEFAULT_RATE` | `60` | Токенов в минуту для защищённых |
| `RATE_LIMIT_DEFAULT_BURST` | `60` | Вместимость ведра |
| `RATE_LIMIT_DEFAULT_INTERVAL` | `1m` | Период восстановления |

### `services/profiles/.env`

| Переменная | Значение по умолчанию | Назначение |
|------------|----------------------|------------|
| `GRPC_PORT` | `:50052` | Порт gRPC-сервера (внутри контейнера) |
| `HTTP_PORT` | `:8080` | Порт HTTP-сервера для /health, /ready и /metrics (внутри контейнера) |
| `ENABLE_REFLECTION` | `true` | gRPC reflection (для grpcurl). В проде — `false` |
| `POSTGRES_HOST` | `postgres-profiles` | Хост PostgreSQL внутри Docker-сети |
| `POSTGRES_PORT` | `5432` | Порт PostgreSQL |
| `POSTGRES_USER` | `test` | Пользователь БД |
| `POSTGRES_PASSWORD` | — | Пароль БД |
| `POSTGRES_DB` | `profiles_db` | Имя базы данных |
| `POSTGRES_TIMEOUT` | `30s` | Таймаут операций с БД |
| `ENABLE_TLS` | `true` | Использовать TLS для gRPC |
| `TLS_CERT_FILE` | `/app/certs/server.crt` | Путь к сертификату (внутри контейнера) |
| `TLS_KEY_FILE` | `/app/certs/server.key` | Путь к приватному ключу |

### Общие настройки

Логгер применяется ко всем сервисам, но переменные задаются в каждом `.env` одинаково:

| Переменная | Значение по умолчанию | Назначение |
|------------|----------------------|------------|
| `LOGGER_LEVEL` | `DEBUG` | Уровень логирования (`DEBUG`, `INFO`, `WARN`, `ERROR`) |
| `LOGGER_FORMAT` | `json` | Формат логов (`text` или `json`) |
| `LOGGER_FOLDER` | `logs` | Папка для файлов логов |

> ⚠️ Реальные `.env` **не коммитятся** в репозиторий (добавлены в `.gitignore`). Для запуска скопируйте `.env.example` в `.env` в каждой из четырёх папок и заполните секреты.

## Architecture Decision Records

Ключевые архитектурные решения проекта зафиксированы в ADR — отдельные файлы в папке [`adr/`](adr/).

Список решений:

- [ADR-001: Профиль пользователя](adr/001-profile.md) — lazy-create, nullable-поля, `profile_completed`
- [ADR-002: `user_workout_score`](adr/002-user-workout-score.md) — почему не хранится в Profiles
- [ADR-003: Транзакции](adr/003-transactions.md) — границы транзакций и sync-вызовы
- [ADR-004: Паттерны микросервисов](adr/004-microservices-patterns.md) — что используем, что нет
- [ADR-005: Распределение портов](adr/005-port-allocation.md) — диапазоны и смещение +10
- [ADR-006: Управление секретами](adr/006-secrets-management.md) — SOPS + age
- [ADR-007: Известные проблемы и технический долг](adr/007-known-issues.md) — что осталось до продакшена

Подробнее — в [`adr/README.md`](adr/README.md).
