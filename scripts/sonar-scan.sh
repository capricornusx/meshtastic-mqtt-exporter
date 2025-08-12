#!/bin/env bash

set -e

# Загрузка переменных из .env
if [ -f .env ]; then
    export $(cat .env | xargs)
fi

echo "Проверка наличия отчетов..."
if [ ! -f "coverage.out" ]; then
    echo "Ошибка: coverage.out не найден. Запустите 'make coverage'"
    exit 1
fi

echo "Запуск SonarQube Scanner..."
docker run --rm \
    --network host \
    -e SONAR_TOKEN="${SONAR_TOKEN}" \
    -v "$(pwd):/usr/src" \
    -v "$(pwd)/.git:/usr/src/.git:ro" \
    -w /usr/src \
    sonarsource/sonar-scanner-cli:latest \
    -Dsonar.host.url=http://192.168.1.77:9000

echo "Анализ завершен. Результаты доступны на http://192.168.1.77:9000"