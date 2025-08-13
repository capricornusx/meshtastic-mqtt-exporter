# API

## Endpoints

- **GET** `/metrics` — Метрики Prometheus
- **GET** `/health` — Проверка состояния
- **POST** `/alerts/webhook` — Webhook для AlertManager

### Требования для AlertManager webhook

**Обязательно**: На gateway узле должен быть настроен канал с именем `mqtt` с включенным Downlink для получения алертов из MQTT.

Подробности настройки: [Meshtastic MQTT Integration](https://meshtastic.org/docs/software/integrations/mqtt/#json-downlink-to-instruct-a-node-to-send-a-message)

## Использование

```bash
# Метрики
curl http://localhost:8100/metrics

# Состояние
curl http://localhost:8100/health
```

Ответ health check:
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
| `meshtastic_rssi_dbm` | Мощность сигнала | `node_id`, `node_name` |
| `meshtastic_snr_db` | Отношение сигнал/шум | `node_id`, `node_name` |
| `meshtastic_node_last_seen_timestamp` | Последняя активность | `node_id`, `node_name` |