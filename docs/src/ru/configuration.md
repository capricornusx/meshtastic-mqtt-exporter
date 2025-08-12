# Конфигурация

## Пример конфигурации

Полный пример конфигурации доступен в файле [`config.yaml`](https://github.com/capricornusx/meshtastic-mqtt-exporter/blob/main/config.yaml) в репозитории.

### Минимальная конфигурация

```yaml
logging:
  level: "info"

mqtt:
  host: 0.0.0.0
  port: 1883
  allow_anonymous: true

hook:
  listen: "0.0.0.0:8100"
  prometheus:
    path: "/metrics"
    topic:
      pattern: "msh/#"
```

## Параметры командной строки

| Параметр | Описание | По умолчанию |
|----------|------------|-------------|
| `--config` | Путь к файлу конфигурации | `config.yaml` |
| `--log-level` | Уровень логирования | `info` |
| `--help` | Показать справку | - |

## Переменные окружения

| Переменная | Описание | Пример |
|------------|------------|--------|
| `MQTT_HOST` | Хост MQTT брокера | `localhost` |
| `MQTT_PORT` | Порт MQTT брокера | `1883` |
| `HOOK_LISTEN` | Адрес сервера метрик | `0.0.0.0:8100` |
| `LOG_LEVEL` | Уровень логирования | `info` |

## Основные параметры

### MQTT топики

Параметр `hook.prometheus.topic.pattern` поддерживает wildcards:
- `+` — один уровень  
- `#` — много уровней

Примеры: `msh/#`, `msh/+/json/+/+`

## Запуск

```bash
# Основной режим
./mqtt-exporter-linux-amd64 --config config.yaml

# С отладкой
./mqtt-exporter-linux-amd64 --config config.yaml --log-level debug
```

## Проверка работы

```bash
# Метрики
curl http://localhost:8100/metrics

# Health check
curl http://localhost:8100/health

# Проверка конфигурации
./mqtt-exporter-linux-amd64 --config config.yaml --validate
```