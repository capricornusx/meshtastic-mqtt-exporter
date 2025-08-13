# Конфигурация

## Быстрый старт

```bash
# Скачать готовую конфигурацию
wget https://raw.githubusercontent.com/capricornusx/meshtastic-mqtt-exporter/main/config.yaml

# Запуск
./mqtt-exporter-linux-amd64 --config config.yaml
```

## Параметры командной строки

| Параметр      | Описание                  | По умолчанию  |
|---------------|---------------------------|---------------|
| `--config`    | Путь к файлу конфигурации | `config.yaml` |
| `--log-level` | Уровень логирования       | `info`        |

## Переменные окружения

| Переменная    | Описание             | Пример         |
|---------------|----------------------|----------------|
| `MQTT_HOST`   | Хост MQTT брокера    | `localhost`    |
| `MQTT_PORT`   | Порт MQTT брокера    | `1883`         |
| `HOOK_LISTEN` | Адрес сервера метрик | `0.0.0.0:8100` |

## Основные секции YAML

### MQTT брокер

```yaml
mqtt:
  host: "0.0.0.0"
  port: 1883
  allow_anonymous: true
  users:
    - username: "meshtastic"
      password: "password"
```

### HTTP сервер

```yaml
hook:
  listen: "0.0.0.0:8100"
  prometheus:
    path: "/metrics"
    topic:
      pattern: "msh/#"
    state_file: "meshtastic_state.json"
```

### AlertManager

```yaml
alertmanager:
  path: "/alerts/webhook"
  channel: "LongFast"
  mode: "broadcast"
  mqtt_downlink_topic: "msh/US/2/json/mqtt/"  # Топик для отправки в LoRa сеть
```

## MQTT топики

Поддерживаются wildcards:
- `+` — один уровень
- `#` — много уровней

Примеры: `msh/#`, `msh/+/json/+/+`

## Персистентность

Метрики автоматически сохраняются в JSON файл:

```yaml
hook:
  prometheus:
    state_file: "meshtastic_state.json"
```

## Проверка конфигурации

```bash
# Валидация
./mqtt-exporter-linux-amd64 --config config.yaml --validate

# Отладка
./mqtt-exporter-linux-amd64 --config config.yaml --log-level debug
```
