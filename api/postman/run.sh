#!/usr/bin/env bash
set -euo pipefail

# Прогоняет все Postman-коллекции по порядку против Docker-стека.
# Требует, чтобы стек был запущен: task docker-up
#
# Использование:
#   ./api/postman/run.sh              # весь набор
#   ./api/postman/run.sh habits       # только Habits (Setup всё равно прогонится)
#   ./api/postman/run.sh setup habits # выборочно, в указанном порядке

ENV="api/postman/env.json"

# Первый запуск: env.json ещё нет — создаём из шаблона.
if [ ! -f "$ENV" ]; then
    echo "→ Creating $ENV from env.example.json"
    cp "api/postman/env.example.json" "$ENV"
fi

# Полный порядок прогона. Setup всегда идёт первым — он пишет токены в env.
# rate_limit в конце: он создаёт нагрузку на /login, остальным тестам это
# не нужно, а после остальных он уже ничего не сломает.
ALL=(setup auth profiles exercises habits rate_limit)

# Если аргументы переданы — берём их, иначе весь набор.
if [ "$#" -gt 0 ]; then
    ORDER=("$@")
else
    ORDER=("${ALL[@]}")
fi

# Проверка, что стек жив (Gateway отвечает на /health).
if ! curl -sf -o /dev/null "http://localhost:8081/health"; then
    echo "❌ Gateway недоступен на http://localhost:8081"
    echo "   Запусти: task docker-up"
    exit 1
fi

# Прогон по порядку.
for name in "${ORDER[@]}"; do
    file="api/postman/${name}_collection.json"

    if [ ! -f "$file" ]; then
        echo "❌ Не найдено: $file"
        exit 1
    fi

    echo ""
    echo "→ ${name}"
    npx newman run "$file" -e "$ENV" --export-environment "$ENV" --reporters cli
done

echo ""
echo "✅ Все коллекции прошли"