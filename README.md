
# Orange Team Microservices

Монорепозиторий с микросервисами на Go, построенными по чистой архитектуре и DDD.

## Требования

- Go 1.26.7 или выше
- Docker и Docker Compose
- Task (для управления задачами)
- grpcurl (для тестирования gRPC)

Установка Task:
```
go install github.com/go-task/task/v3/cmd/task@latest
```

## Быстрый старт

1. Склонируйте репозиторий
```
git clone https://github.com/BladeRunner322/orange-team-microservices
cd orange-team-microservices
```
2. Настройте переменные окружения

Скопируйте .env.example в .env и заполните значения:
```
cp .env.example .env
```
Обязательно укажите:

- JWT_SECRET — секретный ключ для JWT (минимум 32 байта).

- POSTGRES_PASSWORD — пароль для БД.

3. Сгенерируйте TLS-сертификаты для разработки

Auth Service использует gRPC с TLS. Для локальной разработки создайте самоподписанный сертификат:
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

5. Проверьте, что сервис работает
```
grpcurl -insecure localhost:50051 list
```

6. Протестируйте регистрацию и логин
```
grpcurl -insecure -d '{"email":"test@example.com","password":"password123","full_name":"Test User"}' localhost:50051 auth.AuthService/Register
```
```
grpcurl -insecure -d '{"email":"test@example.com","password":"password123"}' localhost:50051 auth.AuthService/Login
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
│   └── leaderboard/
│       └── leaderboard.proto
│
├── internal/
│   └── gen/                          # сгенерированный код из proto
│       └── api/
│           ├── auth/
│           ├── users/
│           ├── exercises/
│           ├── habits/
│           ├── workouts/
│           └── leaderboard/
│
├── pkg/                              # общие пакеты
│   ├── logger/
│   ├── postgres/
│   ├── metrics/
│   └── interceptors/
│
├── services/                         # микросервисы
│   ├── gateway/                      # API Gateway (НОВЫЙ)
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── config/
│   │   │   └── config.go
│   │   ├── internal/
│   │   │   ├── handlers/             # HTTP-хендлеры
│   │   │   ├── middleware/           # аутентификация, логирование
│   │   │   └── clients/              # gRPC-клиенты к сервисам
│   │   └── Dockerfile
│   │
│   ├── auth/                         # ✅ ГОТОВ
│   │   ├── cmd/
│   │   ├── config/
│   │   ├── internal/
│   │   │   ├── bootstrap/
│   │   │   ├── domain/
│   │   │   ├── application/
│   │   │   ├── infrastructure/
│   │   │   └── interfaces/
│   │   ├── migrations/               # auth_db
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
├── docker-compose.yml                # все контейнеры
├── Taskfile.yml                      # задачи для разработки
├── go.mod
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
## Разработка

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
    task run-auth
    ```

  - Сборка Docker-образа:
    ```
    task auth-build
    ```

  - Запуск в Docker Compose:
    ```
    task auth-up
    ```

  - Просмотр логов:
    ```
    task auth-logs
    ```

#### Healthcheck

Auth-сервис предоставляет HTTP-эндпоинт для проверки состояния:

- **Порт:** 8080
- **Эндпоинт:** `/health`
- **Ответ:** `{"status":"ok","service":"auth"}`

Проверка:

```
curl http://localhost:8080/health
```
## Мониторинг и логирование

В проекте настроен полный стек для мониторинга и логирования:

- **Prometheus** — сбор метрик
- **Loki** — агрегация логов
- **Grafana** — визуализация

### Метрики (Prometheus)

Auth-сервис предоставляет эндпоинт с метриками:

- **Порт:** 8080
- **Эндпоинт:** `/metrics`

Проверка:

```
curl http://localhost:8080/metrics
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

Предварительно настроены источники данных:

- **Prometheus** — для метрик
- **Loki** — для логов

Пример дашборда с метриками и логами уже создан.

Для доступа к интерфейсу Grafana используйте `http://localhost:3000`.

### Запуск мониторинга

Все команды доступны через Taskfile:

```
task docker-up        # поднимает всё (приложение + мониторинг + логи)
task monitoring-up    # только Prometheus + Grafana
task logging-up       # только Loki + Promtail
```

### Проверка работы

```
curl http://localhost:8080/metrics   # метрики
curl http://localhost:3100/ready     # Loki готов
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

## Управление миграциями

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

Обязательные переменные:

- JWT_SECRET	(Секрет для подписи JWT)
- POSTGRES_PASSWORD	(Пароль для PostgreSQL)
- POSTGRES_USER	(Имя пользователя БД)
- POSTGRES_DB	(Имя базы данных)

Полный список — в .env.example.
