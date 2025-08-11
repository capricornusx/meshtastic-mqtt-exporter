# Configuration

## Operating Modes

### Hook Mode

```go
hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
    PrometheusAddr: ":8100",
    TopicPrefix:    "msh/",
    MetricsTTL:     30 * time.Minute,
})
```

### Embedded Mode

```bash
./mqtt-exporter-embedded --config config.yaml
```

### Standalone Mode

```bash
./mqtt-exporter-standalone --config config.yaml
```

## YAML Configuration

```yaml
logging:
  level: "info"  # debug, info, warn, error, fatal

mqtt:
  host: 0.0.0.0
  port: 1883
  tls: false
  allow_anonymous: false
  users:
    - username: "meshtastic"
      password: "mesh456"
    - username: "monitor"
      password: "mon789"
  debug:
    log_all_messages: false  # Log all incoming MQTT messages
  broker:
    max_inflight: 50
    max_queued: 1000
    keep_alive: 60

prometheus:
  enabled: true
  host: 0.0.0.0
  port: 8100
  metrics_ttl: "30m"
  topic:
    # MQTT topic pattern (supports wildcards + and #)
    pattern: "msh/#"  # All messages starting with msh/
    # Examples of other patterns:
    # "msh/+/json/+/+"  - only JSON messages
    # "msh/+/c/+/+"     - only channel messages
  state:
    enabled: true
    file: "meshtastic_state.json"

alertmanager:
  enabled: true
  channel: "LongFast"
  mode: "broadcast"
  target_nodes:
    - "ffffffff"
    - "12345678"
  topics:
    broadcast: "msh/2/c/%s/!broadcast"
    direct: "msh/2/c/%s/!%s"
  http:
    port: 8080
    path: "/alerts/webhook"
```

## Configuration Parameters

### Logging

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `level` | string | `info` | Logging level: debug, info, warn, error, fatal |

### MQTT Broker

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `host` | string | `localhost` | MQTT broker host (IPv4/IPv6) |
| `port` | int | `1883` | MQTT broker port |
| `tls` | bool | `false` | Enable TLS encryption |
| `allow_anonymous` | bool | `true` | Allow anonymous connections |
| `users` | array | - | User credentials array |
| `broker.max_inflight` | int | `50` | Max unacknowledged messages per client |
| `broker.max_queued` | int | `1000` | Max queued messages per client |
| `broker.keep_alive` | int | `60` | Keep alive interval in seconds |
| `debug.log_all_messages` | bool | `false` | Log all incoming MQTT messages |

### Prometheus Metrics

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable Prometheus metrics |
| `host` | string | `0.0.0.0` | Metrics server host |
| `port` | int | `8100` | Metrics server port |
| `metrics_ttl` | string | `30m` | Time to keep inactive node metrics |
| `topic.pattern` | string | `msh/#` | MQTT topic pattern (supports wildcards + and #) |
| `state.enabled` | bool | `false` | Enable state persistence |
| `state.file` | string | `meshtastic_state.json` | State file path |

### AlertManager Integration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `false` | Enable AlertManager integration |
| `channel` | string | `LongFast` | Default Meshtastic channel |
| `mode` | string | `broadcast` | Default delivery mode |
| `target_nodes` | array | - | Default target node IDs |
| `topics.broadcast` | string | `msh/2/c/%s/!broadcast` | Broadcast topic pattern |
| `topics.direct` | string | `msh/2/c/%s/!%s` | Direct message topic pattern |
| `http.port` | int | `8080` | HTTP webhook server port |
| `http.path` | string | `/alerts/webhook` | HTTP webhook endpoint path |

## AlertManager Setup

### 1. Configure AlertManager

```yaml
# alertmanager.yml
global:
  smtp_smarthost: 'localhost:587'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'lora-alerts'

receivers:
- name: 'lora-alerts'
  webhook_configs:
  - url: 'http://localhost:8080/alerts/webhook'
    send_resolved: true
```

### 2. Prometheus Rules

```yaml
# meshtastic.rules.yml
groups:
- name: meshtastic
  rules:
  - alert: NodeOffline
    expr: (time() - meshtastic_node_last_seen_timestamp) > 1200
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Meshtastic node {{ $labels.node_id }} is offline"
      
  - alert: LowBattery
    expr: meshtastic_battery_level_percent < 20
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "Node {{ $labels.node_id }} battery low: {{ $value }}%"
```

## MQTT Topic Patterns

The application supports MQTT wildcards for flexible topic configuration:

- `+` - matches a single topic level
- `#` - matches all subsequent topic levels

### Pattern Examples

```yaml
# All messages starting with msh/
prometheus:
  topic:
    pattern: "msh/#"

# Only JSON Meshtastic messages
prometheus:
  topic:
    pattern: "msh/+/json/+/+"

# Only channel messages
prometheus:
  topic:
    pattern: "msh/+/c/+/+"

# Custom topic structure
prometheus:
  topic:
    pattern: "mesh/+/data/#"
```

## MQTT Message Debugging

For debugging, you can enable logging of all incoming MQTT messages:

```yaml
logging:
  level: "debug"

mqtt:
  debug:
    log_all_messages: true
```

With this enabled, all messages will be displayed in logs:

```
DBG mqtt message received topic=msh/test/topic payload={"from":123,"data":"..."}
```

### 3. Alert Delivery Modes

#### Broadcast Mode
Sends alerts to all nodes in the mesh network:
```yaml
alertmanager:
  mode: "broadcast"
  channel: "LongFast"
```

#### Direct Mode
Sends alerts to specific nodes only:
```yaml
alertmanager:
  mode: "direct"
  channel: "ShortFast"
  target_nodes:
    - "admin001"
    - "monitor02"
```

## Systemd Service Installation

### 1. Create System User
```bash
sudo useradd --system --no-create-home --shell /bin/false mqtt-exporter
```

### 2. Setup Directories
```bash
sudo mkdir -p /opt/mqtt-exporter /etc/mqtt-exporter /var/lib/mqtt-exporter
sudo chown mqtt-exporter:mqtt-exporter /var/lib/mqtt-exporter
```

### 3. Install Binary and Config
```bash
sudo cp meshtastic-exporter-embedded /opt/mqtt-exporter/
sudo cp config.yaml /etc/mqtt-exporter/
sudo chown root:root /opt/mqtt-exporter/meshtastic-exporter-embedded
sudo chmod 755 /opt/mqtt-exporter/meshtastic-exporter-embedded
```

### 4. Create Service File
```ini
# /etc/systemd/system/meshtastic-exporter.service
[Unit]
Description=Meshtastic MQTT Exporter
After=network.target

[Service]
Type=simple
User=mqtt-exporter
Group=mqtt-exporter
ExecStart=/opt/mqtt-exporter/meshtastic-exporter-embedded --config /etc/mqtt-exporter/config.yaml
Restart=always
RestartSec=5
WorkingDirectory=/var/lib/mqtt-exporter

# Security
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
ReadWritePaths=/var/lib/mqtt-exporter

[Install]
WantedBy=multi-user.target
```

### 5. Enable Service
```bash
sudo systemctl daemon-reload
sudo systemctl enable meshtastic-exporter
sudo systemctl start meshtastic-exporter
```

## Testing

### Health Check
```bash
curl http://localhost:8100/health
```

### Metrics
```bash
curl http://localhost:8100/metrics
```

### Test Alert
```bash
curl -X POST http://localhost:8080/alerts/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "alerts": [{
      "status": "firing",
      "labels": {"alertname": "TestAlert"},
      "annotations": {"summary": "Test alert message"}
    }]
  }'
```