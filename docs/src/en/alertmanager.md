# AlertManager Integration

Send Prometheus alerts to LoRa mesh network via Meshtastic devices.

## AlertManager Setup

### 1. AlertManager Configuration

Ready-to-use AlertManager configuration is available in [alertmanager.yml](../stack/alertmanager/alertmanager.yml).

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

Ready-to-use Prometheus alert rules are available in [meshtastic-alerts.yml](../stack/alertmanager/meshtastic-alerts.yml).

The file includes alerts for:
- Node offline detection
- Battery level monitoring
- Temperature thresholds
- Signal quality checks
- Service availability

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

## TODO

- [ ] Add MQTT-specific metrics for AlertManager integration monitoring

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