# AlertManager интеграция

## Обзор

AlertManager интеграция позволяет отправлять алерты Prometheus в LoRa mesh сеть через MQTT топики Meshtastic.

## Конфигурация AlertManager

### alertmanager.yml

```yaml
global:
  smtp_smarthost: 'localhost:587'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'lora-alerts'
  routes:
  - match:
      severity: critical
    receiver: 'lora-critical'
  - match:
      severity: warning
    receiver: 'lora-warning'

receivers:
- name: 'lora-alerts'
  webhook_configs:
  - url: 'http://localhost:8080/alerts/webhook'
    send_resolved: true

- name: 'lora-critical'
  webhook_configs:
  - url: 'http://localhost:8080/alerts/webhook'
    send_resolved: true
    http_config:
      headers:
        X-Alert-Severity: critical

- name: 'lora-warning'
  webhook_configs:
  - url: 'http://localhost:8080/alerts/webhook'
    send_resolved: true
    http_config:
      headers:
        X-Alert-Severity: warning
```

## Правила Prometheus

### meshtastic.rules.yml

```yaml
groups:
- name: meshtastic.rules
  rules:
  - alert: NodeOffline
    expr: (time() - meshtastic_node_last_seen_timestamp) > 1200
    for: 5m
    labels:
      severity: warning
      service: meshtastic
    annotations:
      summary: "Узел Meshtastic {{ $labels.node_id }} офлайн"
      description: "Узел {{ $labels.node_name }} ({{ $labels.node_id }}) не отвечает более 20 минут"
      
  - alert: LowBattery
    expr: meshtastic_battery_level_percent < 20
    for: 2m
    labels:
      severity: critical
      service: meshtastic
    annotations:
      summary: "Низкий заряд батареи узла {{ $labels.node_id }}"
      description: "Заряд батареи узла {{ $labels.node_name }} составляет {{ $value }}%"
      
  - alert: HighTemperature
    expr: meshtastic_temperature_celsius > 60
    for: 5m
    labels:
      severity: warning
      service: meshtastic
    annotations:
      summary: "Высокая температура узла {{ $labels.node_id }}"
      description: "Температура узла {{ $labels.node_name }} составляет {{ $value }}°C"
      
  - alert: LowSignalQuality
    expr: meshtastic_snr_db < -10
    for: 10m
    labels:
      severity: info
      service: meshtastic
    annotations:
      summary: "Низкое качество сигнала узла {{ $labels.node_id }}"
      description: "SNR узла {{ $labels.node_name }} составляет {{ $value }} dB"
```

## Конфигурация экспортера

### Базовая конфигурация

```yaml
alertmanager:
  enabled: true
  http:
    port: 8080
    path: "/alerts/webhook"
  channel: "LongFast"
  mode: "broadcast"
  topics:
    broadcast: "msh/2/c/%s/!broadcast"
    direct: "msh/2/c/%s/!%s"
```

### Маршрутизация по severity

```yaml
alertmanager:
  enabled: true
  http:
    port: 8080
    path: "/alerts/webhook"
  
  # Маршрутизация по уровню важности
  routing:
    critical:
      channel: "ShortFast"    # Быстрая доставка для критических алертов
      mode: "broadcast"       # Отправить всем узлам
    warning:
      channel: "LongFast"     # Баланс дальности/скорости
      mode: "direct"          # Отправить только админам
      target_nodes:
        - "admin001"
        - "monitor02"
    info:
      channel: "LongSlow"     # Максимальная дальность для информации
      mode: "broadcast"       # Отправить всем узлам
```

## Форматы сообщений

### Стандартный формат

```json
{
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "NodeOffline",
        "severity": "warning",
        "node_id": "12345678"
      },
      "annotations": {
        "summary": "Узел Meshtastic 12345678 офлайн",
        "description": "Узел не отвечает более 20 минут"
      },
      "startsAt": "2024-01-15T10:30:00Z",
      "endsAt": "0001-01-01T00:00:00Z"
    }
  ]
}
```

### Кастомный формат для LoRa

Экспортер автоматически преобразует алерты в компактный формат для LoRa:

```
🚨 NodeOffline: Узел 12345678 офлайн
⚠️ LowBattery: Батарея узла 87654321 - 15%
✅ NodeOffline: Узел 12345678 восстановлен
```

## Каналы Meshtastic

### Типы каналов

| Канал | Скорость | Дальность | Использование |
|-------|----------|-----------|---------------|
| `ShortFast` | Высокая | Низкая | Критические алерты |
| `MediumFast` | Средняя | Средняя | Важные уведомления |
| `LongFast` | Низкая | Высокая | Обычные алерты |
| `LongSlow` | Очень низкая | Максимальная | Информационные сообщения |

### Выбор канала

```yaml
# Критические алерты - быстрая доставка
critical:
  channel: "ShortFast"
  mode: "broadcast"

# Предупреждения - баланс скорости и дальности  
warning:
  channel: "LongFast"
  mode: "direct"
  target_nodes: ["admin001"]

# Информация - максимальная дальность
info:
  channel: "LongSlow"
  mode: "broadcast"
```

## Режимы доставки

### Broadcast режим

Отправляет сообщения всем узлам в сети:

```yaml
alertmanager:
  mode: "broadcast"
  channel: "LongFast"
```

Топик: `msh/2/c/LongFast/!broadcast`

### Direct режим

Отправляет сообщения конкретным узлам:

```yaml
alertmanager:
  mode: "direct"
  channel: "ShortFast"
  target_nodes:
    - "admin001"
    - "monitor02"
```

Топики: 
- `msh/2/c/ShortFast/!admin001`
- `msh/2/c/ShortFast/!monitor02`

## Тестирование

### Тестовый алерт

```bash
curl -X POST http://localhost:8080/alerts/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "alerts": [{
      "status": "firing",
      "labels": {
        "alertname": "TestAlert",
        "severity": "warning"
      },
      "annotations": {
        "summary": "Тестовое сообщение алерта"
      }
    }]
  }'
```

### Проверка доставки

```bash
# Подписка на MQTT топики для проверки
mosquitto_sub -h localhost -t "msh/2/c/+/!+" -v

# Проверка логов
journalctl -u mqtt-exporter -f | grep alert
```

## Мониторинг AlertManager интеграции

### Метрики

```
# Количество обработанных алертов
meshtastic_alerts_processed_total{status="firing|resolved"}

# Количество отправленных сообщений
meshtastic_alerts_sent_total{channel="LongFast",mode="broadcast"}

# Ошибки обработки алертов
meshtastic_alerts_errors_total{error_type="parse|send"}
```

### Grafana панель

```json
{
  "title": "AlertManager Integration",
  "panels": [
    {
      "title": "Alerts Processed",
      "type": "stat",
      "targets": [
        {
          "expr": "rate(meshtastic_alerts_processed_total[5m])",
          "legendFormat": "{{status}}"
        }
      ]
    },
    {
      "title": "Alert Delivery",
      "type": "timeseries",
      "targets": [
        {
          "expr": "meshtastic_alerts_sent_total",
          "legendFormat": "{{channel}} - {{mode}}"
        }
      ]
    }
  ]
}
```

## Troubleshooting

### Алерты не доставляются

1. Проверьте конфигурацию AlertManager:
```bash
curl http://localhost:9093/api/v1/status
```

2. Проверьте webhook endpoint:
```bash
curl http://localhost:8080/alerts/webhook
```

3. Проверьте MQTT топики:
```bash
mosquitto_sub -h localhost -t "msh/2/c/+/!+" -v
```

### Проблемы с форматированием

1. Проверьте логи экспортера:
```bash
journalctl -u mqtt-exporter -f | grep alert
```

2. Тестируйте с простым алертом:
```bash
curl -X POST http://localhost:8080/alerts/webhook \
  -H "Content-Type: application/json" \
  -d '{"alerts":[{"status":"firing","labels":{"alertname":"Test"}}]}'
```

### Отладка маршрутизации

```yaml
# Включите отладочные логи
alertmanager:
  enabled: true
  debug: true
  http:
    port: 8080
    path: "/alerts/webhook"
```

## Примеры интеграции

### Home Assistant

```yaml
# configuration.yaml
automation:
  - alias: "Meshtastic Alert to LoRa"
    trigger:
      platform: state
      entity_id: binary_sensor.node_offline
      to: 'on'
    action:
      service: rest_command.send_lora_alert
      data:
        message: "Узел {{ trigger.entity_id }} офлайн"

rest_command:
  send_lora_alert:
    url: "http://localhost:8080/alerts/webhook"
    method: POST
    headers:
      Content-Type: "application/json"
    payload: >
      {
        "alerts": [{
          "status": "firing",
          "labels": {"alertname": "HomeAssistant"},
          "annotations": {"summary": "{{ message }}"}
        }]
      }
```

### Node-RED

```json
[
  {
    "id": "lora-alert",
    "type": "http request",
    "method": "POST",
    "url": "http://localhost:8080/alerts/webhook",
    "headers": {"Content-Type": "application/json"},
    "payload": "{\"alerts\":[{\"status\":\"firing\",\"labels\":{\"alertname\":\"NodeRED\"},\"annotations\":{\"summary\":\"{{payload}}\"}}]}"
  }
]
```