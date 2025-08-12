# YAML Configuration

## Full Configuration Example

Complete configuration with comments available in [config.yaml](../../../config.yaml)

```bash
# Download ready configuration
wget https://raw.githubusercontent.com/capricornusx/meshtastic-mqtt-exporter/main/config.yaml
```

## Configuration Sections

### MQTT Broker

**Critical parameters:**
- `host: 0.0.0.0` — bind to all interfaces
- `port: 1883` — standard MQTT port
- `allow_anonymous: true` — allow connections without credentials

### Prometheus Metrics

**Critical parameters:**
- `topic.pattern: "msh/#"` — MQTT topic pattern (supports wildcards + and #)
- `state.file` — state persistence file path
- `metrics_ttl: "30m"` — inactive node metrics retention time

### AlertManager Integration

**Critical parameters:**
- `path: "/alerts/webhook"` — webhook endpoint path
- `channel: "LongFast"` — default Meshtastic channel
- `mode: "broadcast"` — delivery mode (broadcast/direct)