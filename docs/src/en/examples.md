# Usage Examples

## Docker Compose

### Complete Monitoring Stack

```yaml
# docker-compose.yml
version: '3.8'

services:
  mqtt-exporter:
    image: ghcr.io/capricornusx/meshtastic-mqtt-exporter:latest
    ports:
      - "1883:1883"
      - "8100:8100"
      - "8080:8080"
    volumes:
      - ./config.yaml:/config.yaml
      - mqtt_data:/data
    command: --config /config.yaml
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - ./rules:/etc/prometheus/rules
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--web.enable-lifecycle'
    restart: unless-stopped

  alertmanager:
    image: prom/alertmanager:latest
    ports:
      - "9093:9093"
    volumes:
      - ./alertmanager.yml:/etc/alertmanager/alertmanager.yml
      - alertmanager_data:/alertmanager
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./grafana/datasources:/etc/grafana/provisioning/datasources
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin123
    restart: unless-stopped

volumes:
  mqtt_data:
  prometheus_data:
  alertmanager_data:
  grafana_data:
```

### Prometheus Configuration

Ready-to-use Prometheus configuration is available in [prometheus.yml](../stack/prometheus/prometheus.yml).

## Grafana Dashboard

### Main Panel

```json
{
  "dashboard": {
    "id": null,
    "title": "Meshtastic Network Overview",
    "tags": ["meshtastic", "lora"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Active Nodes",
        "type": "stat",
        "targets": [
          {
            "expr": "count(meshtastic_node_last_seen_timestamp)",
            "legendFormat": "Active Nodes"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "color": {
              "mode": "thresholds"
            },
            "thresholds": {
              "steps": [
                {"color": "red", "value": 0},
                {"color": "yellow", "value": 1},
                {"color": "green", "value": 3}
              ]
            }
          }
        },
        "gridPos": {"h": 8, "w": 6, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Battery Levels",
        "type": "bargauge",
        "targets": [
          {
            "expr": "meshtastic_battery_level_percent",
            "legendFormat": "{{node_name}}"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "min": 0,
            "max": 100,
            "unit": "percent",
            "thresholds": {
              "steps": [
                {"color": "red", "value": 0},
                {"color": "yellow", "value": 20},
                {"color": "green", "value": 50}
              ]
            }
          }
        },
        "gridPos": {"h": 8, "w": 6, "x": 6, "y": 0}
      },
      {
        "id": 3,
        "title": "Temperature",
        "type": "timeseries",
        "targets": [
          {
            "expr": "meshtastic_temperature_celsius",
            "legendFormat": "{{node_name}}"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "celsius"
          }
        },
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8}
      },
      {
        "id": 4,
        "title": "Network Map",
        "type": "geomap",
        "targets": [
          {
            "expr": "meshtastic_node_last_seen_timestamp",
            "legendFormat": "{{node_name}}"
          }
        ],
        "gridPos": {"h": 12, "w": 24, "x": 0, "y": 16}
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "30s"
  }
}
```

## Kubernetes Deployment

### Namespace and ConfigMap

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: meshtastic

---
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mqtt-exporter-config
  namespace: meshtastic
data:
  config.yaml: |
    mqtt:
      host: 0.0.0.0
      port: 1883
      allow_anonymous: true
    prometheus:
      enabled: true
      port: 8100
      topic:
        prefix: "msh/"
    alertmanager:
      enabled: true
      http:
        port: 8080
        path: "/alerts/webhook"
```

### Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mqtt-exporter
  namespace: meshtastic
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mqtt-exporter
  template:
    metadata:
      labels:
        app: mqtt-exporter
    spec:
      containers:
      - name: mqtt-exporter
        image: ghcr.io/capricornusx/meshtastic-mqtt-exporter:latest
        ports:
        - containerPort: 1883
          name: mqtt
        - containerPort: 8100
          name: metrics
        - containerPort: 8080
          name: webhook
        volumeMounts:
        - name: config
          mountPath: /config.yaml
          subPath: config.yaml
        args: ["--config", "/config.yaml"]
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
      volumes:
      - name: config
        configMap:
          name: mqtt-exporter-config

---
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: mqtt-exporter
  namespace: meshtastic
spec:
  selector:
    app: mqtt-exporter
  ports:
  - name: mqtt
    port: 1883
    targetPort: 1883
  - name: metrics
    port: 8100
    targetPort: 8100
  - name: webhook
    port: 8080
    targetPort: 8080
  type: LoadBalancer
```

## Systemd Service

### Installation

```bash
#!/bin/bash
# install.sh

# Create user
sudo useradd --system --no-create-home --shell /bin/false mqtt-exporter

# Create directories
sudo mkdir -p /opt/mqtt-exporter /etc/mqtt-exporter /var/lib/mqtt-exporter
sudo chown mqtt-exporter:mqtt-exporter /var/lib/mqtt-exporter

# Copy files
sudo cp mqtt-exporter-embedded /opt/mqtt-exporter/
sudo cp config.yaml /etc/mqtt-exporter/
sudo chmod 755 /opt/mqtt-exporter/mqtt-exporter-embedded

# Create service
cat << 'EOF' | sudo tee /etc/systemd/system/mqtt-exporter.service
[Unit]
Description=Meshtastic MQTT Exporter
After=network.target

[Service]
Type=simple
User=mqtt-exporter
Group=mqtt-exporter
ExecStart=/opt/mqtt-exporter/mqtt-exporter-embedded --config /etc/mqtt-exporter/config.yaml
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
EOF

# Start service
sudo systemctl daemon-reload
sudo systemctl enable mqtt-exporter
sudo systemctl start mqtt-exporter

echo "Installation complete. Check status: sudo systemctl status mqtt-exporter"
```

## Monitoring and Alerts

### Telegram Bot for Alerts

```python
#!/usr/bin/env python3
# telegram_bot.py

import requests
import json
from flask import Flask, request

app = Flask(__name__)

TELEGRAM_BOT_TOKEN = "your_bot_token"
TELEGRAM_CHAT_ID = "your_chat_id"

def send_telegram_message(message):
    url = f"https://api.telegram.org/bot{TELEGRAM_BOT_TOKEN}/sendMessage"
    data = {
        "chat_id": TELEGRAM_CHAT_ID,
        "text": message,
        "parse_mode": "Markdown"
    }
    requests.post(url, data=data)

@app.route('/webhook', methods=['POST'])
def webhook():
    data = request.json
    
    for alert in data.get('alerts', []):
        status = alert.get('status', 'unknown')
        alertname = alert.get('labels', {}).get('alertname', 'Unknown')
        summary = alert.get('annotations', {}).get('summary', 'No summary')
        
        if status == 'firing':
            emoji = "🚨" if alert.get('labels', {}).get('severity') == 'critical' else "⚠️"
        else:
            emoji = "✅"
            
        message = f"{emoji} *{alertname}*\n{summary}"
        send_telegram_message(message)
    
    return "OK"

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
```

### Backup Script

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backup/mqtt-exporter"
DATE=$(date +%Y%m%d_%H%M%S)

# Create directory
mkdir -p "$BACKUP_DIR"

# Backup configuration
cp /etc/mqtt-exporter/config.yaml "$BACKUP_DIR/config_$DATE.yaml"

# Backup state
if [ -f /var/lib/mqtt-exporter/meshtastic_state.json ]; then
    cp /var/lib/mqtt-exporter/meshtastic_state.json "$BACKUP_DIR/state_$DATE.json"
fi

# Export metrics
curl -s http://localhost:8100/metrics > "$BACKUP_DIR/metrics_$DATE.txt"

# Clean old backups (older than 30 days)
find "$BACKUP_DIR" -name "*.yaml" -o -name "*.json" -o -name "*.txt" | \
    head -n -30 | xargs rm -f

echo "Backup completed: $BACKUP_DIR"
```

## Home Assistant Integration

### Configuration

```yaml
# configuration.yaml
sensor:
  - platform: prometheus
    host: localhost
    port: 9090
    queries:
      - name: "Meshtastic Active Nodes"
        query: 'count(meshtastic_node_last_seen_timestamp)'
      - name: "Meshtastic Battery Average"
        query: 'avg(meshtastic_battery_level_percent)'

  - platform: rest
    resource: http://localhost:8100/health
    name: "MQTT Exporter Status"
    value_template: "{{ value_json.status }}"

automation:
  - alias: "Meshtastic Low Battery Alert"
    trigger:
      platform: numeric_state
      entity_id: sensor.meshtastic_battery_average
      below: 20
    action:
      service: notify.mobile_app
      data:
        message: "Low battery in Meshtastic network"
```

## Testing

### Load Testing Script

```python
#!/usr/bin/env python3
# load_test.py

import paho.mqtt.client as mqtt
import json
import time
import random
from threading import Thread

def generate_telemetry():
    return {
        "battery_level": random.randint(10, 100),
        "temperature": random.uniform(15.0, 35.0),
        "humidity": random.uniform(30.0, 80.0),
        "pressure": random.uniform(990.0, 1030.0)
    }

def publish_messages(client, node_id, count):
    for i in range(count):
        topic = f"msh/2/e/LongFast/!{node_id}"
        payload = json.dumps(generate_telemetry())
        client.publish(topic, payload)
        time.sleep(0.1)

def main():
    client = mqtt.Client()
    client.connect("localhost", 1883, 60)
    client.loop_start()
    
    # Create 10 virtual nodes
    threads = []
    for i in range(10):
        node_id = f"test{i:04d}"
        thread = Thread(target=publish_messages, args=(client, node_id, 100))
        threads.append(thread)
        thread.start()
    
    # Wait for completion
    for thread in threads:
        thread.join()
    
    client.loop_stop()
    client.disconnect()

if __name__ == "__main__":
    main()
```