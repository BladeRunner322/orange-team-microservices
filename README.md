
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
  - [Exercises (упражнения)](#exercises-упражнения)
  - [Gateway (API Gateway)](#gateway-api-gateway)
  - [Refresh tokens flow](#refresh-tokens-flow)
  - [Rate Limiting flow](#rate-limiting-flow)
  - [RBAC flow](#rbac-flow)
- [HTTPS](#https)
- [CI/CD и деплой](#cicd-и-деплой)
- [Мониторинг и логирование](#мониторинг-и-логирование)
- [Бэкапы](#бэкапы)
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

В проекте **пять файлов `.env`** — по одному на каждый контекст:

- **Корневой `.env`** — порты для `docker-compose` и Grafana.
- **`services/auth/.env`** — переменные Auth Service.
- **`services/gateway/.env`** — переменные Gateway Service.
- **`services/profiles/.env`** — переменные Profiles Service.
- **`services/exercises/.env`** — переменные Exercises Service.

У каждого есть шаблон `.env.example`. Скопируйте все пять:

```bash
cp .env.example .env
cp services/auth/.env.example services/auth/.env
cp services/gateway/.env.example services/gateway/.env
cp services/profiles/.env.example services/profiles/.env
cp services/exercises/.env.example services/exercises/.env
```

Затем заполните секреты в каждом из них.

**В корневом `.env`:**

- `GRAFANA_PASSWORD` — пароль администратора Grafana (по умолчанию `admin`)
- `DOMAIN` — публичный домен для Caddy (например, `sololevelingms.duckdns.org`). См. [ADR-015](docs/adr/015-https-caddy.md)

**В `services/auth/.env`:**

- `JWT_SECRET` — секретный ключ для JWT (минимум 32 байта). Сгенерировать: `openssl rand -hex 32`
- `POSTGRES_PASSWORD` — пароль для PostgreSQL
- `REDIS_PASSWORD` — пароль для Redis (refresh-токены)

**В `services/gateway/.env`:**

- `AUTH_GRPC_ADDR` — адрес Auth Service (по умолчанию `auth-service:50051`)
- `PROFILES_GRPC_ADDR` — адрес Profiles Service (по умолчанию `profiles-service:50052`)
- `EXERCISES_GRPC_ADDR` — адрес Exercises Service (по умолчанию `exercises-service:50053`)
- `GRPC_CLIENT_TLS_MODE` — режим TLS для исходящих gRPC-соединений (`disabled` / `insecure` / `verify`). По умолчанию `insecure` — для self-signed сертификатов в dev/staging.
- `REDIS_PASSWORD` — пароль для Redis (rate limiting, можно тот же или отдельный)
- `TRUSTED_PROXIES` — CIDR-список доверенных прокси через запятую (опционально). Если пусто — заголовок `X-Forwarded-For` игнорируется, IP клиента берётся из `RemoteAddr`. Заполнять, только если перед Gateway стоит прокси (nginx, ALB).

**В `services/profiles/.env`:**

- `POSTGRES_PASSWORD` — пароль для PostgreSQL

**В `services/exercises/.env`:**

- `POSTGRES_PASSWORD` — пароль для PostgreSQL

> ⚠️ Все пять `.env` добавлены в `.gitignore` и **не коммитятся** в репозиторий. Секреты хранятся только локально.

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
- PostgreSQL Exercises (порт 5434) — для Exercises Service
- Redis Auth (порт 6379) — для refresh-токенов
- Redis Gateway (порт 6380) — для rate limiting
- Миграции Auth, Profiles и Exercises (создание таблиц)
- Auth-сервис (gRPC, порт 50051)
- Profiles-сервис (gRPC, порт 50052)
- Exercises-сервис (gRPC, порт 50053)
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

**Exercises (gRPC):**
```bash
grpcurl -insecure localhost:50053 list
```

Ожидаемый ответ Exercises:
```bash
exercises.ExercisesService
grpc.reflection.v1.ServerReflection
grpc.reflection.v1alpha.ServerReflection
```

**Exercises (HTTP):**
```bash
curl http://localhost:8083/health
curl http://localhost:8083/ready
```

Ожидаемые ответы: `{"service":"exercises","status":"ok"}` и `{"status":"ok","checks":{"postgres":"ok"}}`.

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
│   ├── ctxkeys/                      # ключи context, пропагируемые между сервисами (user_id, role, request_id)
│   ├── grpc/
│   │   ├── client/                   # конструктор gRPC-клиентов (TLS + interceptor)
│   │   ├── interceptors/             # gRPC-интерсепторы (логирование, метрики, recovery, user_id, request_id)
│   │   └── server/                   # конструктор gRPC-сервера (interceptors + TLS + reflection)
│   ├── health/                       # HTTP-хендлеры /health, /ready, /metrics для всех сервисов
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
│   │   │           ├── handlers/     # хендлеры по подпакетам: auth/, profiles/, exercises/, proxy/
│   │   │           ├── middleware/   # аутентификация, RBAC, rate limit, логирование, request_id
│   │   │           └── httputil/     # утилиты (SendJSON, SendError, GrpcErrorToHTTP)
│   │   ├── .env.example              # шаблон переменных Gateway Service
│   │   ├── .env.enc                  # зашифрованный .env (SOPS + age)
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
│   │   │       └── authgrpc/         # gRPC-сервер (обработчики AuthService)
│   │   ├── migrations/               # SQL-миграции для auth_db
│   │   ├── .env.example              # шаблон переменных Auth Service
│   │   ├── .env.enc                  # зашифрованный .env (SOPS + age)
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
│   │   │       └── profilesgrpc/     # gRPC-сервер (обработчики ProfilesService)
│   │   ├── migrations/               # SQL-миграции для profiles_db
│   │   ├── .env.example              # шаблон переменных Profiles Service
│   │   ├── .env.enc                  # зашифрованный .env (SOPS + age)
│   │   └── Dockerfile
│   │
│   ├── exercises/                    # ✅ ГОТОВ — Exercises Service (справочник упражнений)
│   │   ├── cmd/                      # точка входа (main.go)
│   │   ├── config/                   # конфигурация (envconfig)
│   │   ├── internal/
│   │   │   ├── bootstrap/            # сборка зависимостей приложения
│   │   │   ├── domain/               # Exercise, ExercisePatch, VO (name, description, difficulty, type)
│   │   │   ├── application/
│   │   │   │   ├── ports/            # интерфейс Repository
│   │   │   │   └── usecases/         # CreateExercise, GetExercise, GetExercises, PatchExercise, DeleteExercise
│   │   │   ├── infrastructure/
│   │   │   │   └── postgres_repo/    # реализация Repository для PostgreSQL
│   │   │   └── interfaces/
│   │   │       └── exercisesgrpc/    # gRPC-сервер (обработчики ExercisesService)
│   │   ├── migrations/               # SQL-миграции для exercises_db
│   │   ├── .env.example              # шаблон переменных Exercises Service
│   │   ├── .env.enc                  # зашифрованный .env (SOPS + age)
│   │   └── Dockerfile
│   │
│   ├── habits/                       # 📋 В ПЛАНЕ (не реализован)
│   │   └── ... (аналогично)
│   │
│   ├── workouts/                     # 📋 В ПЛАНЕ (не реализован)
│   │   └── ... (аналогично)
│   │
│   └── leaderboard/                  # 📋 В ПЛАНЕ (не реализован)
│       └── ... (аналогично)
│
├── docs/                             # Документация
│   └── adr/                          # Architecture Decision Records
│       ├── README.md
│       ├── 001-profile.md
│       ├── 002-user-workout-score.md
│       ├── 003-transactions.md
│       ├── 004-microservices-patterns.md
│       ├── 005-port-allocation.md
│       ├── 006-secrets-management.md
│       ├── 007-known-issues.md
│       ├── 008-exercise-lifecycle.md
│       ├── 009-service-dependencies.md
│       ├── 010-shared-infrastructure-packages.md
│       ├── 011-naming-and-code-style.md
│       ├── 012-ci-cd-optimization.md
│       ├── 013-infrastructure-requirements.md
│       ├── 014-backups-and-dr.md
│       ├── 015-https-caddy.md
│       └── 016-network-security.md
│
├── scripts/                          # скрипты (бэкапы, вспомогательное)
│   ├── backup-db.sh                  # дамп БД + выгрузка в S3
│   ├── rclone.conf.example           # шаблон конфига rclone
│   └── rclone.conf.enc               # зашифрованный rclone.conf (SOPS)
│
├── .env.example                      # шаблон корневого .env (порты, Grafana)
├── .env.enc                          # зашифрованный корневой .env (SOPS + age)
├── .sops.yaml                        # конфиг SOPS (публичный age-ключ)
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
├── Caddyfile                         # конфиг Caddy (HTTPS-терминация)
├── docker-compose.yml                # все контейнеры
├── Taskfile.yml                      # задачи для разработки
├── prometheus.yml                    # конфигурация Prometheus
├── alerts.yml                        # правила алертов (Prometheus)
├── alertmanager.yml                  # конфигурация Alertmanager (Telegram)
├── promtail-config.yml               # конфигурация Promtail
├── go.mod
├── go.sum
└── README.md
```
## Полная архитектура микросервисного приложения

### Что реализовано сейчас

Актуально на текущий момент работают четыре сервиса: **Auth**, **Gateway**, **Profiles** и **Exercises**. Остальные (Habits, Workouts, Leaderboard) — в плане.

```
┌────────────────────────────────────────────────────────────────────────────┐
│                          КЛИЕНТЫ (Внешние)                                 │
│                    Браузер / Мобильное приложение                          │
└────────────────────────────────┬───────────────────────────────────────────┘
                                 │ HTTPS (443)
                                 ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                      CADDY  (✅ ГОТОВ)                                     │
│                    (TLS-терминация + reverse proxy)                        │
│                                                                            │
│  • Слушает 80 (redirect) и 443 (HTTPS)                                     │
│  • Автоматический сертификат Let's Encrypt (HTTP-challenge)                │
│  • Проксирует в gateway:8081 внутри docker-сети                            │
└────────────────────────────────┬───────────────────────────────────────────┘
                                 │ HTTP (внутри docker-сети)
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
│  • Проксирует /exercises в Exercises (чтение — user, мутации — admin)      │
│  • Habits, Workouts, Leaderboard — пока заглушки 501                       │
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

┌────────────────────────────────────────────────────────────────────────────┐
│                     EXERCISES SERVICE  (✅ ГОТОВ)                          │
│                                                                            │
│  gRPC-методы:                                                              │
│  • CreateExercise, GetExercise, GetExercises, PatchExercise, DeleteExercise│
│                                                                            │
│  • Справочник упражнений (name, description, difficulty, type)             │
│  • type immutable, difficulty — snapshot (ADR-008)                         │
│  • Soft delete через deleted_at, идемпотентный DELETE (ADR-007 F-3)        │
│  • RBAC: чтение — все авторизованные, мутации — admin-only                 │
│  • БД: exercises_db (PostgreSQL)                                           │
│  • Порт: :50053                                                            │
│  • HTTP: /health, /ready, /metrics (:8080)                                 │
└────────────────────────────────────────────────────────────────────────────┘
```

### Целевая архитектура

(Статусы: ✅ — реализован, 📋 — в плане. Описана полная архитектура проекта — то, к чему идём.)

```
┌────────────────────────────────────────────────────────────────────────────┐
│                          КЛИЕНТЫ (Внешние)                                 │
│                    Браузер / Мобильное приложение                          │
└────────────────────────────────┬───────────────────────────────────────────┘
                                 │ HTTPS (443)
                                 ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                      CADDY  (✅ ГОТОВ)                                     │
│                    (TLS-терминация + reverse proxy)                        │
│                                                                            │
│  • Слушает 80 (redirect) и 443 (HTTPS)                                     │
│  • Автоматический сертификат Let's Encrypt (HTTP-challenge)                │
│  • Проксирует в gateway:8081 внутри docker-сети                            │
└────────────────────────────────┬───────────────────────────────────────────┘
                                 │ HTTP (внутри docker-сети)
                                 ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                         API GATEWAY                                        │
│                    (HTTP → gRPC прокси)                                    │
│                                                                            │
│  ✅ Функции:                                                               │
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
│   (✅ ГОТОВ)         │ │   (✅ ГОТОВ)         │ │  (✅ ГОТОВ)          │
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
│   (📋 В ПЛАНЕ)       │ │  SERVICE          │ │  SERVICE            │
│                      │ │  (📋 В ПЛАНЕ)     │ │  (📋 В ПЛАНЕ)       │
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

- **PostgreSQL:** `postgres-auth` (порт 5432) → `auth_db`; `postgres-profiles` (порт 5433) → `profiles_db`; `postgres-exercises` (порт 5434) → `exercises_db`
- **Redis:** `redis-auth` (порт 6379) — refresh-токены; `redis-gateway` (порт 6380) — rate limiting
- **Миграции:** `migrate-auth` → для `auth_db`; `migrate-profiles` → для `profiles_db`; `migrate-exercises` → для `exercises_db`
- **Мониторинг:** Prometheus (9090), Grafana (3000)
- **Логи:** Loki (3100), Promtail (9080)
- **HTTPS:** Caddy (80, 443) — TLS-терминация для Gateway, см. [ADR-015](docs/adr/015-https-caddy.md)
- **Сетевая безопасность:** внутренние порты привязаны к `127.0.0.1`, ufw (только 22/80/443), fail2ban на SSH — см. [ADR-016](docs/adr/016-network-security.md)

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
│  🔷 HTTPS-терминация:                                               │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  Caddy (порты 80, 443) → TLS + reverse proxy на Gateway      │   │
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
| **Profiles** | Health / Ready / Metrics | HTTP | `8080` | `8082` | `8092` | Смещение +10, переопределяется в `docker-compose.yml`
| **Exercises** | gRPC | gRPC | `50053` | `50053` | `50063` | Смещение +10 |
| **Exercises** | Health / Ready / Metrics | HTTP | `8080` | `8083` | `8093` | Смещение +10, переопределяется в `docker-compose.yml` |
| **PostgreSQL (Auth)** | База данных | TCP | `5432` | `5432` | — | Используется через Docker, проброс на хост |
| **PostgreSQL (Profiles)** | База данных | TCP | `5432` | `5433` | — | Используется через Docker, проброс на хост |
| **PostgreSQL (Exercises)** | База данных | TCP | `5432` | `5434` | — | Используется через Docker, проброс на хост | |
| **Redis Auth** | Refresh-токены | TCP | `6379` | `6379` | — | Только для Auth Service |
| **Redis Gateway** | Rate limiting | TCP | `6379` | `6380` | — | Только для Gateway Service |
| **Prometheus** | Метрики | HTTP | `9090` | `9090` | — | Только в Docker |
| **Grafana** | Визуализация | HTTP | `3000` | `3000` | — | Только в Docker |
| **Loki** | Логи | HTTP | `3100` | `3100` | — | Только в Docker |
| **Caddy** | HTTPS | HTTPS | `443` | `443` | — | TLS-терминация, см. [ADR-015](docs/adr/015-https-caddy.md) |
| **Caddy** | HTTP (redirect) | HTTP | `80` | `80` | — | Только редирект на HTTPS |

### Логика смещения портов

- **Docker** — используются стандартные порты:  
  `50051` (Auth gRPC), `8080` (Auth HTTP), `8081` (Gateway HTTP), `50052` (Profiles gRPC), `8082` (Profiles HTTP), `50053` (Exercises gRPC), `8083` (Exercises HTTP).  
  PostgreSQL: `postgres-auth` → `5432`, `postgres-profiles` → `5433`, `postgres-exercises` → `5434` (разные порты на хосте, чтобы не конфликтовать).  
  Redis: `redis-auth` → `6379`, `redis-gateway` → `6380` (разные порты на хосте, чтобы не конфликтовать).

- **Локальная разработка** — порты приложений сдвинуты на **+10**:  
  `50061`, `8090`, `8091`, `50062`, `8092`, `50063`, `8093` — чтобы не конфликтовать с запущенными Docker-контейнерами.  
  PostgreSQL и оба Redis для локальной разработки используются из Docker через проброс на `localhost` (`5432`, `5433`, `5434`, `6379`, `6380`).

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

**Exercises Service:**
```bash
task exercises:run
```

- gRPC: `localhost:50063`
- HTTP (health/ready/metrics): `localhost:8093`

**Требования для локального запуска:**
- PostgreSQL запущен через Docker: `task auth:postgres-up`
- Redis запущен через Docker: `task auth:redis-up`
- Миграции применены: `task auth:migrate-up`
- PostgreSQL для Profiles запущен: `task profiles:postgres-up`
- Миграции Profiles применены: `task profiles:migrate-up`
- PostgreSQL для Exercises запущен: `task exercises:postgres-up`
- Миграции Exercises применены: `task exercises:migrate-up`
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

- **Lazy-create**: пустая запись создаётся при первом чтении профиля (см. [ADR-001](docs/adr/001-profile.md)).

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

### Exercises (упражнения)

- Назначение: справочник упражнений (название, описание, сложность, тип).

- Протокол: gRPC

- gRPC-методы: `CreateExercise`, `GetExercise`, `GetExercises`, `PatchExercise`, `DeleteExercise`

- Порт: 50053

- БД: PostgreSQL (схема exercises, таблица exercises)

- **Lifecycle** (см. [ADR-008](docs/adr/008-exercise-lifecycle.md)):
  - `type` — immutable, после создания не меняется.
  - `difficulty` — снапшот на момент создания workout_exercise (в Workouts).
  - `name`, `description` — читаются live (в Workouts).
  - DELETE — soft через `deleted_at`, идемпотентен (см. [ADR-007](docs/adr/007-known-issues.md) F-3).

- **RBAC:**
  - `GET /exercises`, `GET /exercises/{id}` — любой авторизованный.
  - `POST`, `PATCH`, `DELETE` — admin-only.

- **Список** (`GetExercises`) возвращает только активные (`deleted_at IS NULL`).
- **GetExercise(id)** возвращает и удалённые, с флагом `is_deleted`.

- Команды:

  | Команда | Назначение |
  |---------|------------|
  | `task exercises:run` | Запуск локально (gRPC `50063`, HTTP `8093`) |
  | `task exercises:build` | Сборка Docker-образа |
  | `task exercises:rebuild` | Пересборка без кеша (с обновлением vendor) |
  | `task exercises:up` | Запуск в Docker Compose |
  | `task exercises:restart` | Перезапуск (rebuild + up) |
  | `task exercises:logs` | Просмотр логов Exercises |

- Команды для PostgreSQL:

  | Команда | Назначение |
  |---------|------------|
  | `task exercises:postgres-up` | Запуск PostgreSQL |
  | `task exercises:postgres-down` | Остановка PostgreSQL |
  | `task exercises:postgres-logs` | Просмотр логов PostgreSQL |
  | `task exercises:postgres-psql` | Консоль psql |

- Команды для миграций:

  | Команда | Назначение |
  |---------|------------|
  | `task exercises:migrate-create -- <name>` | Создать новую миграцию |
  | `task exercises:migrate-up` | Применить миграции |
  | `task exercises:migrate-down -- 1` | Откатить последнюю |
  | `task exercises:migrate-version` | Показать текущую версию |

#### Healthcheck и readiness

Exercises предоставляет HTTP-эндпоинты для проверки состояния:

| Эндпоинт | Назначение | Ответ |
|----------|------------|-------|
| `/health` | Liveness — процесс жив | `{"status":"ok","service":"exercises"}` |
| `/ready` | Readiness — PostgreSQL доступен | `{"status":"ok","checks":{"postgres":"ok"}}` или `503` |
| `/metrics` | Prometheus-метрики | text/plain |

**Порт:** `8080` (Docker) / `8093` (локально)

Проверка:

```bash
curl http://localhost:8083/health
curl http://localhost:8083/ready
curl http://localhost:8083/metrics
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
| `/ready` | Readiness — зависимости доступны (Redis + Auth + Profiles + Exercises) | `{"status":"ok","checks":{"redis":"ok","auth":"ok","profiles":"ok","exercises":"ok"}}` или `503` |
| `/metrics` | Prometheus-метрики | text/plain |

**Проверка:**

```bash
curl http://localhost:8081/health
curl http://localhost:8081/ready
curl http://localhost:8081/metrics
```

Если любая из зависимостей недоступна, `/ready` вернёт `503`:

```json
{"status":"not_ready","checks":{"redis":"ok","auth":"fail: auth connection state: TransientFailure","profiles":"ok","exercises":"ok"}}
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
| GET | `/exercises` | Список упражнений → Exercises.GetExercises | ✅ | — |
| GET | `/exercises/{exerciseId}` | Упражнение по ID → Exercises.GetExercise | ✅ | — |
| POST | `/exercises` | Создать упражнение → Exercises.CreateExercise | ✅ | **admin** |
| PATCH | `/exercises/{exerciseId}` | Патч упражнения → Exercises.PatchExercise | ✅ | **admin** |
| DELETE | `/exercises/{exerciseId}` | Soft delete → Exercises.DeleteExercise | ✅ | **admin** |
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
| `PATCH /exercises/{exerciseId}` | `admin` |
| `DELETE /exercises/{exerciseId}` | `admin` |

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

## HTTPS

Клиент ходит на **`https://sololevelingms.duckdns.org`** (DuckDNS-поддомен, см. [ADR-015](docs/adr/015-https-caddy.md)). TLS-терминация — **Caddy** перед Gateway.

### Схема

```
Клиент → HTTPS :443 → Caddy → HTTP gateway:8081 (внутри docker-сети) → Auth/Profiles/Exercises
       → HTTP  :80  → Caddy → 308 Redirect → HTTPS
```

- **Порты 80 и 443** — единственные открытые наружу. Всё остальное — внутри docker-сети.
- **Caddy** сам получает и продлевает сертификаты Let's Encrypt (HTTP-challenge, без плагинов).
- **Конфиг** — `Caddyfile` в корне репо, 3 строки.
- **Сертификаты** хранятся в volume `caddy_data`, переживают перезапуск контейнера.

### Проверка

```bash
# HTTPS + healthcheck
curl https://sololevelingms.duckdns.org/health
# → {"status":"ok","service":"gateway"}

# HTTP → HTTPS редирект
curl -I http://sololevelingms.duckdns.org/health
# → 308 Permanent Redirect

# Логи Caddy (успешная выдача сертификата)
docker logs caddy --tail 50
```

### Переезд на платный домен

15 минут:

1. Купить домен у регистратора.
2. Настроить A-запись на `136.234.4.93`.
3. Поменять `DOMAIN=` в `.env` на новый домен.
4. `task sops:encrypt`, закоммитить, задеплоить.

Caddy автоматически выпустит новый сертификат, старый перестанет использоваться.

## Сетевая безопасность

Полностью — в [ADR-016](docs/adr/016-network-security.md). Ниже — что нужно знать при работе с проектом.

### Что открыто снаружи

Только три порта доступны из интернета:

| Порт | Сервис | Назначение |
|---|---|---|
| 22 | SSH | Ручной доступ + CD (`appleboy/ssh-action`) |
| 80 | Caddy | HTTP → HTTPS redirect |
| 443 | Caddy | HTTPS |

**Всё остальное закрыто** — на двух уровнях:

1. **Docker bind.** В `docker-compose.yml` внутренние порты привязаны к `127.0.0.1`:
   ```
   ports:
     - "127.0.0.1:${AUTH_GRPC_PORT}:50051"
   ```
   Порт доступен только с самого сервера, снаружи — нет. Caddy — исключение, у него `0.0.0.0:80` и `0.0.0.0:443`.

2. **ufw.** Файрвол дропает всё, кроме 22/80/443:
   ```
   ufw status
   # Status: active
   # 22/tcp  ALLOW
   # 80/tcp  ALLOW
   # 443/tcp ALLOW
   ```

**Двойная защита:** если один слой сломается (баг в Docker, случайная правка compose без `127.0.0.1`), второй закроет.

### fail2ban на SSH

Установлен `fail2ban` с кастомным `/etc/fail2ban/jail.local`:

```
[DEFAULT]
bantime = 1h
findtime = 10m
maxretry = 5

[sshd]
enabled = true
```

**Как работает:** после 5 неудачных SSH-попыток с одного IP — бан на час. Боты, перебирающие пароли, отсекаются. Логи `/var/log/auth.log` чистые.

**Проверить статус:**
```bash
fail2ban-client status sshd
```

Увидишь список забаненных IP (`Banned IP list`). Скорее всего, там будут IP из Ирана/Нидерландов/Китая — это фоновые боты, которые сканируют весь интернет.

**Разбанить вручную (если попал свой IP):**
```bash
fail2ban-client set sshd unbanip <IP>
```

### gRPC между сервисами — `insecure` (осознанно)

Gateway ходит к Auth/Profiles/Exercises с `GRPC_CLIENT_TLS_MODE=insecure`. Сертификаты не проверяются.

**Почему так:**
- Все 4 сервиса в одной docker-сети, gRPC-порты закрыты снаружи.
- MITM возможен только тем, кто уже внутри — то есть уже скомпрометировал систему.
- Переход на `verify` требует нормальных сертификатов с SAN (`auth-service`, `profiles-service`) — это 2–3 часа работы без реальной выгоды сейчас.

**Триггеры для перехода на `verify` + mTLS:**
- Сервисы выйдут из одной docker-сети.
- Появится внешний gRPC-потребитель.
- Compliance/аудит.

Детали — в [ADR-016](docs/adr/016-network-security.md).

### Проверка периметра

С локальной машины (не с сервера):
```bash
nmap -Pn -p 22,80,443,3000,3100,50051,50052,50053,5432,5433,5434,6379,6380,8080,8081,8082,8083,9090 136.234.4.93
```

Ожидаемо: `open` только для 22, 80, 443. Остальные — `filtered`.

### Ручная работа с сервером

Если нужен доступ к внутренним сервисам (посмотреть метрики, зайти в psql):
- **PostgreSQL:** `docker compose exec postgres-auth psql -U test -d auth_db`
- **Redis:** `docker compose exec redis-auth redis-cli -a $REDIS_PASSWORD`
- **Метрики:** `curl http://localhost:8080/metrics` (с самого сервера)
- **Prometheus UI:** через SSH-туннель: `ssh -L 9090:localhost:9090 root@136.234.4.93`, потом `http://localhost:9090` в браузере.

## CI/CD и деплой

Проект использует **GitHub Actions** для автоматической проверки, сборки и деплоя сервисов.

### Требования к серверу

Стек состоит из 16+ контейнеров (4 сервиса, 3×PostgreSQL, 2×Redis,
Prometheus, Grafana, Loki, Promtail, Caddy). Для комфортной работы нужен
сервер со следующими характеристиками:

| Параметр | Минимум | Комфортно |
|----------|---------|-----------|
| **RAM** | 4 ГБ | 8 ГБ |
| **CPU** | 2 vCPU | 2–4 vCPU |
| **Диск** | 60 ГБ SSD | 80+ ГБ SSD/NVMe |

> ⚠️ **На 512 МБ стек не запустится.** Docker daemon, Prometheus,
> Grafana и Loki вместе съедают ~800–1200 МБ. На слабом сервере
> система уходит в swap, OOM killer убивает контейнеры, деплой
> тормозит. Если бюджет ограничен — отключай мониторинг
> (`task monitoring-down`, `task logging-down`), это самая тяжёлая
> часть стека.

**Требуется:** Ubuntu 22.04+ (или другой Linux с ядром 5.x+), Docker 24+,
Docker Compose v2.20+, SSH-доступ для деплоя.

**Проверено на:** Selectel VDS (8 ГБ / 4 vCPU).
Подходит любой VPS с указанными характеристиками.

> 💡 Полная архитектура (7 сервисов, 7 PostgreSQL) потребует **16 ГБ RAM / 4-6 vCPU / 200+ ГБ NVMe**. Подробнее — [ADR-013](docs/adr/013-infrastructure-requirements.md).

### Триггеры

- **Pull Request в `main`** → запускается только **CI** (тесты, сборка, smoke-тест).
- **Push в `main`** (после вливания PR) → запускается **CD** (сборка образа, публикация в GHCR, деплой на сервер).

### CI (Continuous Integration)

При каждом PR выполняются:

- Установка Go.
- Кеширование модулей.
- `go test -v ./... -short` — юнит-тесты.
- `go build` для `auth`, `gateway`, `profiles` и `exercises` — smoke-тест сборки.
- `go test -tags=integration -v -timeout=15m ./...` — интеграционные тесты (Testcontainers).

Интеграционные тесты добавляют к CI ~3-5 минут. Если CI зелёный — PR готов к слиянию.

### CD (Continuous Deployment)

При пуше в `main`:

1. Определяются изменённые сервисы (`services/auth/**`, `services/gateway/**`, `services/profiles/**`, `services/exercises/**`, `.github/workflows/ci-cd.yml`, `pkg/**`).
2. Собираются Docker-образы для изменённых сервисов через **matrix** (параллельно).
3. Образы публикуются в **GitHub Container Registry**:
   - `ghcr.io/bladerunner322/orange-team-microservices/auth:latest` (+ `:<git-sha>`)
   - `ghcr.io/bladerunner322/orange-team-microservices/gateway:latest` (+ `:<git-sha>`)
   - `ghcr.io/bladerunner322/orange-team-microservices/profiles:latest` (+ `:<git-sha>`)
   - `ghcr.io/bladerunner322/orange-team-microservices/exercises:latest` (+ `:<git-sha>`)
4. По SSH выполняется деплой на продакшен-сервер (только если `any_changed == 'true'`):
   - Обновление кода (`git pull origin main`).
   - Логин в GHCR.
   - `docker compose pull auth gateway profiles exercises caddy` — скачивание свежих образов.
   - `docker compose up -d --no-build auth gateway profiles exercises caddy` — перезапуск.

Используется **Docker layer caching** (`type=gha`) — повторные сборки быстрее в 2-3 раза. Подробнее — [ADR-012](docs/adr/012-ci-cd-optimization.md).

### Секреты GitHub Actions

Для деплоя используются секреты репозитория (Settings → Secrets and variables → Actions):

| Секрет | Назначение |
|--------|------------|
| `SERVER_HOST` | IP или домен продакшен-сервера |
| `SERVER_USER` | SSH-пользователь (обычно `root`) |
| `SERVER_SSH_KEY` | Приватный SSH-ключ для доступа к серверу |
| `SOPS_AGE_KEY` | Приватный age-ключ для расшифровки `.env.enc` |

`TELEGRAM_BOT_TOKEN` и `TELEGRAM_CHAT_ID` **не дублируются в GitHub Secrets** — расшифровываются из `.env.enc` через `sops --extract` в notify job. См. [ADR-012](docs/adr/012-ci-cd-optimization.md).

### Уведомления о деплое

После деплоя (или провала) в Telegram приходит уведомление:

- `✅ Deploy success` — все job'ы прошли успешно.
- `❌ Deploy failed` — какой-то job упал, в сообщении указан упавший.
- `⏭️ Nothing to deploy` — изменений нет, деплой пропущен.

Уведомление отправляется с GitHub-раннера напрямую в Telegram (без WARP — блокировки на GitHub-раннере нет).

### Проверка после деплоя

После успешного CD проверь на сервере:

**Health:**

```bash
curl http://<SERVER_HOST>:8081/health   # Gateway
curl http://<SERVER_HOST>:8080/health   # Auth
curl http://<SERVER_HOST>:8082/health   # Profiles
curl http://<SERVER_HOST>:8083/health   # Exercises
```

Все четыре должны вернуть `{"status":"ok","service":"..."}`.

**Ready:**

```bash
curl http://<SERVER_HOST>:8081/ready    # Gateway → проверка Redis + Auth/Profiles/Exercises
curl http://<SERVER_HOST>:8080/ready    # Auth → проверка Postgres + Redis
curl http://<SERVER_HOST>:8082/ready    # Profiles → проверка Postgres
curl http://<SERVER_HOST>:8083/ready    # Exercises → проверка Postgres
```

Должны вернуть `{"status":"ok","checks":{...}}`.

**Prometheus targets:**

Открой `http://<SERVER_HOST>:9090` → **Status → Targets**. Все четыре job'а должны быть в статусе **UP**:

- `auth` — target `auth:8080`
- `gateway` — target `gateway:8081`
- `profiles` — target `profiles:8080`
- `exercises` — target `exercises:8080`

### Обновление `.env` на сервере (SOPS + age)

Пять файлов `.env` (корневой + по одному на сервис) **не хранятся в git** (в `.gitignore`), но их **зашифрованные версии** (`.env.enc`) коммитятся. CD расшифровывает их при деплое. Никаких `nano` на сервере — все изменения проходят через git, ревью и CI/CD.

**Структура в репозитории:**

```
orange-team-microservices/
├── .env.enc                          # зашифрованный корневой (порты, Grafana)
└── services/
    ├── auth/.env.enc
    ├── gateway/.env.enc
    ├── profiles/.env.enc
    └── exercises/.env.enc
```

**Как отредактировать переменную:**

1. Локально:
   ```bash
   task sops:edit -- services/auth/.env.enc
   ```
   SOPS расшифрует файл, откроет `$EDITOR`, при сохранении зашифрует обратно.
2. Закоммить изменения `.env.enc` и запушь в `main`.
3. CD расшифрует `.env.enc` на раннере и зальёт `.env` на сервер.

**Что делает CD (deploy job):**

1. Устанавливает `sops`, расшифровывает `.env.enc` → `.env` (используя `SOPS_AGE_KEY` из GitHub Secrets).
2. Заливает `.env` на сервер через `scp-action`.
3. Валидирует `.env` против `.env.example` — падает с явной ошибкой, если чего-то не хватает.
4. `docker compose pull` + `up -d --no-build`.

**Ручные команды (для локальной настройки):**

```bash
task sops:encrypt   # .env → .env.enc (все 5 файлов)
task sops:decrypt   # .env.enc → .env (например, на новой машине)
```

> ⚠️ **Приватный age-ключ** хранится в GitHub Secrets (`SOPS_AGE_KEY`) и локально. Если он утечёт — все `.env.enc` скомпрометированы. См. [ADR-006](docs/adr/006-secrets-management.md).

## Мониторинг и логирование

В проекте настроен полный стек для мониторинга и логирования:

- **Prometheus** — сбор метрик
- **Alertmanager** — маршрутизация алертов в Telegram
- **node-exporter** — метрики хоста (CPU, RAM, диск)
- **Loki** — агрегация логов
- **Grafana** — визуализация

### Метрики (Prometheus)

Все четыре сервиса отдают метрики в формате Prometheus.

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

**Exercises:**

- **Порт:** `8080` (Docker) / `8093` (локально)
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
curl http://localhost:8083/metrics   # Exercises
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
- `{service="exercises"}` — логи Exercises в Loki

> ⚠️ Дашборды Grafana не сохраняются при `task docker-down-v` (удаление volume `grafana-storage`). Для постоянного хранения настрой **provisioning** (папка `grafana/provisioning/`) или экспортируй дашборд в JSON и положи его в репозиторий.

### Запуск мониторинга

Все команды доступны через Taskfile:

```bash
task docker-up        # поднимает всё (приложение + мониторинг + логи)
task monitoring-up    # только Prometheus + Grafana
task logging-up       # только Loki + Promtail
task alertmanager-up  # только Alertmanager
```

### Управление Alertmanager

```bash
task secrets:generate      # создать secrets/* из локального .env (автоматически в docker-up)
task alertmanager-up       # запустить
task alertmanager-down     # остановить
task alertmanager-logs     # логи
task alertmanager-restart  # пересоздать (для применения изменений alertmanager.yml)
task alertmanager-reload   # hot-reload конфига без рестарта
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

## Бэкапы

PostgreSQL-базы (`auth_db`, `profiles_db`, `exercises_db`) бэкапятся скриптом
[`scripts/backup-db.sh`](scripts/backup-db.sh).

Redis **не бэкапится** — там только refresh-токены (TTL 30 дней) и rate-limit-вёдра
(TTL секунды-минуты). При потере Redis пользователи просто перелогинятся.

### Что делает скрипт

- Дампит три базы через `pg_dump` (контейнеры `auth-postgres`, `profiles-postgres`, `exercises-postgres`).
- Сжимает каждый дамп через `gzip`.
- Кладёт в `/root/backups/YYYY-MM-DD_HH-MM-SS/` (папка на каждую дату).
- Удаляет папки старше `RETENTION_DAYS` (по умолчанию 7 дней).
- Падает с явной ошибкой при `pg_dump` failure или подозрительно маленьком дампе (< 100 байт).

### Как запускается

Через cron на продакшн-сервере:

```
0 3 * * * /root/projects/orange-team-microservices/scripts/backup-db.sh >> /var/log/backup.log 2>&1
```

Раз в сутки в 03:00 UTC. Логи — в `/var/log/backup.log`.

**Где лежат бэкапы**

**Локально на сервере:** `/root/backups/YYYY-MM-DD_HH-MM-SS/`. Ротация — 7 дней.

**Off-site (S3):** `s3://orange-team-backups/` в Selectel Object Storage (см. [ADR-014](docs/adr/014-backups-and-dr.md)). Структура:

```
s3://orange-team-backups/
├── auth/2026-10-07_03-48-09/auth_db.sql.gz
├── profiles/2026-10-07_03-48-09/profiles_db.sql.gz
└── exercises/2026-10-07_03-48-09/exercises_db.sql.gz
```

Ротация — 30 дней. Трафик между VDS и S3 внутри Selectel не тарифицируется.

**Как посмотреть, что в S3:**

```bash
rclone lsf -R selectel:orange-team-backups
```

**Переменные (в `scripts/backup-db.sh`):**

| Переменная | По умолчанию | Назначение |
|---|---|---|
| `RCLONE_REMOTE` | `selectel` | Имя remote в `~/.config/rclone/rclone.conf`. Пустое — off-site отключается. |
| `S3_BUCKET` | `orange-team-backups` | Имя бакета. |
| `S3_RETENTION_DAYS` | `30` | Срок хранения в S3. |

**Конфиг rclone:** `~/.config/rclone/rclone.conf` на сервере. Расшифровывается из `scripts/rclone.conf.enc` (SOPS + age) при деплое через CD. См. [ADR-006](docs/adr/006-secrets-management.md).

> ⚠️ **Локальные бэкапы не защищают от смерти диска или удаления сервера.** Off-site — обязательное дополнение, реализовано в [ADR-014](docs/adr/014-backups-and-dr.md).

### Метрика бэкапа

После успешного бэкапа скрипт пишет в `backup-metrics/backup.prom` Unix-timestamp последнего успешного запуска. Директория монтируется в node-exporter через `--collector.textfile.directory`, метрика `backup_last_success_timestamp` доступна Prometheus.

Алерт `BackupTooOld` сработает, если бэкап старше 36 часов. Алерт `BackupMetricMissing` — если метрика пропала совсем. См. [ADR-017](docs/adr/017-alerting.md).

Проверить метрику:
```bash
curl -s http://localhost:9100/metrics | grep backup_last_success_timestamp
```
(порт 9100 — node-exporter; снаружи закрыт, доступен через SSH-туннель или с самого сервера)


### Восстановление из локального бэкапа

**1. Распаковать дамп:**

```bash
gunzip -c /root/backups/YYYY-MM-DD_HH-MM-SS/auth_db.sql.gz > /tmp/auth_db.sql
```

**2. Создать временную базу:**

```bash
docker exec -it auth-postgres psql -U test -d postgres -c "CREATE DATABASE auth_restore;"
```

**3. Залить дамп:**

```bash
docker exec -i auth-postgres psql -U test -d auth_restore < /tmp/auth_db.sql
```

**4. Проверить, что данные на месте:**

```bash
docker exec -it auth-postgres psql -U test -d auth_restore -c "SELECT COUNT(*) FROM auth.users;"
```

### Проверка восстановления

**Раз в месяц** — прогонять восстановление последнего дампа в тестовую базу
(`auth_restore`), убеждаться, что таблицы и данные на месте, удалять тестовую БД.

Бэкап без проверки восстановления — **не бэкап**. Классическая ошибка: cron годами
пишет дампы, при первом инциденте выясняется, что они побитые.

### Восстановление из S3 (off-site)

Если сервер потерян полностью — сначала восстановить проект на новом сервере (клонировать репо, развернуть через CD, чтобы `.env` и `rclone.conf` расшифровались), потом:

```bash
# 1. Список доступных дампов
rclone lsf selectel:orange-team-backups/auth/

# 2. Скачать нужный дамп
rclone copy selectel:orange-team-backups/auth/2026-10-07_03-48-09/auth_db.sql.gz /tmp/restore/

# 3. Распаковать
gunzip -c /tmp/restore/auth_db.sql.gz > /tmp/restore/auth_db.sql

# 4. Восстановить в чистую БД
docker exec -it auth-postgres psql -U test -d postgres -c "CREATE DATABASE auth_restore;"
docker exec -i auth-postgres psql -U test -d auth_restore < /tmp/restore/auth_db.sql

# 5. Проверить, что данные на месте
docker exec -it auth-postgres psql -U test -d auth_restore -c "SELECT COUNT(*) FROM auth.users;"
```

Полная процедура и цели RPO/RTO — в [ADR-014](docs/adr/014-backups-and-dr.md).

### Настройка cron (одноразово, при первом деплое)

```bash
ssh root@<SERVER_HOST>
crontab -e
# Добавить строку:
# 0 3 * * * /root/projects/orange-team-microservices/scripts/backup-db.sh >> /var/log/backup.log 2>&1
```

Проверка:

```bash
crontab -l              # должна быть строка
cat /var/log/backup.log # утром следующего дня
```

### Отключение бэкапов

Удалить строку из `crontab -e`. Скрипт останется в репозитории, но запускаться не будет.

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
- **Покрытие:** 86 проверок (50 запросов, организованы в 7 папок)
- **Структура:** Healthcheck, Auth (Register, Login/Refresh/Logout), Profiles (`/users/me`), Exercises (CRUD + RBAC), Stubs, Errors
- **Автоматизация:** pre-request скрипт генерирует уникальный email; `access_token`, `refresh_token`, `userId` сохраняются в collection variables. Environment не требуется — работает сразу после импорта.

**Что покрыто:**

- Healthcheck
- Регистрация (успех, дубликат, невалидный email, слабый пароль)
- Логин (успех, неверные учётные данные)
- Обновление токенов через `/refresh` (успех, невалидный refresh)
- Logout и проверка, что refresh после logout не работает
- `/users/me` — профиль пользователя: GET (lazy-create), PATCH (все поля, частичный патч, null-сброс, невалидное значение), DELETE, повторный GET
- `/exercises` — CRUD: GET (список, по id), POST (admin, дубликат, валидация), PATCH (admin), DELETE (admin, идемпотентный, soft delete через `is_deleted=true`)
- Защищённые эндпоинты (с токеном, без токена, с невалидным токеном)
- RBAC: `POST /exercises` от non-admin → 403, чтение от user → 200
- Заглушки для будущих сервисов (Habits, Workouts, Leaderboard)

**Как использовать:**

1. Импортируй файл `api/postman/gateway_collection.json` в Postman.
2. Убедись, что переменная `base_url` = `http://localhost:8081` (Docker) или `http://localhost:8091` (локально).
3. Запусти коллекцию через **Run collection** — все тесты должны пройти.

> 💡 Environment не нужен — коллекция использует collection variables. Достаточно импортировать JSON и нажать **Run collection**.

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

> 💡 При запуске `task docker-up` миграции применяются **автоматически** (сервисы `migrate-auth`, `migrate-profiles` и `migrate-exercises` в `docker-compose.yml`). Ручные команды ниже нужны только для случаев, когда миграции запускаются отдельно (например, локальная разработка или откат).

> ⚠️ Миграции **не создают базу данных** — они только создают таблицы и схему внутри существующей БД. Если база `auth_db` (или `profiles_db`) отсутствует, `task auth:migrate-up` (или `task profiles:migrate-up`) упадёт с ошибкой `database "..." does not exist`.

Для создания БД вручную:
```bash
docker exec -it auth-postgres psql -U test -c "CREATE DATABASE auth_db;"
docker exec -it profiles-postgres psql -U test -c "CREATE DATABASE profiles_db;"
docker exec -it exercises-postgres psql -U test -c "CREATE DATABASE exercises_db;"
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

В проекте **пять файлов `.env`** — по одному на каждый контекст:

- **Корневой `.env`** — порты для `docker-compose` и Grafana.
- **`services/auth/.env`** — переменные Auth Service.
- **`services/gateway/.env`** — переменные Gateway Service.
- **`services/profiles/.env`** — переменные Profiles Service.
- **`services/exercises/.env`** — переменные Exercises Service.

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
| `EXERCISES_POSTGRES_PORT` | `5434` | Порт PostgreSQL (Exercises) на хосте |
| `EXERCISES_GRPC_PORT` | `50053` | Порт gRPC Exercises на хосте |
| `EXERCISES_HTTP_PORT` | `8083` | Порт HTTP Exercises (health/metrics/ready) на хосте |
| `PROMETHEUS_PORT` | `9090` | Порт Prometheus на хосте |
| `GRAFANA_PORT` | `3000` | Порт Grafana на хосте |
| `GRAFANA_PASSWORD` | `admin` | Пароль администратора Grafana |
| `LOKI_PORT` | `3100` | Порт Loki на хосте |
| `NODE_EXPORTER_PORT` | `9100` | Порт node-exporter на хосте (только локально) |
| `ALERTMANAGER_PORT` | `9093` | Порт Alertmanager на хосте |
| `TELEGRAM_BOT_TOKEN` | — | Токен Telegram-бота (от `@BotFather`). Секрет, в `.env.enc` |
| `TELEGRAM_CHAT_ID` | — | chat_id получателя уведомлений. Узнать: `https://api.telegram.org/bot<TOKEN>/getUpdates` |
| `DOMAIN` | `sololevelingms.duckdns.org` | Публичный домен для Caddy (см. [ADR-015](docs/adr/015-https-caddy.md)) |

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
| `PROFILES_GRPC_ADDR` | `profiles-service:50052` | Адрес Profiles Service для gRPC-вызовов |
| `EXERCISES_GRPC_ADDR` | `exercises-service:50053` | Адрес Exercises Service для gRPC-вызовов |
| `GATEWAY_TIMEOUT` | `10s` | Таймаут gRPC-запросов к Auth, Profiles и Exercises |
| `TRUSTED_PROXIES` | — | CIDR-список доверенных прокси через запятую. Пусто — X-Forwarded-For игнорируется |
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
| `GRPC_CLIENT_TLS_MODE` | `insecure` | Режим TLS для исходящих gRPC-соединений (`disabled` / `insecure` / `verify`) |

### `services/exercises/.env`

| Переменная | Значение по умолчанию | Назначение |
|------------|----------------------|------------|
| `GRPC_PORT` | `:50053` | Порт gRPC-сервера (внутри контейнера) |
| `HTTP_PORT` | `:8080` | Порт HTTP-сервера для /health, /ready и /metrics (внутри контейнера) |
| `ENABLE_REFLECTION` | `true` | gRPC reflection (для grpcurl). В проде — `false` |
| `POSTGRES_HOST` | `postgres-exercises` | Хост PostgreSQL внутри Docker-сети |
| `POSTGRES_PORT` | `5432` | Порт PostgreSQL |
| `POSTGRES_USER` | `test` | Пользователь БД |
| `POSTGRES_PASSWORD` | — | Пароль БД |
| `POSTGRES_DB` | `exercises_db` | Имя базы данных |
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

> ⚠️ Реальные `.env` **не коммитятся** в репозиторий (добавлены в `.gitignore`). Для запуска скопируйте `.env.example` в `.env` в каждой из пяти папок и заполните секреты.

## Architecture Decision Records

Ключевые архитектурные решения проекта зафиксированы в ADR — отдельные файлы в папке [`docs/adr/`](docs/adr/).

Список решений:

- [ADR-001: Профиль пользователя](docs/adr/001-profile.md) — lazy-create, nullable-поля, `profile_completed`
- [ADR-002: user_workout_score](docs/adr/002-user-workout-score.md) — почему не хранится в Profiles
- [ADR-003: Транзакции](docs/adr/003-transactions.md) — границы транзакций и sync-вызовы
- [ADR-004: Паттерны микросервисов](docs/adr/004-microservices-patterns.md) — что используем, что нет
- [ADR-005: Распределение портов](docs/adr/005-port-allocation.md) — диапазоны и смещение +10
- [ADR-006: Управление секретами](docs/adr/006-secrets-management.md) — SOPS + age
- [ADR-007: Известные проблемы и технический долг](docs/adr/007-known-issues.md) — что осталось до продакшена
- [ADR-008: Жизненный цикл упражнения](docs/adr/008-exercise-lifecycle.md) — immutable type, snapshot difficulty, soft delete
- [ADR-009: Карта зависимостей сервисов](docs/adr/009-service-dependencies.md) — кто кого зовёт, правила
- [ADR-010: Общие инфраструктурные пакеты](docs/adr/010-shared-infrastructure-packages.md) — что выносим в pkg, что оставляем локально
- [ADR-011: Naming conventions и Code style](docs/adr/011-naming-and-code-style.md) — соглашения по именованию и стилю
- [ADR-012: Оптимизация CI/CD](docs/adr/012-ci-cd-optimization.md) — matrix + docker cache
- [ADR-013: Требования к инфраструктуре](docs/adr/013-infrastructure-requirements.md) — сколько RAM/CPU/диска нужно на каждом этапе
- [ADR-014: Бэкапы и DR](docs/adr/014-backups-and-dr.md) — off-site бэкапы в Selectel S3
- [ADR-015: HTTPS через Caddy](docs/adr/015-https-caddy.md) — TLS-терминация, Let's Encrypt, DuckDNS
- [ADR-016: Сетевая безопасность](docs/adr/016-network-security.md) — bind на 127.0.0.1, ufw, fail2ban, insecure gRPC
- [ADR-017: Alerting](docs/adr/017-alerting.md) — Alertmanager + Telegram, правила алертов

Подробнее — в [docs/adr/README.md](docs/adr/README.md).
