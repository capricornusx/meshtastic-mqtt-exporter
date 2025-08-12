# Quick Start

## Installation

### Download Binary

```bash
# Linux AMD64
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-linux-amd64

# Linux ARM64
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-linux-arm64

# macOS
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-darwin-amd64

chmod +x mqtt-exporter-*
```

### Build from Source

```bash
git clone https://github.com/capricornusx/meshtastic-mqtt-exporter.git
cd meshtastic-mqtt-exporter
make build
```

## Configuration

Download example configuration:

```bash
wget https://raw.githubusercontent.com/capricornusx/meshtastic-mqtt-exporter/main/config.yaml
```

Or create minimal `config.yaml`. See [Configuration](configuration.md) for details.

## Running

```bash
./mqtt-exporter-linux-amd64 --config config.yaml
```

## Verification

### Prometheus Metrics

```bash
curl http://localhost:8100/metrics
```

### Health Check

```bash
curl http://localhost:8100/health
```

### Debug

```bash
# With debug logs
./mqtt-exporter-linux-amd64 --config config.yaml --log-level debug
```

## Meshtastic Integration

### Device Setup

1. Connect to device via Meshtastic CLI or app
2. Configure MQTT:

```bash
meshtastic --set mqtt.enabled true
meshtastic --set mqtt.address your-mqtt-server.com
meshtastic --set mqtt.username your-username
meshtastic --set mqtt.password your-password
meshtastic --set mqtt.encryption_enabled false
```

### Topic Verification

Devices should publish to:
- `msh/2/c/LongFast/!<node_id>` — messages
- `msh/2/e/LongFast/!<node_id>` — telemetry

## Docker

```bash
docker run -p 1883:1883 -p 8100:8100 -v $(pwd)/config.yaml:/config.yaml \
  ghcr.io/capricornusx/meshtastic-mqtt-exporter:latest --config /config.yaml
```

## Troubleshooting

### No Metrics Appearing

1. Check topic prefix in configuration
2. Ensure devices are publishing data
3. Check logs: `journalctl -u mqtt-exporter-embedded -f`

### MQTT Connection Issues

1. Check firewall: `sudo ufw allow 1883`
2. Check interface binding: `netstat -tlnp | grep 1883`
3. Verify credentials in configuration

### Prometheus Not Scraping

1. Add job to `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'meshtastic'
    static_configs:
      - targets: ['localhost:8100']
```

2. Restart Prometheus: `sudo systemctl restart prometheus`