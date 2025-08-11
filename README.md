# Meshtastic MQTT Exporter

[![Build Status](https://github.com/capricornusx/meshtastic-mqtt-exporter/workflows/Build%20and%20Test/badge.svg)](https://github.com/capricornusx/meshtastic-mqtt-exporter/actions)
[![codecov](https://codecov.io/gh/capricornusx/meshtastic-mqtt-exporter/graph/badge.svg?token=P0409HCBFS)](https://codecov.io/gh/capricornusx/meshtastic-mqtt-exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/capricornusx/meshtastic-mqtt-exporter)](https://goreportcard.com/report/github.com/capricornusx/meshtastic-mqtt-exporter)

Экспорт телеметрии Meshtastic устройств в метрики Prometheus с интеграцией AlertManager для отправки алертов в LoRa сеть.

## Возможности

- **mochi-mqtt хук**: Интеграция с существующими серверами (рекомендуется)
- **Embedded режим**: Встроенный MQTT брокер с YAML конфигурацией
- **Prometheus метрики**: Батарея, температура, влажность, давление, качество сигнала
  - [ ] добавить MQTT-специфичные метрики (обработано сообщений, uptime, расход памяти т.д.)
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

## Режимы работы

### 1. Embedded режим (рекомендуется)
```bash
./mqtt-exporter-embedded --config config.yaml
```

### 2. mochi-mqtt хук
```go
// Простой хук
hook := hooks.NewMeshtasticHookSimple()
server.AddHook(hook, nil)

// Современный хук с конфигурацией
hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
    ServerAddr:   ":8100",
    EnableHealth: true,
    TopicPrefix:  "msh/",
})
server.AddHook(hook, nil)
```

### 3. Standalone режим (для существующих MQTT серверов)
```bash
./mqtt-exporter-standalone --config config.yaml
```

## Пример конфигурации

```yaml
logging:
  level: "info"  # debug, info, warn, error, fatal

mqtt:
  host: 0.0.0.0
  port: 1883
  allow_anonymous: true

hook:
  listen: "0.0.0.0:8100"
  prometheus:
    path: "/metrics"
    topic:
      # MQTT topic pattern (поддерживает wildcards + и #)
      pattern: "msh/#"  # Все сообщения начинающиеся с msh/
      # Примеры других паттернов:
      # "msh/+/json/+/+"  - только JSON сообщения
      # "msh/+/c/+/+"     - только канальные сообщения
    debug:
      log_all_messages: false  # Логировать MQTT сообщения соответствующие pattern
    state:
      file: "meshtastic_state.json"  # Файл для сохранения состояния метрик
  alertmanager:
    path: "/alerts/webhook"
```

## Документация

### Конфигурация
- [Основные параметры](docs/src/ru/configuration/basic.md) — Режимы работы и параметры
- [YAML конфигурация](docs/src/ru/configuration/yaml.md) — Полное описание параметров
- [Переменные окружения](docs/src/ru/configuration/environment.md) — Настройка через ENV

### Развертывание
- [Docker](docs/src/ru/deployment/docker.md) — Контейнеризация и Docker Compose
- [Systemd](docs/src/ru/deployment/systemd.md) — Системный сервис

### Интеграция
- [mochi-mqtt хук](docs/src/ru/integration/hook.md) — Интеграция с MQTT сервером
- [AlertManager](docs/src/ru/integration/alertmanager.md) — LoRa mesh алерты
- [Prometheus](docs/src/ru/integration/prometheus.md) — Метрики и мониторинг

### Примеры
- [API документация](docs/src/ru/api.md) — REST API endpoints
- [Примеры использования](docs/src/ru/examples.md) — Интеграция с различными системами

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

Для отключения персистентности уберите параметр `hook.prometheus.state.file` из конфигурации.

## Благодарности

Построен с использованием отличного MQTT брокера [mochi-mqtt](https://github.com/mochi-mqtt/server) от [@mochi-co](https://github.com/mochi-co).

## Лицензия

MIT License
