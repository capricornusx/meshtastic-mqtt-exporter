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

### Pre-built Binaries

Download the latest release for your platform from [GitHub Releases](https://github.com/capricornusx/meshtastic-mqtt-exporter/releases).

### Build from Source

```bash
git clone https://github.com/capricornusx/meshtastic-mqtt-exporter
cd meshtastic-mqtt-exporter
go mod download
```

### Building for Raspberry Pi

```bash
# Build for all Raspberry Pi variants
make build-rpi

# Or build specific versions:
make build-rpi-arm64  # Pi 4/5 (64-bit)
make build-rpi-arm32  # Older Pi (32-bit)
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


### TODO:
 - [ ] 💯 Add support for all sensor types from [telemetry.proto](https://github.com/meshtastic/protobufs/blob/master/meshtastic/telemetry.proto)
 - [ ] ♻ Auto-release (sync) on upstream .proto 👆 changes 
 - [ ] 📊 Create an example Grafana dashboard
 - [ ] 🔥 Create basic AlertManager rules

## Acknowledgments

Built using the excellent [mochi-mqtt](https://github.com/mochi-mqtt/server) MQTT broker by [@mochi-co](https://github.com/mochi-co).

## License

MIT License
