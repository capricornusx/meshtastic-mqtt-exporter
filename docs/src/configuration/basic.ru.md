# Основные параметры конфигурации

## Режимы работы

### Hook режим

Интеграция с существующим mochi-mqtt сервером:

```go
f := factory.NewFactory(cfg)
hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
    ServerAddr:  ":8100",
    TopicPrefix: "msh/",
    MetricsTTL:  30 * time.Minute,
}, f)
```

### Embedded режим

Встроенный MQTT брокер:

```bash
./mqtt-exporter-embedded --config config.yaml
```

### Standalone режим

Подключение к внешнему MQTT брокеру:

```bash
./mqtt-exporter-standalone --config config.yaml
```

## Параметры командной строки

| Параметр      | Описание                                       | По умолчанию  |
|---------------|------------------------------------------------|---------------|
| `--config`    | Путь к файлу конфигурации                      | `config.yaml` |
| `--log-level` | Уровень логирования (debug, info, warn, error) | `info`        |
| `--help`      | Показать справку                               | -             |

## Переменные окружения

| Переменная        | Описание               | Пример      |
|-------------------|------------------------|-------------|
| `MQTT_HOST`       | Хост MQTT брокера      | `localhost` |
| `MQTT_PORT`       | Порт MQTT брокера      | `1883`      |
| `PROMETHEUS_PORT` | Порт метрик Prometheus | `8100`      |
| `LOG_LEVEL`       | Уровень логирования    | `info`      |

## Валидация конфигурации

Проверка корректности конфигурации:

```bash
./mqtt-exporter-embedded --config config.yaml --validate
```
