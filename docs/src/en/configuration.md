# Configuration

## Configuration Example

Full configuration example is available in [`config.yaml`](https://github.com/capricornusx/meshtastic-mqtt-exporter/blob/main/config.yaml) file in the repository.

### Minimal Configuration

```yaml
logging:
  level: "info"

mqtt:
  host: 0.0.0.0
  port: 1883
  allow_anonymous: true

hook:
  listen: "0.0.0.0:8100"
  prometheus:
    path: "/metrics"
    topic:
      pattern: "msh/#"
```

## Key Parameters

### MQTT Topics

The `hook.prometheus.topic.pattern` parameter supports wildcards:
- `+` — single level
- `#` — multiple levels

Examples: `msh/#`, `msh/+/json/+/+`

## Command Line Parameters

| Parameter | Description | Default |
|-----------|-------------|----------|
| `--config` | Configuration file path | `config.yaml` |
| `--log-level` | Logging level | `info` |
| `--help` | Show help | - |

## Environment Variables

| Variable | Description | Example |
|----------|-------------|----------|
| `MQTT_HOST` | MQTT broker host | `localhost` |
| `MQTT_PORT` | MQTT broker port | `1883` |
| `HOOK_LISTEN` | Metrics server address | `0.0.0.0:8100` |
| `LOG_LEVEL` | Logging level | `info` |

## Running

```bash
./mqtt-exporter-linux-amd64 --config config.yaml
```

## Verification

```bash
# Metrics
curl http://localhost:8100/metrics

# Health check
curl http://localhost:8100/health

# Validate config
./mqtt-exporter-linux-amd64 --config config.yaml --validate
```