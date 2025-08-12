# Meshtastic MQTT Exporter

Export Meshtastic device telemetry to Prometheus metrics with AlertManager integration for sending alerts to LoRa network.

## Features

- **Built-in MQTT broker** with YAML configuration
- **TLS support** - TCP (1883) + TLS (8883) ports simultaneously
- **Prometheus metrics**: Battery, temperature, humidity, pressure, signal quality
- **AlertManager integration**: Send alerts to LoRa mesh network
- **State persistence**: Save/restore metrics between restarts

## Quick Start

### Docker Compose (full stack)

```bash
# Full monitoring stack
cd docs/stack
docker-compose up -d
```

### Standalone binary

```bash
# Download binary
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-linux-amd64

# Run embedded mode
./mqtt-exporter-linux-amd64 --config config.yaml

# Check metrics
curl http://localhost:8100/metrics
```

## Documentation

- [Quick Start](quick-start/) — Installation and first run
- [Configuration](configuration/) — YAML file setup
- [API](api/) — REST API endpoints