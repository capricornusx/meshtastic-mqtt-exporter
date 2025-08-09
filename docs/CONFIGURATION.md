# Configuration Guide

## Basic Configuration

```yaml
mqtt:
  host: localhost
  port: 1883
  allow_anonymous: true

prometheus:
  enabled: true
  host: 0.0.0.0
  port: 8100
```

## Secure Configuration

```yaml
mqtt:
  host: localhost
  port: 1883
  allow_anonymous: false
  users:
    - username: "meshtastic"
      password: "secure123"

prometheus:
  enabled: true
  host: 127.0.0.1
  port: 8100

state:
  enabled: true
  file: "meshtastic_state.json"
```

## Configuration Options

### MQTT Section
- `host` - MQTT broker host
- `port` - MQTT broker port
- `username/password` - Single user credentials
- `tls` - Enable TLS encryption
- `allow_anonymous` - Allow anonymous connections
- `users` - Multiple user credentials
- `broker.max_inflight` - Max unacknowledged messages per client (default: 50)
- `broker.max_queued` - Max queued messages per client (default: 1000)
- `broker.retain_available` - Enable retained messages (default: true)
- `broker.max_packet_size` - Max packet size in bytes (default: 131072)
- `broker.keep_alive` - Keep alive interval in seconds (default: 60)

### Prometheus Section
- `enabled` - Enable Prometheus metrics
- `host` - Metrics server host
- `port` - Metrics server port

### State Section
- `enabled` - Enable state persistence
- `file` - State file path

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

### 4. Install Service
```bash
sudo cp docs/mqtt-exporter-embedded.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable mqtt-exporter-embedded
sudo systemctl start mqtt-exporter-embedded
```

### Security Features
The systemd service includes security hardening:
- Dedicated system user without shell access
- No privilege escalation
- Protected system and home directories
- Isolated temporary directory
- Restricted write access