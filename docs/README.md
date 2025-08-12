# Документация Meshtastic MQTT Exporter

```
docs/
├── src/
│   ├── ru/
│   └── en/
├── site/
├── mkdocs.yml
├── requirements.txt
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## Быстрый старт

### Локальная разработка

```bash
# Запуск dev сервера (http://localhost:8000)
make serve

# Сборка документации
make build

# Проверка конфигурации
make check

# Очистка
make clean
```

### Использование Docker

```bash
# Запуск dev сервера
docker-compose -f docker-compose.yml up mkdocs

# Сборка документации
docker-compose -f docker-compose.yml run mkdocs-build
```

## Добавление контента

1. Создайте файл в `src/ru/` для русской версии
2. Создайте соответствующий файл в `src/en/` для английской версии
3. Добавьте ссылку в `nav` секцию `mkdocs.yml`
4. Проверьте сборку: `make check`

## Публикация

Документация автоматически публикуется на GitHub Pages при изменениях в папке `docs/` через GitHub Actions.

URL: https://capricornusx.github.io/meshtastic-mqtt-exporter/

