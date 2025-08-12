# API

## Endpoints

- **GET** `/metrics` — Prometheus metrics
- **GET** `/health` — Health check
- **POST** `/alerts/webhook` — AlertManager webhook

## Usage

### Metrics

```bash
curl http://localhost:8100/metrics
```

### Health Check

```bash
curl http://localhost:8100/health
```

Response:
```json
{
  "status": "ok",
  "uptime": "2h30m15s",
  "active_nodes": 5
}
```

## Metrics

| Metric | Description | Labels |
|--------|-------------|--------|
| `meshtastic_battery_level_percent` | Battery level | `node_id`, `node_name` |
| `meshtastic_temperature_celsius` | Temperature | `node_id`, `node_name` |
| `meshtastic_humidity_percent` | Humidity | `node_id`, `node_name` |
| `meshtastic_pressure_hpa` | Pressure | `node_id`, `node_name` |
| `meshtastic_node_last_seen_timestamp` | Last activity | `node_id`, `node_name` |

## Prometheus Integration

Ready-to-use Prometheus configuration is available in [prometheus.yml](../stack/prometheus/prometheus.yml).

## AlertManager Rules

Full set of alert rules is available in [meshtastic-alerts.yml](../stack/alertmanager/meshtastic-alerts.yml).

Includes alerts for:
- Node offline
- Low battery
- High temperature  
- Weak signal
- Exporter unavailable