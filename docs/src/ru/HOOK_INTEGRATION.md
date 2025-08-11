# Hook Integration Guide

Интеграция с существующими mochi-mqtt серверами через хук.

## Быстрый старт

```go
package main

import (
    "log"
    "time"
    
    mqtt "github.com/mochi-mqtt/server/v2"
    "github.com/mochi-mqtt/server/v2/listeners"
    "github.com/capricornusx/meshtastic-mqtt-exporter/pkg/hooks"
)

func main() {
    server := mqtt.New(nil)
    
    hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
        PrometheusAddr: ":8100",
        EnableHealth:   true,
        TopicPrefix:    "msh/",
        MetricsTTL:     30 * time.Minute,
    })
    server.AddHook(hook, nil)
    
    tcp := listeners.NewTCP("tcp", ":1883", nil)
    server.AddListener(tcp)
    
    log.Fatal(server.Serve())
}
```

## Конфигурация хука

### Базовая конфигурация
```go
config := hooks.MeshtasticHookConfig{
    PrometheusAddr: ":8100",        // Адрес Prometheus метрик
    TopicPrefix:    "msh/",         // Префикс MQTT топиков
    EnableHealth:   true,           // Включить /health endpoint
}
```

### Полная конфигурация
```go
config := hooks.MeshtasticHookConfig{
    PrometheusAddr: ":8100",
    TopicPrefix:    "msh/",
    EnableHealth:   true,
    MetricsTTL:     30 * time.Minute,
    
    // AlertManager интеграция
    AlertManager: struct {
        Enabled bool
        Addr    string
        Path    string
    }{
        Enabled: true,
        Addr:    ":8080",
        Path:    "/alerts/webhook",
    },
    
    // Персистентность состояния
    StateFile: "meshtastic_state.json",
}
```

## Примеры интеграции

### Базовый сервер
См. [mochi-mqtt-integration/main.go](../mochi-mqtt-integration/main.go)

### С AlertManager
См. [with-alertmanager/main.go](with-alertmanager/main.go)

## Endpoints

После интеграции хука доступны:

- `GET /metrics` - Prometheus метрики
- `GET /health` - Health check
- `POST /alerts/webhook` - AlertManager webhook (если включен)

## Метрики

Хук экспортирует следующие метрики:

- `meshtastic_battery_level_percent` - Уровень батареи
- `meshtastic_temperature_celsius` - Температура
- `meshtastic_humidity_percent` - Влажность
- `meshtastic_pressure_hpa` - Давление
- `meshtastic_node_last_seen_timestamp` - Время последней активности

## Troubleshooting

### Метрики не появляются
1. Проверьте префикс топиков (`TopicPrefix`)
2. Убедитесь, что устройства публикуют в правильные топики
3. Проверьте логи хука

### AlertManager не получает алерты
1. Проверьте конфигурацию `AlertManager.Addr`
2. Убедитесь, что AlertManager настроен на правильный URL
3. Проверьте формат webhook payload

### Производительность
- Используйте `MetricsTTL` для очистки старых метрик
- Настройте `max_inflight` и `max_queued` в MQTT брокере