
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
git clone https://github.com/BladeRunner322/orange-team-microservices cd orange-team-microservices
```

2. Настройте переменные окружения

Скопируйте .env.example в .env и заполните значения:
```
cp .env.example .env
```

Обязательно укажите:

- JWT_SECRET — секретный ключ для JWT (минимум 32 байта).

- POSTGRES_PASSWORD — пароль для БД.

3. Запустите всё окружение (PostgreSQL + миграции + сервисы)
```
task docker-up
```

Это поднимет:

- PostgreSQL (порт 5432)

- Миграции (создание таблиц)

- Auth-сервис (gRPC, порт 50051)

4. Проверьте, что сервис работает
```
grpcurl -plaintext localhost:50051 list
```

5. Протестируйте регистрацию и логин
```
grpcurl -plaintext -d '{"email":"test@example.com","password":"password123","full_name":"Test User"}' localhost:50051 auth.AuthService/Register
```
```
grpcurl -plaintext -d '{"email":"test@example.com","password":"password123"}' localhost:50051 auth.AuthService/Login
```

## Структура проекта
```
├── api/                      # gRPC-контракты (.proto)
│   └── auth/
│       └── auth.proto
├── internal/                 # сгенерированный код из proto
│   └── gen/
├── migrations/               # SQL-миграции
├── pkg/                      # общие инфраструктурные пакеты
│   ├── logger/               # структурированное логирование
│   └── postgres/             # работа с PostgreSQL (pgx)
├── services/                 # микросервисы
│   └── auth/                 # сервис аутентификации
│       ├── cmd/              # точка входа
│       ├── config/           # конфигурация
│       ├── internal/         # приватный код
│       │   ├── bootstrap/    # сборка зависимостей
│       │   ├── domain/       # DDD: сущности, value objects, ошибки
│       │   ├── application/  # use cases и порты
│       │   ├── infrastructure/ # реализации (JWT, PostgreSQL)
│       │   └── interfaces/   # gRPC-адаптеры
│       ├── Dockerfile
│       └── .env.example
├── docker-compose.yaml
├── Taskfile.yml
├── go.mod
└── README.md
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
task migrate-create -- create_users_table
```

### Применить миграции
```
task migrate-up
```

### Откатить последнюю
```
task migrate-down -- 1
```

### Показать текущую версию
```
task migrate-version
```

## Переменные окружения

Обязательные переменные:

- JWT_SECRET	(Секрет для подписи JWT)
- POSTGRES_PASSWORD	(Пароль для PostgreSQL)
- POSTGRES_USER	(Имя пользователя БД)
- POSTGRES_DB	(Имя базы данных)

Полный список — в .env.example.
