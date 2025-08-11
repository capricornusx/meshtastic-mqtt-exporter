# API Документация

## Endpoints

### Prometheus Metrics
- **GET** `/metrics` - Возвращает метрики в формате Prometheus
- **GET** `/health` - Health check endpoint

### AlertManager Webhook  
- **POST** `/alerts/webhook` - Принимает алерты от AlertManager

## OpenAPI Specification

Полная спецификация API доступна в файле [api/openapi.yaml](../../api/openapi.yaml).

Для просмотра используйте:
```bash
# Swagger UI
docker run -p 8080:8080 -e SWAGGER_JSON=/api/openapi.yaml -v $(pwd)/api:/api swaggerapi/swagger-ui

# Redoc
npx redoc-cli serve api/openapi.yaml
```

## Примеры использования

### Получение метрик
```bash
curl http://localhost:8100/metrics
```

Пример ответа:
```
# HELP meshtastic_battery_level_percent Battery level percentage
# TYPE meshtastic_battery_level_percent gauge
meshtastic_battery_level_percent{node_id="12345678",node_name="Node1"} 85

# HELP meshtastic_temperature_celsius Temperature in Celsius
# TYPE meshtastic_temperature_celsius gauge
meshtastic_temperature_celsius{node_id="12345678",node_name="Node1"} 23.5

# HELP meshtastic_humidity_percent Humidity percentage
# TYPE meshtastic_humidity_percent gauge
meshtastic_humidity_percent{node_id="12345678",node_name="Node1"} 45.2

# HELP meshtastic_pressure_hpa Barometric pressure in hPa
# TYPE meshtastic_pressure_hpa gauge
meshtastic_pressure_hpa{node_id="12345678",node_name="Node1"} 1013.25

# HELP meshtastic_node_last_seen_timestamp Last seen timestamp
# TYPE meshtastic_node_last_seen_timestamp gauge
meshtastic_node_last_seen_timestamp{node_id="12345678",node_name="Node1"} 1640995200
```

### Health Check
```bash
curl http://localhost:8100/health
```

Пример ответа:
```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "v1.0.0",
  "uptime": "2h30m15s",
  "metrics": {
    "active_nodes": 5,
    "total_messages": 1234,
    "last_message": "2024-01-15T10:29:45Z"
  }
}
```

### Отправка алерта
```bash
curl -X POST http://localhost:8100/alerts/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "alerts": [{
      "status": "firing",
      "labels": {
        "alertname": "NodeDown",
        "severity": "critical",
        "node_id": "12345678"
      },
      "annotations": {
        "summary": "Node is offline",
        "description": "Meshtastic node 12345678 has been offline for 10 minutes"
      }
    }]
  }'
```

## Метрики

### Основные метрики

| Метрика | Тип | Описание | Лейблы |
|---------|-----|----------|--------|
| `meshtastic_battery_level_percent` | gauge | Уровень батареи в процентах | `node_id`, `node_name` |
| `meshtastic_temperature_celsius` | gauge | Температура в градусах Цельсия | `node_id`, `node_name` |
| `meshtastic_humidity_percent` | gauge | Влажность в процентах | `node_id`, `node_name` |
| `meshtastic_pressure_hpa` | gauge | Барометрическое давление в гПа | `node_id`, `node_name` |
| `meshtastic_node_last_seen_timestamp` | gauge | Unix timestamp последней активности | `node_id`, `node_name` |

### Системные метрики

| Метрика | Тип | Описание |
|---------|-----|----------|
| `meshtastic_exporter_messages_total` | counter | Общее количество обработанных сообщений |
| `meshtastic_exporter_errors_total` | counter | Общее количество ошибок |
| `meshtastic_exporter_active_nodes` | gauge | Количество активных узлов |
| `meshtastic_exporter_uptime_seconds` | gauge | Время работы экспортера в секундах |

## AlertManager Webhook

### Формат запроса

```json
{
  "receiver": "lora-alerts",
  "status": "firing",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "NodeDown",
        "severity": "critical",
        "node_id": "12345678",
        "instance": "node1.example.com"
      },
      "annotations": {
        "summary": "Node is offline",
        "description": "Meshtastic node has been offline for 10 minutes"
      },
      "startsAt": "2024-01-15T10:20:00Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": "http://prometheus:9090/graph?g0.expr=up%7Bjob%3D%22meshtastic%22%7D+%3D%3D+0&g0.tab=1",
      "fingerprint": "b294c7c7c7c7c7c7"
    }
  ],
  "groupLabels": {
    "alertname": "NodeDown"
  },
  "commonLabels": {
    "alertname": "NodeDown",
    "severity": "critical"
  },
  "commonAnnotations": {
    "summary": "Node is offline"
  },
  "externalURL": "http://alertmanager:9093",
  "version": "4",
  "groupKey": "{}:{alertname=\"NodeDown\"}"
}
```

### Формат ответа

```json
{
  "status": "success",
  "message": "Alert sent to LoRa network",
  "details": {
    "channel": "LongFast",
    "mode": "broadcast",
    "target_nodes": ["ffffffff"],
    "message_sent": "🚨 NodeDown: Node is offline"
  }
}
```

## Коды ошибок

| Код | Описание |
|-----|----------|
| 200 | Успешный запрос |
| 400 | Неверный формат запроса |
| 404 | Endpoint не найден |
| 500 | Внутренняя ошибка сервера |
| 503 | Сервис недоступен |

## Примеры интеграции

### Prometheus конфигурация

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'meshtastic'
    static_configs:
      - targets: ['localhost:8100']
    scrape_interval: 30s
    metrics_path: /metrics
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Meshtastic Network",
    "panels": [
      {
        "title": "Battery Levels",
        "type": "stat",
        "targets": [
          {
            "expr": "meshtastic_battery_level_percent",
            "legendFormat": "{{node_name}}"
          }
        ]
      },
      {
        "title": "Temperature",
        "type": "graph",
        "targets": [
          {
            "expr": "meshtastic_temperature_celsius",
            "legendFormat": "{{node_name}}"
          }
        ]
      }
    ]
  }
}
```

### AlertManager правила

```yaml
# meshtastic.rules.yml
groups:
- name: meshtastic
  rules:
  - alert: NodeOffline
    expr: (time() - meshtastic_node_last_seen_timestamp) > 600
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Узел {{ $labels.node_name }} офлайн"
      
  - alert: LowBattery
    expr: meshtastic_battery_level_percent < 20
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "Низкий заряд батареи: {{ $labels.node_name }} ({{ $value }}%)"
```