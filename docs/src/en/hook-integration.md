# Hook Integration

Integration with existing mochi-mqtt servers via hook.

## Quick Start

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
        PrometheusAddr: ":8101",
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

## Hook Configuration

### Basic Configuration
```go
config := hooks.MeshtasticHookConfig{
    PrometheusAddr: ":8101",        // Prometheus metrics address
    TopicPrefix:    "msh/",         // MQTT topic prefix
    EnableHealth:   true,           // Enable /health endpoint
}
```

### Full Configuration
```go
config := hooks.MeshtasticHookConfig{
    PrometheusAddr: ":8101",
    TopicPrefix:    "msh/",
    EnableHealth:   true,
    MetricsTTL:     30 * time.Minute,
    
    // AlertManager integration
    AlertManager: struct {
        Enabled bool
        Addr    string
        Path    string
    }{
        Enabled: true,
        Addr:    ":8080",
        Path:    "/alerts/webhook",
    },
    
    // State persistence
    StateFile: "meshtastic_state.json",
}
```

## Integration Examples

### Basic Server
```go
package main

import (
    "log"
    
    mqtt "github.com/mochi-mqtt/server/v2"
    "github.com/mochi-mqtt/server/v2/listeners"
    "github.com/capricornusx/meshtastic-mqtt-exporter/pkg/hooks"
)

func main() {
    server := mqtt.New(nil)
    
    // Simple hook
    hook := hooks.NewMeshtasticHookSimple()
    server.AddHook(hook, nil)
    
    tcp := listeners.NewTCP("tcp", ":1883", nil)
    server.AddListener(tcp)
    
    log.Fatal(server.Serve())
}
```

### With AlertManager
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
        PrometheusAddr: ":8101",
        TopicPrefix:    "msh/",
        EnableHealth:   true,
        MetricsTTL:     30 * time.Minute,
        
        AlertManager: struct {
            Enabled bool
            Addr    string
            Path    string
        }{
            Enabled: true,
            Addr:    ":8080",
            Path:    "/alerts/webhook",
        },
    })
    server.AddHook(hook, nil)
    
    tcp := listeners.NewTCP("tcp", ":1883", nil)
    server.AddListener(tcp)
    
    log.Fatal(server.Serve())
}
```

## Endpoints

After hook integration, available endpoints:

- `GET /metrics` - Prometheus metrics
- `GET /health` - Health check
- `POST /alerts/webhook` - AlertManager webhook (if enabled)

## Metrics

Hook exports the following metrics:

- `meshtastic_battery_level_percent` - Battery level
- `meshtastic_temperature_celsius` - Temperature
- `meshtastic_humidity_percent` - Humidity
- `meshtastic_pressure_hpa` - Pressure
- `meshtastic_node_last_seen_timestamp` - Last activity timestamp

## Troubleshooting

### No Metrics Appearing
1. Check topic prefix (`TopicPrefix`)
2. Ensure devices publish to correct topics
3. Check hook logs

### AlertManager Not Receiving Alerts
1. Check `AlertManager.Addr` configuration
2. Ensure AlertManager is configured for correct URL
3. Verify webhook payload format

### Performance
- Use `MetricsTTL` to clean up old metrics
- Configure `max_inflight` and `max_queued` in MQTT broker

## Hook Interfaces

### MeshtasticHook
Main hook interface implementing all mochi-mqtt events:

```go
type MeshtasticHook interface {
    mqtt.Hook
    ID() string
    Provides(byte) bool
    Init(*mqtt.HookOptions) error
    Stop() error
    OnConnect(*mqtt.Client, mqtt.Packet) error
    OnDisconnect(*mqtt.Client, error, bool)
    OnSubscribed(*mqtt.Client, mqtt.Packet, []byte, []int, []byte) error
    OnUnsubscribed(*mqtt.Client, mqtt.Packet, []byte, []int) error
    OnPublished(*mqtt.Client, mqtt.Packet) error
    OnPublish(*mqtt.Client, mqtt.Packet) (mqtt.Packet, error)
    OnRetainMessage(*mqtt.Client, mqtt.Packet, int64) error
    OnQosPublish(*mqtt.Client, mqtt.Packet, int64, int64) error
    OnQosComplete(*mqtt.Client, mqtt.Packet) error
    OnQosDropped(*mqtt.Client, mqtt.Packet) error
    OnPacketRead(*mqtt.Client, mqtt.Packet) (mqtt.Packet, error)
    OnPacketEncode(*mqtt.Client, mqtt.Packet) mqtt.Packet
    OnConnectAuthenticate(*mqtt.Client, mqtt.Packet) bool
    OnACLCheck(*mqtt.Client, string, bool) bool
    OnSysInfoTick(mqtt.SysInfo)
    OnClientExpired(*mqtt.Client)
    OnRetainedExpired(string)
    OnWillSent(*mqtt.Client, mqtt.Packet) error
}
```

### Constructors

```go
// Simple hook with default settings
func NewMeshtasticHookSimple() *MeshtasticHook

// Hook with full configuration
func NewMeshtasticHook(config MeshtasticHookConfig) *MeshtasticHook
```