# Meshtastic MQTT Exporter

Export Meshtastic device telemetry to Prometheus with AlertManager integration.

## Features

- **Built-in MQTT broker** based on mochi-mqtt
- **Prometheus metrics**: Battery, temperature, humidity, pressure, signal quality
- **AlertManager integration**: Send alerts to LoRa network
- **State persistence**: Save/restore metrics between restarts

## Quick Start

```bash
# Download and run
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-linux-amd64
wget https://raw.githubusercontent.com/capricornusx/meshtastic-mqtt-exporter/main/config.yaml
./mqtt-exporter-linux-amd64 --config config.yaml

# Check
curl http://localhost:8100/metrics
```

## Metrics

- `meshtastic_battery_level_percent` — Battery level
- `meshtastic_temperature_celsius` — Temperature
- `meshtastic_humidity_percent` — Humidity
- `meshtastic_pressure_hpa` — Pressure
- `meshtastic_rssi_dbm` — Signal strength
- `meshtastic_node_last_seen_timestamp` — Last activity