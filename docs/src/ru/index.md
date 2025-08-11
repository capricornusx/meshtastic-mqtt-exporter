# Meshtastic MQTT Exporter

Экспорт телеметрии Meshtastic устройств в метрики Prometheus с интеграцией AlertManager для отправки алертов в LoRa сеть.

## Возможности

- **mochi-mqtt хук**: Интеграция с существующими серверами (рекомендуется)
- **Embedded режим**: Встроенный MQTT брокер с YAML конфигурацией  
- **Prometheus метрики**: Батарея, температура, влажность, давление, качество сигнала
- **AlertManager интеграция**: Отправка алертов в LoRa mesh сеть
- **Персистентность состояния**: Сохранение/восстановление метрик между перезапусками

## Быстрый старт

```bash
# Скачать бинарник
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-linux-amd64

# Запустить embedded режим
./mqtt-exporter-linux-amd64 --config config.yaml

# Проверить метрики
curl http://localhost:8101/metrics
```

## Режимы работы

### 1. Embedded режим (рекомендуется)
```bash
./mqtt-exporter-embedded --config config.yaml
```

### 2. mochi-mqtt хук
```go
hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
    PrometheusAddr: ":8101",
    EnableHealth:   true,
    TopicPrefix:    "msh/",
})
server.AddHook(hook, nil)
```

### 3. Standalone режим
```bash
./mqtt-exporter-standalone --config config.yaml
```

## Метрики

- `meshtastic_battery_level_percent` — Уровень батареи
- `meshtastic_temperature_celsius` — Температура
- `meshtastic_humidity_percent` — Влажность  
- `meshtastic_pressure_hpa` — Барометрическое давление
- `meshtastic_node_last_seen_timestamp` — Время последней активности

## Архитектура

Проект следует принципам Clean Architecture и SOLID:

- **Domain**: Бизнес-логика и интерфейсы
- **Application**: Обработка сообщений и координация
- **Infrastructure**: MQTT, HTTP, метрики
- **Adapters**: Конфигурация и внешние интерфейсы

## Навигация по документации

### Конфигурация
- **[Основные параметры](configuration/basic.md)** - Режимы работы и параметры командной строки
- **[YAML конфигурация](configuration/yaml.md)** - Подробное описание всех параметров
- **[Переменные окружения](configuration/environment.md)** - Настройка через переменные окружения

### Развертывание
- **[Docker](deployment/docker.md)** - Контейнеризация и Docker Compose
- **[Systemd](deployment/systemd.md)** - Установка как системный сервис

### Интеграция
- **[mochi-mqtt хук](integration/hook.md)** - Интеграция с существующим MQTT сервером
- **[AlertManager](integration/alertmanager.md)** - Отправка алертов в LoRa сеть
- **[Prometheus](integration/prometheus.md)** - Настройка метрик и мониторинга

## Благодарности

Построен с использованием [mochi-mqtt](https://github.com/mochi-mqtt/server) от [@mochi-co](https://github.com/mochi-co).