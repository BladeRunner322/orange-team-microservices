# Architecture Decision Records

Здесь фиксируются ключевые архитектурные решения проекта.

## Формат

Каждый ADR — отдельный файл в формате:

- **Контекст** — почему возник вопрос, что было в монолите
- **Решение** — что выбрали
- **Последствия** — что это даёт и какие минусы

## Список решений

- [ADR-001: Профиль пользователя](001-profile.md)
- [ADR-002: user_workout_score](002-user-workout-score.md)
- [ADR-003: Транзакции](003-transactions.md)
- [ADR-004: Паттерны микросервисов](004-microservices-patterns.md)
- [ADR-005: Распределение портов](005-port-allocation.md)
- [ADR-006: Управление секретами](006-secrets-management.md)
- [ADR-007: Известные проблемы и технический долг](007-known-issues.md)
- [ADR-008: Жизненный цикл упражнения](008-exercise-lifecycle.md)
- [ADR-009: Карта зависимостей сервисов](009-service-dependencies.md)
- [ADR-010: Общие инфраструктурные пакеты](010-shared-infrastructure-packages.md)
- [ADR-011: Naming conventions и Code style](011-naming-and-code-style.md)
- [ADR-012: Оптимизация CI/CD](012-ci-cd-optimization.md)
- [ADR-013: Требования к инфраструктуре](013-infrastructure-requirements.md)
- [ADR-014: Бэкапы и DR](014-backups-and-dr.md)
- [ADR-015: HTTPS через Caddy](015-https-caddy.md)
- [ADR-016: Сетевая безопасность](016-network-security.md)
- [ADR-017: Alerting (Alertmanager + Telegram)](017-alerting.md)
- [ADR-018: Retention метрик и логов](018-storage-retention.md)

## Как добавить

1. Создай файл `NNN-краткое-название.md`
2. Используй шаблон ниже
3. Добавь ссылку в этот файл

## Шаблон

\`\`\`markdown
# ADR-NNN: Название

## Контекст
Что было, какие ограничения.

## Решение
Что выбрали.

## Последствия
Что получаем и чем платим.
\`\`\`