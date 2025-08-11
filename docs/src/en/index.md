# Meshtastic MQTT Exporter

Export Meshtastic device telemetry to Prometheus metrics with AlertManager integration for sending alerts to LoRa network.

## Features

- **mochi-mqtt hook**: Integration with existing servers (recommended)
- **Embedded mode**: Built-in MQTT broker with YAML configuration
- **Prometheus metrics**: Battery, temperature, humidity, pressure, signal quality
- **AlertManager integration**: Send alerts to LoRa mesh network
- **State persistence**: Save/restore metrics between restarts

## Quick Start

```bash
# Download binary
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-linux-amd64

# Run embedded mode
./mqtt-exporter-linux-amd64 --config config.yaml

# Check metrics
curl http://localhost:8101/metrics
```

## Operating Modes

### 1. Embedded mode (recommended)
```bash
./mqtt-exporter-embedded --config config.yaml
```

### 2. mochi-mqtt hook
```go
hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
    PrometheusAddr: ":8101",
    EnableHealth:   true,
    TopicPrefix:    "msh/",
})
server.AddHook(hook, nil)
```

### 3. Standalone mode
```bash
./mqtt-exporter-standalone --config config.yaml
```

## Metrics

- `meshtastic_battery_level_percent` — Battery level
- `meshtastic_temperature_celsius` — Temperature
- `meshtastic_humidity_percent` — Humidity
- `meshtastic_pressure_hpa` — Barometric pressure
- `meshtastic_node_last_seen_timestamp` — Last activity timestamp

## Architecture

Project follows Clean Architecture and SOLID principles:

- **Domain**: Business logic and interfaces
- **Application**: Message processing and coordination
- **Infrastructure**: MQTT, HTTP, metrics
- **Adapters**: Configuration and external interfaces

## Acknowledgments

Built using [mochi-mqtt](https://github.com/mochi-mqtt/server) by [@mochi-co](https://github.com/mochi-co).