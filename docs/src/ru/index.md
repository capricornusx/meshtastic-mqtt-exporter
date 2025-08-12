# Meshtastic MQTT Exporter

Экспорт телеметрии Meshtastic устройств в метрики Prometheus с интеграцией AlertManager для отправки алертов в LoRa сеть.

## Возможности

- **Встроенный MQTT брокер** с YAML конфигурацией
- **Prometheus метрики**: Батарея, температура, влажность, давление, качество сигнала
- **AlertManager интеграция**: Отправка алертов в LoRa mesh сеть
- **Персистентность состояния**: Сохранение/восстановление метрик между перезапусками

## Быстрый старт

```bash
# Скачать бинарник
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-linux-amd64

# Запустить embedded режим
./mqtt-exporter-linux-amd64 --config config.yaml

# Проверить метрики
curl http://localhost:8100/metrics
```

## Запуск

```bash
# Скачать и запустить
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-linux-amd64
chmod +x mqtt-exporter-linux-amd64
./mqtt-exporter-linux-amd64 --config config.yaml
```

## Метрики

- `meshtastic_battery_level_percent` — Уровень батареи
- `meshtastic_temperature_celsius` — Температура
- `meshtastic_humidity_percent` — Влажность  
- `meshtastic_pressure_hpa` — Барометрическое давление
- `meshtastic_node_last_seen_timestamp` — Время последней активности

## Документация

- **[Быстрый старт](quick-start.md)** — Установка и первый запуск
- **[Конфигурация](configuration.md)** — Настройка YAML файла
- **[API](api.md)** — REST API endpoints

## Благодарности

Построен с использованием [mochi-mqtt](https://github.com/mochi-mqtt/server) от [@mochi-co](https://github.com/mochi-co).