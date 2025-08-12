# AlertManager Integration

Send Prometheus alerts to Meshtastic mesh networks via LoRa radio.

## Configuration

### Hook Mode

```go
hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
AlertManager: struct {
Enabled bool
Addr    string
Path    string
}{
Enabled: true,
Addr:    ":8080",
Path:    "/alerts/webhook",
},
})
```

### Embedded Mode

```yaml
alertmanager:
  enabled: true
  http:
    port: 8080
    path: "/alerts/webhook"
```

## Routing Configuration

```yaml
alertmanager:
  routing:
    critical:
      channel: "ShortFast"
      mode: "broadcast"
    warning:
      channel: "LongFast"
      mode: "direct"
      target_nodes: [ "admin001" ]
```

## Delivery Modes

| Mode        | Delivery       | Channel     |
|-------------|----------------|-------------|
| `broadcast` | All nodes      | `LongFast`  |
| `direct`    | Specific nodes | `ShortFast` |

## Setup

1. **AlertManager config**: [alertmanager.yml](../alertmanager/alertmanager.yml)
2. **Prometheus rules**: [meshtastic-alerts.yml](../alertmanager/meshtastic-alerts.yml)

## Alert Routing

- **Critical** → ShortFast, broadcast
- **Warning** → LongFast, direct to admins
- **Info** → LongSlow, broadcast

## Message Format

- 🚨 **Firing**: `🚨 firing: AlertName - Summary`
- ✅ **Resolved**: `✅ resolved: AlertName - Summary`

## Testing

```bash
# Test alert
curl -X POST http://localhost:8080/alerts/webhook \
  -H "Content-Type: application/json" \
  -d '{"alerts":[{"status":"firing","labels":{"alertname":"TestAlert"},"annotations":{"summary":"Test alert"}}]}'

# Check MQTT topics
mosquitto_sub -h localhost -t "msh/2/c/+/!+"
```