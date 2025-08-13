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
- todo position...

## Grafana

TODO: `docs/stack/grafana/dashboards/`

```promql
# Количество активных узлов
count(meshtastic_node_last_seen_timestamp)

# Средний уровень батареи
avg(meshtastic_battery_level_percent)

# Узлы с низким зарядом батареи
meshtastic_battery_level_percent < 20

# Узлы офлайн более 20 минут
(time() - meshtastic_node_last_seen_timestamp) > 1200

# Средняя температура по всем узлам
avg(meshtastic_temperature_celsius)

# Максимальная температура за последний час
max_over_time(meshtastic_temperature_celsius[1h])


# Топ-5 узлов с самым низким зарядом батареи
topk(5, meshtastic_battery_level_percent)

# Узлы с температурой выше среднего + 2 стандартных отклонения
meshtastic_temperature_celsius > (avg(meshtastic_temperature_celsius) + 2 * stddev(meshtastic_temperature_celsius))

# Скорость изменения заряда батареи (разряд/заряд)
rate(meshtastic_battery_level_percent[5m]) * 60

# Узлы с плохим качеством связи
avg_over_time(meshtastic_snr_db[1h]) < -10

# Количество сообщений в минуту по типам
rate(meshtastic_messages_processed_total[1m]) * 60
```

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
