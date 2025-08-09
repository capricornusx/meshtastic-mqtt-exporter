# Meshtastic [mochi-mqtt](https://github.com/mochi-mqtt/server) Plugin

[![Build Status](https://github.com/capricornusx/meshtastic-mqtt-exporter/workflows/Build%20and%20Test/badge.svg)](https://github.com/capricornusx/meshtastic-mqtt-exporter/actions)
[![codecov](https://codecov.io/gh/capricornusx/meshtastic-mqtt-exporter/graph/badge.svg?token=P0409HCBFS)](https://codecov.io/gh/capricornusx/meshtastic-mqtt-exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/capricornusx/meshtastic-mqtt-exporter)](https://goreportcard.com/report/github.com/capricornusx/meshtastic-mqtt-exporter)

A mochi-mqtt server plugin that exports Meshtastic device telemetry to Prometheus metrics.

## Features

- **Standalone mode**: Connect to external MQTT broker (for existing setups)
- **Embedded mode**: Built-in MQTT broker with Prometheus hook (recommended)
- **mochi-mqtt Hook**: Standalone hook for existing mochi-mqtt servers
- **Prometheus metrics**: Battery, temperature, humidity, pressure, signal quality
- **Authentication**: Support for multiple users and anonymous connections
- **State persistence**: Save/restore metrics between restarts

## Installation

```bash
git clone https://github.com/capricornusx/meshtastic-mqtt-exporter
cd meshtastic-mqtt-exporter
go mod download
```

## Usage

### Embedded Mode (Recommended)

```bash
go run ./cmd/embedded-hook --config config.yaml
```

### Standalone Mode

```bash
go run ./cmd/standalone --config config.yaml
```

### As mochi-mqtt Hook

```go
import "meshtastic-exporter/pkg/hooks"

// Add to your existing mochi-mqtt server
hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
PrometheusAddr: ":8100",
EnableHealth:   true,
})
server.AddHook(hook, nil)
```

See [example](examples/mochi-mqtt-integration/README.md) for complete integration.

## Mode Comparison

| Feature         | Embedded Mode              | Standalone Mode               |
|-----------------|----------------------------|-------------------------------|
| **Setup**       | Single binary              | Requires external MQTT broker |
| **Performance** | Higher (direct processing) | Lower (network overhead)      |
| **Resources**   | Lower                      | Higher                        |
| **Use Case**    | New deployments            | Existing MQTT infrastructure  |
| **Recommended** | ✅ Yes                      | For legacy setups             |

## Configuration

See [Configuration Guide](docs/CONFIGURATION.md) for detailed options.

## Docker Deployment

See docs/ for container deployment with health checks.

Basic example:

```yaml
mqtt:
  host: localhost
  port: 1883
  allow_anonymous: true

prometheus:
  enabled: true
  port: 8100
```

## Metrics

- `meshtastic_messages_total` - Total messages by type
- `meshtastic_battery_level_percent` - Battery level
- `meshtastic_temperature_celsius` - Temperature
- `meshtastic_humidity_percent` - Humidity
- `meshtastic_pressure_hpa` - Barometric pressure
- `meshtastic_rssi_dbm` - Signal strength
- `meshtastic_mqtt_up` - MQTT connection status
- `meshtastic_node_last_seen_timestamp` - Last seen timestamp

## Architecture

### Embedded Mode (Recommended)

```
Meshtastic Devices → Built-in MQTT Broker → Prometheus Hook → Metrics
```

### Standalone Mode

```
Meshtastic Devices → External MQTT Broker → MQTT Client → Exporter → Prometheus
```

## License

MIT License
