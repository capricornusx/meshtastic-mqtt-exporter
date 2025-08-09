# Meshtastic Hook for mochi-mqtt

This example shows how to integrate the Meshtastic Prometheus hook into any existing mochi-mqtt server.

## Usage

```go
import "meshtastic-exporter/pkg/hooks"

// Create hook
meshtasticHook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
    PrometheusAddr: ":8100",
    EnableHealth:   true,
})

// Add to your mochi-mqtt server
server.AddHook(meshtasticHook, nil)
```

## Configuration

```go
type MeshtasticHookConfig struct {
    PrometheusAddr string // Address to serve metrics (e.g., ":8100")
    EnableHealth   bool   // Enable /health endpoint
}
```

## Endpoints

- `GET /metrics` - Prometheus metrics
- `GET /health` - Health check (if enabled)

## Metrics

- `meshtastic_messages_total` - Total messages by type
- `meshtastic_battery_level_percent` - Battery level
- `meshtastic_temperature_celsius` - Temperature
- `meshtastic_humidity_percent` - Humidity
- `meshtastic_pressure_hpa` - Barometric pressure
- `meshtastic_node_last_seen_timestamp` - Last seen timestamp
- `meshtastic_node_info` - Node information

## Run Example

```bash
cd examples/mochi-mqtt-integration
go run main.go
```

Then test with:
```bash
curl http://localhost:8100/metrics
curl http://localhost:8100/health
```