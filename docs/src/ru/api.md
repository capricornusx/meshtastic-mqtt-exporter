# API

## Endpoints

- **GET** `/metrics` — Метрики Prometheus
- **GET** `/health` — Проверка состояния
- **POST** `/alerts/webhook` — Webhook для AlertManager

## Использование

### Метрики

```bash
curl http://localhost:8100/metrics
```

### Health Check

```bash
curl http://localhost:8100/health
```

Ответ:
```json
{
  "status": "ok",
  "uptime": "2h30m15s",
  "active_nodes": 5
}
```

## Метрики

| Метрика | Описание | Лейблы |
|---------|----------|--------|
| `meshtastic_battery_level_percent` | Уровень батареи | `node_id`, `node_name` |
| `meshtastic_temperature_celsius` | Температура | `node_id`, `node_name` |
| `meshtastic_humidity_percent` | Влажность | `node_id`, `node_name` |
| `meshtastic_pressure_hpa` | Давление | `node_id`, `node_name` |
| `meshtastic_node_last_seen_timestamp` | Последняя активность | `node_id`, `node_name` |

## Prometheus интеграция

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'meshtastic'
    static_configs:
      - targets: ['localhost:8100']
    scrape_interval: 30s
```

## AlertManager правила

Полный набор правил доступен в файле [`docs/alertmanager/meshtastic-alerts.yml`](https://github.com/capricornusx/meshtastic-mqtt-exporter/blob/main/docs/alertmanager/meshtastic-alerts.yml).

Включает алерты для:
- Офлайн узлов
- Низкого заряда батареи  
- Высокой температуры
- Слабого сигнала
- Недоступности экспортера