# Docker развертывание

## Быстрый запуск

```bash
# Простой запуск
docker run -p 1883:1883 -p 8100:8100 \
  ghcr.io/capricornusx/meshtastic-mqtt-exporter:latest

# С конфигурационным файлом
docker run -p 1883:1883 -p 8100:8100 \
  -v $(pwd)/config.yaml:/config.yaml \
  ghcr.io/capricornusx/meshtastic-mqtt-exporter:latest --config /config.yaml
```

## Docker Compose

### Минимальная конфигурация

```yaml
# docker-compose.yml
version: '3.8'

services:
  mqtt-exporter:
    image: ghcr.io/capricornusx/meshtastic-mqtt-exporter:latest
    ports:
      - "1883:1883"
      - "8100:8100"
      - "8080:8080"
    volumes:
      - ./config.yaml:/config.yaml
    command: --config /config.yaml
    restart: unless-stopped
```

### Полный стек мониторинга

```yaml title="docker-compose.yml"
--8<-- "stack/docker-compose.yml"
```

## Конфигурационные файлы

Полный стек мониторинга с готовыми конфигурациями доступен в [stack/](../stack/):

- [stack/alertmanager/alertmanager.yml](../stack/alertmanager/alertmanager.yml) — конфигурация AlertManager
- [stack/prometheus/prometheus.yml](../stack/prometheus/prometheus.yml) — конфигурация Prometheus
- [stack/docker-compose.yml](../stack/docker-compose.yml) — Docker Compose стек

**Критичные параметры:**
- `webhook_configs.url` — должен указывать на `http://mqtt-exporter:8080/alerts/webhook`
- `send_resolved: true` — для получения уведомлений о восстановлении

## Управление контейнерами

### Запуск

```bash
# Запуск всех сервисов
docker-compose up -d

# Запуск только MQTT экспортера
docker-compose up -d mqtt-exporter

# Запуск с мониторингом
docker-compose --profile monitoring up -d
```

### Мониторинг

```bash
# Просмотр логов
docker-compose logs -f mqtt-exporter

# Статус сервисов
docker-compose ps

# Использование ресурсов
docker stats
```

### Обновление

```bash
# Обновление образов
docker-compose pull

# Перезапуск с новыми образами
docker-compose up -d --force-recreate
```

## Персистентность данных

### Volumes

```yaml
services:
  mqtt-exporter:
    volumes:
      - mqtt_data:/data
      - ./config.yaml:/config.yaml:ro
      - ./logs:/var/log/mqtt-exporter

volumes:
  mqtt_data:
    driver: local
```

### Bind mounts

```yaml
services:
  mqtt-exporter:
    volumes:
      - ./data:/data
      - ./config.yaml:/config.yaml:ro
      - ./logs:/var/log/mqtt-exporter
```

## Безопасность

### Пользователь без привилегий

```dockerfile
# Dockerfile
FROM alpine:latest
RUN addgroup -g 1001 mqtt && adduser -D -u 1001 -G mqtt mqtt
USER mqtt
```

### Secrets

```yaml
services:
  mqtt-exporter:
    secrets:
      - mqtt_password
    environment:
      - MQTT_PASSWORD_FILE=/run/secrets/mqtt_password

secrets:
  mqtt_password:
    file: ./secrets/mqtt_password.txt
```

## Troubleshooting

### Проверка подключения

```bash
# Проверка портов
docker-compose exec mqtt-exporter netstat -tlnp

# Проверка метрик
curl http://localhost:8100/metrics

# Проверка health
curl http://localhost:8100/health
```

### Логи

```bash
# Все логи
docker-compose logs

# Логи конкретного сервиса
docker-compose logs mqtt-exporter

# Следить за логами
docker-compose logs -f mqtt-exporter
```