# Meshtastic MQTT Exporter

[![Build Status](https://github.com/capricornusx/meshtastic-mqtt-exporter/workflows/Build%20and%20Test/badge.svg)](https://github.com/capricornusx/meshtastic-mqtt-exporter/actions)
[![codecov](https://codecov.io/gh/capricornusx/meshtastic-mqtt-exporter/graph/badge.svg?token=P0409HCBFS)](https://codecov.io/gh/capricornusx/meshtastic-mqtt-exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/capricornusx/meshtastic-mqtt-exporter)](https://goreportcard.com/report/github.com/capricornusx/meshtastic-mqtt-exporter)

Экспорт телеметрии Meshtastic устройств в метрики Prometheus с интеграцией AlertManager для отправки алертов в LoRa сеть.

## Возможности

- **Встроенный MQTT брокер** с YAML конфигурацией
- **TLS поддержка** - TCP (1883) + TLS (8883) порты одновременно
- **Prometheus метрики**: Батарея, температура, влажность, давление, качество сигнала
- **AlertManager интеграция**: Отправка алертов в LoRa mesh сеть
- **Персистентность состояния**: Сохранение/восстановление метрик между перезапусками

## Быстрый старт

### Docker Compose (полный стек)

```bash
# Полный стек мониторинга
cd docs/stack
docker-compose up -d
```

### Отдельный бинарник

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
    
## Конфигурация

Полный пример конфигурации доступен в файле [`config.yaml`](config.yaml).

Для быстрого старта скачайте готовую конфигурацию:

```bash
wget https://raw.githubusercontent.com/capricornusx/meshtastic-mqtt-exporter/main/config.yaml
```

## Документация

- [Быстрый старт](docs/src/ru/quick-start.md) — Установка и первый запуск
- [Конфигурация](docs/src/ru/configuration.md) — Настройка YAML файла
- [API](docs/src/ru/api.md) — REST API endpoints

## Метрики

- `meshtastic_battery_level_percent` — Уровень батареи
- `meshtastic_temperature_celsius` — Температура
- `meshtastic_humidity_percent` — Влажность
- `meshtastic_pressure_hpa` — Барометрическое давление
- `meshtastic_rssi_dbm` — Мощность сигнала (dBm)
- `meshtastic_snr_db` — Отношение сигнал/шум (dB)
- `meshtastic_node_last_seen_timestamp` — Время последней активности

## Персистентность состояния

Метрики автоматически сохраняются и восстанавливаются между перезапусками:

- **Автоматическое сохранение**: Каждые 5 минут и при завершении работы
- **Восстановление при запуске**: Метрики загружаются из файла состояния
- **JSON формат**: Читаемый формат для отладки

Для отключения персистентности уберите параметр `hook.prometheus.state_file` из конфигурации.

## TODO
- [ ] добавить MQTT-специфичные метрики (обработано сообщений, uptime, расход памяти т.д.)
- [ ] from_node vs node_id labels
- [ ] синхронизация метрик с meshtastic .proto файлами
- [x] TLS поддержка для MQTT брокера

## Благодарности

Построен с использованием отличного MQTT брокера [mochi-mqtt](https://github.com/mochi-mqtt/server) от [@mochi-co](https://github.com/mochi-co).

## Лицензия

MIT License
