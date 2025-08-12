# Переменные окружения

## Основные переменные

| Переменная             | Описание              | Тип    | По умолчанию |
|------------------------|-----------------------|--------|--------------|
| `MQTT_HOST`            | Хост MQTT брокера     | string | `localhost`  |
| `MQTT_PORT`            | Порт MQTT брокера     | int    | `1883`       |
| `MQTT_TLS`             | Включить TLS          | bool   | `false`      |
| `MQTT_ALLOW_ANONYMOUS` | Анонимные подключения | bool   | `true`       |
| `PROMETHEUS_ENABLED`   | Включить метрики      | bool   | `true`       |
| `PROMETHEUS_HOST`      | Хост сервера метрик   | string | `0.0.0.0`    |
| `PROMETHEUS_PORT`      | Порт сервера метрик   | int    | `8100`       |
| `ALERTMANAGER_ENABLED` | Включить AlertManager | bool   | `false`      |
| `ALERTMANAGER_PORT`    | Порт webhook сервера  | int    | `8080`       |
| `LOG_LEVEL`            | Уровень логирования   | string | `info`       |

## Пример .env файла

```bash
# .env
MQTT_HOST=0.0.0.0
MQTT_PORT=1883
MQTT_ALLOW_ANONYMOUS=true

PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=8100

ALERTMANAGER_ENABLED=false
ALERTMANAGER_PORT=8080

LOG_LEVEL=info
```

## Использование с Docker

```bash
# Загрузка из .env файла
docker run --env-file .env ghcr.io/capricornusx/meshtastic-mqtt-exporter:latest

# Прямое указание переменных
docker run -e MQTT_HOST=0.0.0.0 -e PROMETHEUS_PORT=8100 \
  ghcr.io/capricornusx/meshtastic-mqtt-exporter:latest
```

## Приоритет конфигурации

1. **Переменные окружения** (высший приоритет)
2. **YAML файл конфигурации**
3. **Значения по умолчанию** (низший приоритет)

## Безопасность

### Чувствительные данные

Используйте переменные окружения для паролей:

```bash
export MQTT_PASSWORD="secure_password"
export ALERTMANAGER_TOKEN="webhook_token"
```

### Docker Secrets

```yaml
# docker-compose.yml
services:
  mqtt-exporter:
    image: ghcr.io/capricornusx/meshtastic-mqtt-exporter:latest
    secrets:
      - mqtt_password
    environment:
      - MQTT_PASSWORD_FILE=/run/secrets/mqtt_password

secrets:
  mqtt_password:
    file: ./secrets/mqtt_password.txt
```
