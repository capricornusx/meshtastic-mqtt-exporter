# Интеграция

## Prometheus

### Конфигурация

```yaml
scrape_configs:
  - job_name: 'meshtastic'
    static_configs:
      - targets: ['localhost:8100']
    scrape_interval: 30s
```

### Правила алертов

```yaml
groups:
  - name: meshtastic
    rules:
      - alert: MeshtasticNodeOffline
        expr: time() - meshtastic_node_last_seen_timestamp > 3600
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Узел {{ $labels.node_name }} офлайн"

      - alert: MeshtasticLowBattery
        expr: meshtastic_battery_level_percent < 20
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Низкий заряд батареи: {{ $labels.node_name }}"
```

## AlertManager

### Требования

**Обязательно**: На gateway узле должен быть настроен канал с именем `mqtt` с включенным Downlink для получения алертов из MQTT.

Подробности настройки: [Meshtastic MQTT Integration](https://meshtastic.org/docs/software/integrations/mqtt/#json-downlink-to-instruct-a-node-to-send-a-message)

### Webhook конфигурация

```yaml
route:
  group_by: ['alertname']
  receiver: 'meshtastic-webhook'

receivers:
  - name: 'meshtastic-webhook'
    webhook_configs:
      - url: 'http://localhost:8100/alerts/webhook'
        send_resolved: true
```

### Режимы доставки

**Broadcast** — всем узлам:
```yaml
alertmanager:
  mode: "broadcast"
  channel: "LongFast"
```

**Direct** — конкретным узлам:
```yaml
alertmanager:
  mode: "direct"
  target_nodes: ["admin001", "monitor02"]
```

## Meshtastic устройства

### Настройка MQTT

```bash
meshtastic --set mqtt.enabled true
meshtastic --set mqtt.address your-server.com
meshtastic --set mqtt.username meshtastic
meshtastic --set mqtt.password password
```

### Проверка топиков

Устройства публикуют в:
- `msh/2/c/LongFast/!<node_id>` — сообщения
- `msh/2/e/LongFast/!<node_id>` — телеметрия

## Grafana

### Готовые дашборды

Доступны в `docs/stack/grafana/dashboards/`

### Основные панели

- Карта узлов сети
- Уровень батареи
- Температура/влажность
- Качество сигнала (RSSI/SNR)
- Активность узлов

## Hook интеграция

Для интеграции с существующим mochi-mqtt:

```go
import "github.com/capricornusx/meshtastic-mqtt-exporter/internal/hooks"

hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
    ServerAddr:  ":8100",
    TopicPrefix: "msh/",
    MetricsTTL:  30 * time.Minute,
}, factory)

server.AddHook(hook, nil)
```
