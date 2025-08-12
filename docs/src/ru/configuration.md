# Конфигурация

## Пример конфигурации

Полный пример конфигурации с комментариями доступен в корневом файле [`config.yaml`](../../../config.yaml).

Для быстрого старта скачайте готовую конфигурацию:

```bash
wget https://raw.githubusercontent.com/capricornusx/meshtastic-mqtt-exporter/main/config.yaml
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

### MQTT Capabilities

Раздел `mqtt.capabilities` позволяет настроить возможности встроенного MQTT брокера:

- `maximum_inflight` — максимум неподтвержденных QoS 1/2 сообщений на клиента (по умолчанию: 1024)
- `maximum_client_writes_pending` — максимум сообщений в очереди клиента (по умолчанию: 1000)
- `receive_maximum` — максимум concurrent QoS сообщений на клиента (по умолчанию: 512)
- `maximum_qos` — максимальный уровень QoS: 0, 1, 2 (по умолчанию: 2)
- `retain_available` — поддержка retain сообщений (по умолчанию: true)
- `maximum_message_expiry_interval` — время жизни сообщений: "24h", "1h", "0" (по умолчанию: "24h")
- `maximum_clients` — максимум одновременных клиентов (по умолчанию: 1000)

### MQTT топики

Параметр `hook.prometheus.topic.pattern` поддерживает wildcards:
- `+` — один уровень  
- `#` — много уровней

Примеры: `msh/#`, `msh/+/json/+/+`

### Персистентность состояния

Параметр `hook.prometheus.state_file` задает файл для сохранения метрик между перезапусками.

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