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

Готовые правила алертов доступны в [stack/alertmanager/meshtastic-alerts.yml](../stack/alertmanager/meshtastic-alerts.yml)

**Включают алерты для:**
- Обнаружения офлайн узлов
- Мониторинга уровня батареи  
- Контроля температуры
- Проверки качества сигнала
- Доступности сервиса

```yaml title="meshtastic-alerts.yml"
--8<-- "stack/alertmanager/meshtastic-alerts.yml"
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

### Grafana конфигурация

Готовые конфигурации Grafana доступны в [stack/grafana/](../stack/grafana/)

## Конфигурация Prometheus

Полная конфигурация Prometheus доступна в [stack/prometheus/prometheus.yml](../stack/prometheus/prometheus.yml)

**Критичные параметры:**
- `scrape_interval: 30s` — интервал сбора метрик
- `targets: ['localhost:8100']` — адрес MQTT экспортера
- `retention.time: 30d` — время хранения метрик

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