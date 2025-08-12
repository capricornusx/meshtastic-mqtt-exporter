# Configuration

## Configuration Example

Full configuration example with detailed comments is available in the root [`config.yaml`](../../../config.yaml) file.

For quick start, download the ready-to-use configuration:

```bash
wget https://raw.githubusercontent.com/capricornusx/meshtastic-mqtt-exporter/main/config.yaml
```

## Key Parameters

### MQTT Capabilities

The `mqtt.capabilities` section allows configuring the embedded MQTT broker capabilities:

- `maximum_inflight` — maximum unacknowledged QoS 1/2 messages per client (default: 1024)
- `maximum_client_writes_pending` — maximum messages in client queue (default: 1000)
- `receive_maximum` — maximum concurrent QoS messages per client (default: 512)
- `maximum_qos` — maximum QoS level: 0, 1, 2 (default: 2)
- `retain_available` — support for retain messages (default: true)
- `maximum_message_expiry_interval` — message lifetime: "24h", "1h", "0" (default: "24h")
- `maximum_clients` — maximum concurrent clients (default: 1000)

### MQTT Topics

The `hook.prometheus.topic.pattern` parameter supports wildcards:
- `+` — single level
- `#` — multiple levels

Examples: `msh/#`, `msh/+/json/+/+`

### State Persistence

The `hook.prometheus.state_file` parameter specifies the file for saving metrics state between restarts.

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