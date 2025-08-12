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
    volumes:
      - ./config.yaml:/config.yaml
    command: --config /config.yaml
    restart: unless-stopped
```

### Полный стек мониторинга

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
      - mqtt_data:/data
    command: --config /config.yaml
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - ./rules:/etc/prometheus/rules
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--web.enable-lifecycle'
    restart: unless-stopped

  alertmanager:
    image: prom/alertmanager:latest
    ports:
      - "9093:9093"
    volumes:
      - ./alertmanager.yml:/etc/alertmanager/alertmanager.yml
      - alertmanager_data:/alertmanager
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./grafana/datasources:/etc/grafana/provisioning/datasources
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin123
    restart: unless-stopped

volumes:
  mqtt_data:
  prometheus_data:
  alertmanager_data:
  grafana_data:
```

## Конфигурационные файлы

### Prometheus

Полный стек мониторинга доступен в [docs/stack/](../stack/).

### AlertManager

```yaml
# alertmanager.yml
global:
  smtp_smarthost: 'localhost:587'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'lora-alerts'

receivers:
- name: 'lora-alerts'
  webhook_configs:
  - url: 'http://mqtt-exporter:8080/alerts/webhook'
    send_resolved: true
```

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