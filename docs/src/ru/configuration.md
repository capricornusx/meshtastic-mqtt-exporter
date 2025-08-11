# Конфигурация

Данный раздел содержит подробную информацию о конфигурации Meshtastic MQTT Exporter.

## Разделы конфигурации

- **[Основные параметры](basic.md)** - Режимы работы, параметры командной строки, переменные окружения
- **[YAML конфигурация](yaml.md)** - Подробное описание всех параметров YAML файла
- **[Переменные окружения](environment.md)** - Настройка через переменные окружения

## Быстрый старт

### Минимальная конфигурация

```yaml
mqtt:
  host: 0.0.0.0
  port: 1883
  allow_anonymous: true

prometheus:
  enabled: true
  port: 8100
  topic:
    prefix: "msh/"
```

### Запуск

```bash
# Embedded режим
./mqtt-exporter-embedded --config config.yaml

# Standalone режим  
./mqtt-exporter-standalone --config config.yaml
```

## Валидация конфигурации

```bash
# Проверка синтаксиса YAML
yamllint config.yaml

# Проверка конфигурации приложением
./mqtt-exporter-embedded --config config.yaml --validate
```