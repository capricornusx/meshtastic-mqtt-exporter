# Meshtastic MQTT Exporter

[![Build Status](https://github.com/capricornusx/meshtastic-mqtt-exporter/workflows/Build%20and%20Test/badge.svg)](https://github.com/capricornusx/meshtastic-mqtt-exporter/actions)
[![codecov](https://codecov.io/gh/capricornusx/meshtastic-mqtt-exporter/graph/badge.svg?token=P0409HCBFS)](https://codecov.io/gh/capricornusx/meshtastic-mqtt-exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/capricornusx/meshtastic-mqtt-exporter)](https://goreportcard.com/report/github.com/capricornusx/meshtastic-mqtt-exporter)

Экспорт телеметрии Meshtastic устройств в Prometheus с интеграцией AlertManager для отправки алертов в LoRa сеть.

## Возможности

- **Встроенный MQTT брокер** на основе mochi-mqtt
- **Поддержка TLS/QoS/Retention**
- **Prometheus метрики**: Батарея, температура, влажность, давление, качество сигнала
- **AlertManager интеграция**: Отправка алертов в LoRa mesh сеть
- **Персистентность состояния**: Сохранение/восстановление метрик между перезапусками

## Быстрый старт

```bash
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-linux-amd64

# Запустить
./mqtt-exporter-linux-amd64 --config config.yaml

# Проверить
curl http://localhost:8100/metrics
```

## Конфигурация

Полный пример конфигурации доступен в файле [`config.yaml`](config.yaml).

## Документация
- [Быстрый старт](docs/src/ru/quick-start.md) — Установка и первый запуск
- [Конфигурация](docs/src/ru/configuration.md) — Настройка YAML файла
- [API](docs/src/ru/api.md) — REST API endpoints
- [Pages](https://capricornusx.github.io/meshtastic-mqtt-exporter/)

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

## TODO
- [ ] настроить formatter через .golangci.yml 
- [ ] добавить MQTT-специфичные метрики (обработано сообщений, uptime, расход памяти т.д.)
- [ ] from_node vs node_id labels
- [ ] синхронизация метрик с meshtastic .proto файлами
- [ ] проверить код на избыточные функции, которые могут быть в стандартной библиотеке

## Благодарности

Построен с использованием отличного MQTT брокера [mochi-mqtt](https://github.com/mochi-mqtt/server) от [@mochi-co](https://github.com/mochi-co).

## Лицензия

MIT License
