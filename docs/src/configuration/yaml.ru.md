# YAML конфигурация

## Полный пример конфигурации

Полная конфигурация с комментариями доступна в файле [config.yaml](../../../config.yaml)

```bash
# Скачать готовую конфигурацию
wget https://raw.githubusercontent.com/capricornusx/meshtastic-mqtt-exporter/main/config.yaml
```

## Секции конфигурации

### Логирование

| Параметр | Тип    | По умолчанию | Описание                                             |
|----------|--------|--------------|------------------------------------------------------|
| `level`  | string | `info`       | Уровень логирования: debug, info, warn, error, fatal |

### MQTT Брокер

| Параметр              | Тип    | По умолчанию | Описание                                    |
|-----------------------|--------|--------------|---------------------------------------------|
| `host`                | string | `localhost`  | Хост MQTT брокера (IPv4/IPv6)               |
| `port`                | int    | `1883`       | Порт MQTT брокера                           |
| `tls`                 | bool   | `false`      | Включить TLS шифрование                     |
| `allow_anonymous`     | bool   | `true`       | Разрешить анонимные подключения             |
| `users`               | array  | -            | Массив учетных записей пользователей        |
| `broker.max_inflight` | int    | `50`         | Макс. неподтвержденных сообщений на клиента |
| `broker.max_queued`   | int    | `1000`       | Макс. сообщений в очереди на клиента        |

| `debug.log_all_messages` | bool | `false`      | Логировать все входящие MQTT сообщения |

### HTTP Hook Server

| Параметр | Тип    | По умолчанию     | Описание                  |
|----------|--------|------------------|---------------------------|
| `listen` | string | `localhost:8100` | Адрес и порт HTTP сервера |

### Prometheus метрики

| Параметр                 | Тип    | По умолчанию | Описание                                                            |
|--------------------------|--------|--------------|---------------------------------------------------------------------|
| `path`                   | string | `/metrics`   | Путь к endpoint метрик                                              |
| `metrics_ttl`            | string | `30m`        | Время хранения метрик неактивных узлов                              |
| `topic.pattern`          | string | `msh/#`      | Паттерн MQTT топиков (поддерживает wildcards + и #)                 |
| `keep_alive`             | string | `"60s"`      | MQTT client keep alive (только standalone режим)                    |
| `topic.log_all_messages` | bool   | `false`      | Логировать MQTT сообщения соответствующие pattern                   |
| `state.file`             | string | -            | Путь к файлу состояния (если не указан - персистентность отключена) |

### AlertManager интеграция

| Параметр           | Тип    | По умолчанию            | Описание                             |
|--------------------|--------|-------------------------|--------------------------------------|
| `path`             | string | `/alerts/webhook`       | Путь HTTP webhook endpoint           |
| `channel`          | string | `LongFast`              | Канал Meshtastic по умолчанию        |
| `mode`             | string | `broadcast`             | Режим доставки по умолчанию          |
| `topics.broadcast` | string | `msh/2/c/%s/!broadcast` | Шаблон топика для broadcast          |
| `topics.direct`    | string | `msh/2/c/%s/!%s`        | Шаблон топика для прямых сообщений   |
| `routing.critical` | object | -                       | Настройки для критических алертов    |
| `routing.warning`  | object | -                       | Настройки для предупреждений         |
| `routing.info`     | object | -                       | Настройки для информационных алертов |

## Паттерны MQTT топиков

Приложение поддерживает MQTT wildcards для гибкой настройки топиков:

- `+` - заменяет один уровень топика
- `#` - заменяет все последующие уровни топика

### Примеры паттернов

```yaml
# Все сообщения начинающиеся с msh/
prometheus:
  topic:
    pattern: "msh/#"

# Только JSON сообщения Meshtastic
prometheus:
  topic:
    pattern: "msh/+/json/+/+"

# Только канальные сообщения
prometheus:
  topic:
    pattern: "msh/+/c/+/+"

# Конкретная структура топиков
prometheus:
  topic:
    pattern: "mesh/+/data/#"
```

## Персистентность состояния

Метрики могут автоматически сохраняться и восстанавливаться между перезапусками:

```yaml
hook:
  prometheus:
    state:
      file: "meshtastic_state.json"  # Включает персистентность
```

### Особенности:

- **Автоматическое сохранение**: Каждые 5 минут и при завершении работы
- **Восстановление при запуске**: Метрики загружаются из файла
- **JSON формат**: Читаемый формат для отладки
- **Отключение**: Уберите параметр `state.file` для отключения

### Пример файла состояния:

```json
{
  "version": "1.0",
  "timestamp": 1754947727,
  "nodes": [
    {
      "node_id": "node001",
      "timestamp": 1754947727,
      "metrics": {
        "meshtastic_battery_level_percent": 85.5,
        "meshtastic_temperature_celsius": 23.4,
        "meshtastic_node_info": 1
      },
      "labels": {
        "node_id": "node001",
        "longname": "Base Station",
        "hardware": "TBEAM"
      }
    }
  ]
}
```

## Отладка MQTT сообщений

Для отладки можно включить логирование MQTT сообщений:

```yaml
logging:
  level: "debug"

hook:
  prometheus:
    topic:
      log_all_messages: true
```

При включенном режиме в логах будут отображаться сообщения соответствующие pattern:

```
DBG mqtt message received topic=msh/test/topic payload={"from":123,"data":"..."}
```

## Режимы доставки алертов

### Broadcast режим

Отправляет алерты всем узлам в mesh сети:

```yaml
alertmanager:
  mode: "broadcast"
  channel: "LongFast"
```

### Direct режим

Отправляет алерты только указанным узлам:

```yaml
alertmanager:
  mode: "direct"
  channel: "ShortFast"
  target_nodes:
    - "admin001"
    - "monitor02"
```

## Валидация YAML

Проверка синтаксиса конфигурации:

```bash
# Проверка YAML синтаксиса
yamllint config.yaml

# Проверка конфигурации приложением
./mqtt-exporter-embedded --config config.yaml --validate
```
