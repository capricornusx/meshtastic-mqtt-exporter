# Docker 


```bash
# Простой запуск
docker run -p 1883:1883 -p 8100:8100 \
  ghcr.io/capricornusx/meshtastic-mqtt-exporter:latest

# С конфигурационным файлом
docker run -p 1883:1883 -p 8100:8100 \
  -v $(pwd)/config.yaml:/config.yaml \
  ghcr.io/capricornusx/meshtastic-mqtt-exporter:latest --config /config.yaml
```

### Docker Compose

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

### Конфигурационные файлы

Полный стек мониторинга с готовыми конфигурациями доступен в [stack/](../stack/):

- [stack/alertmanager/alertmanager.yml](../stack/alertmanager/alertmanager.yml) — конфигурация AlertManager
- [stack/prometheus/prometheus.yml](../stack/prometheus/prometheus.yml) — конфигурация Prometheus
- [stack/docker-compose.yml](../stack/docker-compose.yml) — Docker Compose стек

**Критичные параметры:**
- `webhook_configs.url` — должен указывать на `http://mqtt-exporter:8080/alerts/webhook`
- `send_resolved: true` — для получения уведомлений о восстановлении

### Запуск стека

```bash
cd docs/stack
docker-compose up -d
```

**Доступные сервисы:**
- MQTT Exporter: http://localhost:8100/metrics
- Prometheus: http://localhost:9090
- AlertManager: http://localhost:9093
- Grafana: http://localhost:3000 (admin/admin123)

