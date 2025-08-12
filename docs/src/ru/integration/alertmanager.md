# AlertManager интеграция

## Обзор

AlertManager интеграция позволяет отправлять алерты Prometheus в LoRa mesh сеть через MQTT топики Meshtastic.

## Конфигурация AlertManager

Готовая конфигурация AlertManager доступна в файле [alertmanager.yml](../../stack/alertmanager/alertmanager.yml).

## Правила Prometheus

Готовые правила алертов доступны в файле [meshtastic-alerts.yml](../../stack/alertmanager/meshtastic-alerts.yml).

Включают алерты для:

- Обнаружения офлайн узлов
- Мониторинга уровня батареи
- Контроля температуры
- Проверки качества сигнала
- Доступности сервиса

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

| Канал        | Скорость     | Дальность    | Использование            |
|--------------|--------------|--------------|--------------------------|
| `ShortFast`  | Высокая      | Низкая       | Критические алерты       |
| `MediumFast` | Средняя      | Средняя      | Важные уведомления       |
| `LongFast`   | Низкая       | Высокая      | Обычные алерты           |
| `LongSlow`   | Очень низкая | Максимальная | Информационные сообщения |

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
  target_nodes: [ "admin001" ]

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

## TODO

- [ ] Добавить MQTT-специфичные метрики для мониторинга AlertManager интеграции

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
    "headers": {
      "Content-Type": "application/json"
    },
    "payload": "{\"alerts\":[{\"status\":\"firing\",\"labels\":{\"alertname\":\"NodeRED\"},\"annotations\":{\"summary\":\"{{payload}}\"}}]}"
  }
]
```