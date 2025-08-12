# mochi-mqtt хук интеграция

## Обзор

mochi-mqtt хук позволяет интегрировать экспортер с существующим MQTT сервером без необходимости запуска отдельного процесса.

## Простая интеграция

Готовый пример простой интеграции доступен в файле [mochi-mqtt-integration/main.go](../mochi-mqtt-integration/main.go).

## Расширенная конфигурация

```go
package main

import (
    "log"
    "time"

    mqtt "github.com/mochi-mqtt/server/v2"
    "github.com/mochi-mqtt/server/v2/hooks/auth"
    "github.com/mochi-mqtt/server/v2/listeners"
    
    "meshtastic-exporter/pkg/config"
    "meshtastic-exporter/pkg/factory"
    "meshtastic-exporter/pkg/hooks"
)

func main() {
    server := mqtt.New(&mqtt.Options{
        InlineClient: true,
    })

    // TCP листенер
    tcp := listeners.NewTCP("tcp1", ":1883", nil)
    server.AddListener(tcp)

    // WebSocket листенер
    ws := listeners.NewWebsocket("ws1", ":8080", nil)
    server.AddListener(ws)

    // Аутентификация
    server.AddHook(new(auth.AllowHook), nil)

    // Загрузка конфигурации
    cfg, err := config.LoadUnifiedConfig("config.yaml")
    if err != nil {
        log.Fatal(err)
    }
    f := factory.NewFactory(cfg)

    // Meshtastic хук с конфигурацией
    hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
        ServerAddr:   ":8100",
        EnableHealth: true,
        TopicPrefix:  "msh/",
        MetricsTTL:   30 * time.Minute,
        AlertPath:    "/alerts/webhook",
    }, f)
    
    server.AddHook(hook, nil)

    // Запуск сервера
    server.Serve()
}
```

## Конфигурация хука

### MeshtasticHookConfig

| Параметр | Тип | По умолчанию | Описание |
|----------|-----|--------------|----------|
| `ServerAddr` | string | `localhost:8100` | Адрес сервера метрик |
| `EnableHealth` | bool | `true` | Включить health endpoint |
| `TopicPrefix` | string | `msh/` | Префикс MQTT топиков |
| `MetricsTTL` | time.Duration | `30m` | TTL метрик неактивных узлов |
| `AlertPath` | string | `/alerts` | Путь AlertManager webhook |

**Примечание:** Персистентность состояния настраивается через YAML конфигурацию, а не через параметры хука.

## Интеграция с существующим сервером

### Добавление к существующему проекту

```go
// main.go
package main

import (
    "github.com/your-org/your-mqtt-server/internal/server"
    "meshtastic-exporter/pkg/config"
    "meshtastic-exporter/pkg/factory"
    "meshtastic-exporter/pkg/hooks"
)

func main() {
    // Ваш существующий сервер
    mqttServer := server.New()
    
    // Загрузка конфигурации
    cfg, _ := config.LoadUnifiedConfig("config.yaml")
    f := factory.NewFactory(cfg)
    
    // Добавление Meshtastic хука
    meshtasticHook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
        ServerAddr:  ":8100",
        TopicPrefix: "msh/",
    }, f)
    
    mqttServer.AddHook(meshtasticHook, nil)
    
    // Запуск сервера
    mqttServer.Serve()
}
```

### go.mod

```go
module your-mqtt-server

go 1.21

require (
    github.com/mochi-mqtt/server/v2 v2.4.1
    meshtastic-exporter v0.1.0
)
```

## Middleware интеграция

```go
package main

import (
    "context"
    "log"

    mqtt "github.com/mochi-mqtt/server/v2"
    "github.com/mochi-mqtt/server/v2/packets"
    
    "meshtastic-exporter/pkg/factory"
    "meshtastic-exporter/pkg/hooks"
)

// Кастомный хук с middleware
type CustomHook struct {
    mqtt.HookBase
    meshtasticHook *hooks.MeshtasticHook
}

func NewCustomHook() *CustomHook {
    f := factory.NewDefaultFactory()
    return &CustomHook{
        meshtasticHook: hooks.NewMeshtasticHookSimple(f),
    }
}

func (h *CustomHook) ID() string {
    return "custom-meshtastic-hook"
}

func (h *CustomHook) Provides(b byte) bool {
    return h.meshtasticHook.Provides(b)
}

func (h *CustomHook) Init(config any) error {
    return h.meshtasticHook.Init(config)
}

func (h *CustomHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
    // Кастомная логика перед обработкой
    log.Printf("Processing message from client %s on topic %s", cl.ID, pk.TopicName)
    
    // Передача в Meshtastic хук
    return h.meshtasticHook.OnPublish(cl, pk)
}

func (h *CustomHook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
    log.Printf("Client %s disconnected", cl.ID)
    h.meshtasticHook.OnDisconnect(cl, err, expire)
}

func main() {
    server := mqtt.New(nil)
    
    // Добавление кастомного хука
    server.AddHook(NewCustomHook(), nil)
    
    server.Serve()
}
```

## Мониторинг хука

### Метрики хука

```go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "net/http"
)

var (
    hookMessages = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "meshtastic_hook_messages_total",
            Help: "Total number of messages processed by hook",
        },
        []string{"topic", "status"},
    )
)

func init() {
    prometheus.MustRegister(hookMessages)
}

func main() {
    // Ваш MQTT сервер с хуком
    server := setupMQTTServer()
    
    // Prometheus метрики
    http.Handle("/metrics", promhttp.Handler())
    go http.ListenAndServe(":8102", nil)
    
    server.Serve()
}
```

### Health check

```go
package main

import (
    "encoding/json"
    "net/http"
    "time"
)

type HealthStatus struct {
    Status    string    `json:"status"`
    Timestamp time.Time `json:"timestamp"`
    Uptime    string    `json:"uptime"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    status := HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Uptime:    time.Since(startTime).String(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(status)
}

func main() {
    startTime = time.Now()
    
    // Health endpoint
    http.HandleFunc("/health", healthHandler)
    go http.ListenAndServe(":8103", nil)
    
    // MQTT сервер
    server := setupMQTTServer()
    server.Serve()
}
```

## Troubleshooting

### Отладка хука

```go
package main

import (
    "log"
    "os"
    
    "meshtastic-exporter/pkg/factory"
    "meshtastic-exporter/pkg/hooks"
)

func main() {
    // Включение отладочного режима
    os.Setenv("LOG_LEVEL", "debug")
    
    f := factory.NewDefaultFactory()
    hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
        ServerAddr:  ":8100",
        TopicPrefix: "msh/",
    }, f)
    
    server := mqtt.New(nil)
    server.AddHook(hook, nil)
    
    log.Println("Starting MQTT server with Meshtastic hook...")
    server.Serve()
}
```

### Проверка интеграции

```bash
# Проверка метрик
curl http://localhost:8100/metrics | grep meshtastic

# Проверка health
curl http://localhost:8100/health

# Тестирование MQTT
mosquitto_pub -h localhost -t "msh/2/e/LongFast/!test" -m '{"battery_level": 85}'
```