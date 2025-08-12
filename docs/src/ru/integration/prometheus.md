# Prometheus интеграция

## Конфигурация Prometheus

Готовая конфигурация Prometheus доступна в файле [prometheus.yml](../stack/prometheus/prometheus.yml).

## Доступные метрики

### Основные метрики узлов

| Метрика                               | Тип   | Описание                                    | Лейблы                 |
|---------------------------------------|-------|---------------------------------------------|------------------------|
| `meshtastic_battery_level_percent`    | gauge | Уровень заряда батареи (%)                  | `node_id`, `node_name` |
| `meshtastic_temperature_celsius`      | gauge | Температура (°C)                            | `node_id`, `node_name` |
| `meshtastic_humidity_percent`         | gauge | Влажность (%)                               | `node_id`, `node_name` |
| `meshtastic_pressure_hpa`             | gauge | Давление (hPa)                              | `node_id`, `node_name` |
| `meshtastic_node_last_seen_timestamp` | gauge | Время последней активности (Unix timestamp) | `node_id`, `node_name` |

### Метрики качества сигнала

| Метрика                | Тип   | Описание                  | Лейблы                              |
|------------------------|-------|---------------------------|-------------------------------------|
| `meshtastic_snr_db`    | gauge | Отношение сигнал/шум (dB) | `node_id`, `node_name`, `from_node` |
| `meshtastic_rssi_dbm`  | gauge | Мощность сигнала (dBm)    | `node_id`, `node_name`, `from_node` |
| `meshtastic_hop_limit` | gauge | Лимит переходов           | `node_id`, `node_name`              |

### Системные метрики

| Метрика                               | Тип     | Описание                                | Лейблы                  |
|---------------------------------------|---------|-----------------------------------------|-------------------------|
| `meshtastic_messages_processed_total` | counter | Общее количество обработанных сообщений | `topic`, `message_type` |
| `meshtastic_nodes_active`             | gauge   | Количество активных узлов               | -                       |
| `meshtastic_exporter_uptime_seconds`  | gauge   | Время работы экспортера (секунды)       | -                       |

## Правила алертов

### Базовые правила

```yaml
# /etc/prometheus/rules/meshtastic-basic.yml
groups:
  - name: meshtastic.basic
    rules:
      - alert: MeshtasticExporterDown
        expr: up{job="meshtastic-exporter"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Meshtastic экспортер недоступен"
          description: "Экспортер не отвечает более 1 минуты"

      - alert: NoActiveNodes
        expr: meshtastic_nodes_active == 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Нет активных Meshtastic узлов"
          description: "В сети нет активных узлов более 5 минут"
```

### Правила для узлов

```yaml
# /etc/prometheus/rules/meshtastic-nodes.yml
groups:
  - name: meshtastic.nodes
    rules:
      - alert: NodeOffline
        expr: (time() - meshtastic_node_last_seen_timestamp) > 1200
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Узел {{ $labels.node_name }} офлайн"
          description: "Узел {{ $labels.node_name }} ({{ $labels.node_id }}) не отвечает {{ $value | humanizeDuration }}"

      - alert: CriticalBatteryLevel
        expr: meshtastic_battery_level_percent < 10
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Критически низкий заряд батареи"
          description: "Заряд батареи узла {{ $labels.node_name }} составляет {{ $value }}%"

      - alert: LowBatteryLevel
        expr: meshtastic_battery_level_percent < 20
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Низкий заряд батареи"
          description: "Заряд батареи узла {{ $labels.node_name }} составляет {{ $value }}%"

      - alert: HighTemperature
        expr: meshtastic_temperature_celsius > 60
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Высокая температура узла"
          description: "Температура узла {{ $labels.node_name }} составляет {{ $value }}°C"

      - alert: ExtremeTemperature
        expr: meshtastic_temperature_celsius > 80 or meshtastic_temperature_celsius < -20
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Экстремальная температура узла"
          description: "Температура узла {{ $labels.node_name }} составляет {{ $value }}°C"
```

### Правила качества сигнала

```yaml
# /etc/prometheus/rules/meshtastic-signal.yml
groups:
  - name: meshtastic.signal
    rules:
      - alert: PoorSignalQuality
        expr: meshtastic_snr_db < -10
        for: 15m
        labels:
          severity: info
        annotations:
          summary: "Плохое качество сигнала"
          description: "SNR между {{ $labels.node_name }} и {{ $labels.from_node }} составляет {{ $value }} dB"

      - alert: VeryPoorSignalQuality
        expr: meshtastic_snr_db < -15
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Очень плохое качество сигнала"
          description: "SNR между {{ $labels.node_name }} и {{ $labels.from_node }} составляет {{ $value }} dB"

      - alert: WeakSignalStrength
        expr: meshtastic_rssi_dbm < -120
        for: 10m
        labels:
          severity: info
        annotations:
          summary: "Слабый сигнал"
          description: "RSSI между {{ $labels.node_name }} и {{ $labels.from_node }} составляет {{ $value }} dBm"
```

## Запросы PromQL

### Основные запросы

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
```

### Продвинутые запросы

```promql
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

## Grafana интеграция

### Datasource конфигурация

```yaml
# grafana/provisioning/datasources/prometheus.yml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
```

### Основные панели

```json
{
  "dashboard": {
    "title": "Meshtastic Network Overview",
    "panels": [
      {
        "title": "Active Nodes",
        "type": "stat",
        "targets": [
          {
            "expr": "count(meshtastic_node_last_seen_timestamp)",
            "legendFormat": "Active Nodes"
          }
        ]
      },
      {
        "title": "Average Battery Level",
        "type": "stat",
        "targets": [
          {
            "expr": "avg(meshtastic_battery_level_percent)",
            "legendFormat": "Average Battery %"
          }
        ]
      },
      {
        "title": "Battery Levels by Node",
        "type": "bargauge",
        "targets": [
          {
            "expr": "meshtastic_battery_level_percent",
            "legendFormat": "{{node_name}}"
          }
        ]
      },
      {
        "title": "Temperature Trends",
        "type": "timeseries",
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

## Recording Rules

### Агрегированные метрики

```yaml
# /etc/prometheus/rules/meshtastic-recording.yml
groups:
  - name: meshtastic.recording
    interval: 30s
    rules:
      - record: meshtastic:nodes_active
        expr: count(meshtastic_node_last_seen_timestamp)

      - record: meshtastic:battery_avg
        expr: avg(meshtastic_battery_level_percent)

      - record: meshtastic:battery_min
        expr: min(meshtastic_battery_level_percent)

      - record: meshtastic:temperature_avg
        expr: avg(meshtastic_temperature_celsius)

      - record: meshtastic:temperature_max
        expr: max(meshtastic_temperature_celsius)

      - record: meshtastic:messages_rate_5m
        expr: rate(meshtastic_messages_processed_total[5m])

      - record: meshtastic:nodes_offline
        expr: count((time() - meshtastic_node_last_seen_timestamp) > 1200)
```

## Retention и Storage

### Конфигурация хранения

```yaml
# prometheus.yml
global:
  scrape_interval: 30s
  evaluation_interval: 30s

# Retention настройки
storage:
  tsdb:
    retention.time: 30d
    retention.size: 10GB
    wal-compression: true

# Удаленное хранение (опционально)
remote_write:
  - url: "https://prometheus-remote-write-endpoint"
    queue_config:
      max_samples_per_send: 1000
      max_shards: 200
```

## Мониторинг производительности

### Метрики экспортера

```promql
# Время обработки запросов
prometheus_http_request_duration_seconds{job="meshtastic-exporter"}

# Использование памяти
process_resident_memory_bytes{job="meshtastic-exporter"}

# Количество горутин
go_goroutines{job="meshtastic-exporter"}

# Сборка мусора
rate(go_gc_duration_seconds_count{job="meshtastic-exporter"}[5m])
```

### Оптимизация запросов

```promql
# Используйте recording rules для часто используемых запросов
meshtastic:nodes_active

# Ограничивайте временные диапазоны
meshtastic_battery_level_percent[1h]

# Используйте агрегацию для больших наборов данных
avg by (node_id) (meshtastic_temperature_celsius)
```

## Troubleshooting

### Отсутствующие метрики

1. Проверьте статус экспортера:

```bash
curl http://localhost:8100/health
```

2. Проверьте доступность метрик:

```bash
curl http://localhost:8100/metrics | grep meshtastic
```

3. Проверьте конфигурацию Prometheus:

```bash
curl http://localhost:9090/api/v1/targets
```

### Проблемы с производительностью

1. Мониторинг времени scrape:

```promql
prometheus_target_scrape_duration_seconds{job="meshtastic-exporter"}
```

2. Проверка размера ответа:

```promql
prometheus_target_scrape_samples_scraped{job="meshtastic-exporter"}
```

3. Оптимизация интервала сбора:

```yaml
scrape_configs:
  - job_name: 'meshtastic-exporter'
    scrape_interval: 60s  # Увеличить интервал
```