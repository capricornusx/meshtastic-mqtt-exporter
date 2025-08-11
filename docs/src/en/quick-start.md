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

Create `config.yaml`:

```yaml
mqtt:
  host: 0.0.0.0
  port: 1883
  allow_anonymous: true

prometheus:
  enabled: true
  port: 8101
  topic:
    pattern: "msh/#"  # MQTT topic pattern with wildcards support

alertmanager:
  enabled: false
```

## Running

### Embedded Mode

```bash
./mqtt-exporter-embedded --config config.yaml
```

### Standalone Mode

```bash
# Connect to external MQTT broker
./mqtt-exporter-standalone --config config.yaml
```

## Verification

### Prometheus Metrics

```bash
curl http://localhost:8101/metrics
```

### Health Check

```bash
curl http://localhost:8101/health
```

### Logs

```bash
# Embedded mode with debug
./mqtt-exporter-embedded --config config.yaml --log-level debug
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
# Run with Docker
docker run -p 1883:1883 -p 8101:8101 -v $(pwd)/config.yaml:/config.yaml \
  ghcr.io/capricornusx/meshtastic-mqtt-exporter:latest --config /config.yaml
```

## Systemd Service

```bash
# Copy files
sudo cp mqtt-exporter-embedded /usr/local/bin/
sudo cp config.yaml /etc/mqtt-exporter/

# Create service
sudo cp docs/mqtt-exporter-embedded.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable mqtt-exporter-embedded
sudo systemctl start mqtt-exporter-embedded
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
      - targets: ['localhost:8101']
```

2. Restart Prometheus: `sudo systemctl restart prometheus`