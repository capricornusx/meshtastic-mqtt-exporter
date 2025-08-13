# Meshtastic MQTT Exporter

Экспорт телеметрии Meshtastic устройств в Prometheus с интеграцией AlertManager.

## Возможности

- **Встроенный MQTT брокер** на основе mochi-mqtt
- **Prometheus метрики**: Батарея, температура, влажность, давление, сигнал
- **AlertManager интеграция**: Отправка алертов в LoRa сеть
- **Персистентность**: Сохранение метрик между перезапусками

## Быстрый старт

```bash
# Скачать и запустить
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-linux-amd64
wget https://raw.githubusercontent.com/capricornusx/meshtastic-mqtt-exporter/main/config.yaml
./mqtt-exporter-linux-amd64 --config config.yaml

# Проверить
curl http://localhost:8100/metrics
```

## Метрики

- `meshtastic_battery_level_percent` — Уровень батареи
- `meshtastic_temperature_celsius` — Температура
- `meshtastic_humidity_percent` — Влажность
- `meshtastic_pressure_hpa` — Давление
- `meshtastic_rssi_dbm` — Мощность сигнала
- `meshtastic_node_last_seen_timestamp` — Последняя активность
