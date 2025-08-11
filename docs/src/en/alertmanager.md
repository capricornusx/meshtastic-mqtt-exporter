# AlertManager Integration

Send Prometheus alerts to LoRa mesh network via Meshtastic devices.

## AlertManager Setup

### 1. AlertManager Configuration

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
    http_config:
      basic_auth:
        username: 'alertmanager'
        password: 'secret123'
```

### 2. Exporter Configuration

```yaml
# config.yaml
alertmanager:
  enabled: true
  channel: "LongFast"
  mode: "broadcast"
  target_nodes:
    - "ffffffff"
  topics:
    broadcast: "msh/2/c/%s/!broadcast"
    direct: "msh/2/c/%s/!%s"
  http:
    port: 8080
    path: "/alerts/webhook"
  routing:
    critical:
      channel: "ShortFast"
      mode: "direct"
      target_nodes:
        - "admin001"
        - "ops002"
    warning:
      channel: "LongFast"
      mode: "broadcast"
```

## Prometheus Rules

### Basic Rules

```yaml
# meshtastic.rules.yml
groups:
- name: meshtastic.rules
  rules:
  - alert: NodeOffline
    expr: (time() - meshtastic_node_last_seen_timestamp) > 1200
    for: 5m
    labels:
      severity: warning
      service: meshtastic
    annotations:
      summary: "Meshtastic node {{ $labels.node_name }} is offline"
      description: "Node {{ $labels.node_id }} has been unresponsive for {{ $value }} seconds"
      
  - alert: LowBattery
    expr: meshtastic_battery_level_percent < 20
    for: 2m
    labels:
      severity: critical
      service: meshtastic
    annotations:
      summary: "Low battery: {{ $labels.node_name }}"
      description: "Node {{ $labels.node_id }} battery level is {{ $value }}%"
      
  - alert: HighTemperature
    expr: meshtastic_temperature_celsius > 50
    for: 5m
    labels:
      severity: warning
      service: meshtastic
    annotations:
      summary: "High temperature: {{ $labels.node_name }}"
      description: "Node {{ $labels.node_id }} temperature is {{ $value }}°C"
      
  - alert: NetworkPartition
    expr: count(meshtastic_node_last_seen_timestamp) < 3
    for: 10m
    labels:
      severity: critical
      service: meshtastic
    annotations:
      summary: "Mesh network partition"
      description: "Active nodes in network: {{ $value }}"
```

### Advanced Rules

```yaml
# advanced.rules.yml
groups:
- name: meshtastic.advanced
  rules:
  - alert: BatteryDrainRate
    expr: |
      (
        meshtastic_battery_level_percent - 
        meshtastic_battery_level_percent offset 1h
      ) < -10
    for: 15m
    labels:
      severity: warning
    annotations:
      summary: "Fast battery drain: {{ $labels.node_name }}"
      description: "Battery drained {{ $value }}% in one hour"
      
  - alert: TemperatureSpike
    expr: |
      abs(
        meshtastic_temperature_celsius - 
        avg_over_time(meshtastic_temperature_celsius[1h])
      ) > 15
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Temperature spike: {{ $labels.node_name }}"
      
  - alert: HumidityAnomaly
    expr: |
      abs(
        meshtastic_humidity_percent - 
        avg_over_time(meshtastic_humidity_percent[6h])
      ) > 30
    for: 10m
    labels:
      severity: info
    annotations:
      summary: "Humidity anomaly: {{ $labels.node_name }}"
```

## Delivery Modes

### Broadcast Mode

Sends alerts to all nodes in mesh network:

```yaml
alertmanager:
  mode: "broadcast"
  channel: "LongFast"
```

Message will be sent to topic: `msh/2/c/LongFast/!broadcast`

### Direct Mode

Sends alerts to specific nodes:

```yaml
alertmanager:
  mode: "direct"
  channel: "ShortFast"
  target_nodes:
    - "admin001"
    - "ops002"
    - "12345678"
```

Messages will be sent to topics:
- `msh/2/c/ShortFast/!admin001`
- `msh/2/c/ShortFast/!ops002`
- `msh/2/c/ShortFast/!12345678`

## Severity-based Routing

### Routing Configuration

```yaml
alertmanager:
  enabled: true
  # Default settings
  channel: "LongFast"
  mode: "broadcast"
  target_nodes: ["ffffffff"]
  
  # Severity-based routing
  routing:
    critical:
      channel: "ShortFast"
      mode: "direct"
      target_nodes:
        - "admin001"
        - "ops002"
    warning:
      channel: "LongFast"
      mode: "broadcast"
    info:
      channel: "LongSlow"
      mode: "broadcast"
```

### Routing Logic

1. **critical** → Direct delivery to admins via ShortFast
2. **warning** → Broadcast to all nodes via LongFast
3. **info** → Broadcast via LongSlow for power saving
4. **No severity** → Use default settings

## Message Formatting

### Message Templates

```yaml
alertmanager:
  message_templates:
    critical: "🚨 {{ .CommonLabels.alertname }}: {{ .CommonAnnotations.summary }}"
    warning: "⚠️ {{ .CommonLabels.alertname }}: {{ .CommonAnnotations.summary }}"
    info: "ℹ️ {{ .CommonLabels.alertname }}: {{ .CommonAnnotations.summary }}"
    resolved: "✅ Resolved: {{ .CommonLabels.alertname }}"
```

### Message Constraints

- Maximum length: 200 characters
- Unicode emoji support
- Automatic truncation of long messages

## Alert Examples

### Test Alert

```bash
curl -X POST http://localhost:8080/alerts/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "alerts": [{
      "status": "firing",
      "labels": {
        "alertname": "TestAlert",
        "severity": "info"
      },
      "annotations": {
        "summary": "Test message"
      }
    }]
  }'
```

### Low Battery Alert

```bash
curl -X POST http://localhost:8080/alerts/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "alerts": [{
      "status": "firing",
      "labels": {
        "alertname": "LowBattery",
        "severity": "critical",
        "node_id": "12345678",
        "node_name": "Gateway"
      },
      "annotations": {
        "summary": "Low battery: Gateway (15%)"
      }
    }]
  }'
```

## Monitoring AlertManager Integration

### Metrics

```
# Number of sent alerts
meshtastic_alertmanager_alerts_sent_total{severity="critical"} 5
meshtastic_alertmanager_alerts_sent_total{severity="warning"} 12

# Send errors
meshtastic_alertmanager_errors_total{type="mqtt_publish"} 2

# Last alert timestamp
meshtastic_alertmanager_last_alert_timestamp 1640995200
```

### Logs

```bash
# View AlertManager integration logs
journalctl -u mqtt-exporter -f | grep alertmanager

# Example logs
2024-01-15T10:30:00Z INFO AlertManager webhook received alert: NodeOffline
2024-01-15T10:30:01Z INFO Sending alert to LoRa network: channel=LongFast mode=broadcast
2024-01-15T10:30:02Z INFO Alert sent successfully: message_id=abc123
```

## Troubleshooting

### Alerts Not Reaching Devices

1. Check MQTT connection:
```bash
mosquitto_sub -h localhost -p 1883 -t "msh/2/c/+/!+"
```

2. Verify channel configuration in devices
3. Ensure devices are subscribed to correct topics

### Webhook Not Receiving Alerts

1. Check URL in AlertManager configuration
2. Verify port availability: `netstat -tlnp | grep 8080`
3. Check AlertManager logs: `journalctl -u alertmanager -f`

### Messages Being Truncated

1. Use shorter message templates
2. Configure prioritization of important information
3. Consider using codes/abbreviations for frequent alerts