# Примеры использования

Данный раздел содержит практические примеры интеграции с различными системами.

> **Примечание:** Примеры Docker и Kubernetes развертывания находятся в разделе [Развертывание](../deployment/docker.md).

## Grafana Dashboard

Готовые dashboard для Grafana доступны в [stack/grafana/dashboards/](../stack/grafana/dashboards/)

**Основные панели:**
- Active Nodes — количество активных узлов
- Battery Levels — уровень заряда по узлам
- Temperature Trends — температурные тренды
- Network Map — карта сети

## Мониторинг и алерты

### Telegram бот для алертов

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

### Скрипт резервного копирования

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backup/mqtt-exporter"
DATE=$(date +%Y%m%d_%H%M%S)

# Создание директории
mkdir -p "$BACKUP_DIR"

# Резервное копирование конфигурации
cp /etc/mqtt-exporter/config.yaml "$BACKUP_DIR/config_$DATE.yaml"

# Резервное копирование состояния
if [ -f /var/lib/mqtt-exporter/meshtastic_state.json ]; then
    cp /var/lib/mqtt-exporter/meshtastic_state.json "$BACKUP_DIR/state_$DATE.json"
fi

# Экспорт метрик
curl -s http://localhost:8100/metrics > "$BACKUP_DIR/metrics_$DATE.txt"

# Очистка старых резервных копий (старше 30 дней)
find "$BACKUP_DIR" -name "*.yaml" -o -name "*.json" -o -name "*.txt" | \
    head -n -30 | xargs rm -f

echo "Резервное копирование завершено: $BACKUP_DIR"
```

## Интеграция с Home Assistant

### Конфигурация

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
        message: "Низкий заряд батареи в Meshtastic сети"
```

## Node-RED интеграция

### Flow для обработки алертов

```json
[
  {
    "id": "mqtt-input",
    "type": "mqtt in",
    "topic": "msh/+/+/+/+",
    "qos": "0",
    "broker": "mqtt-broker",
    "x": 100,
    "y": 100
  },
  {
    "id": "parse-meshtastic",
    "type": "function",
    "func": "const topic = msg.topic.split('/');\nconst nodeId = topic[4];\nconst payload = JSON.parse(msg.payload);\n\nmsg.nodeId = nodeId;\nmsg.telemetry = payload;\n\nreturn msg;",
    "x": 300,
    "y": 100
  },
  {
    "id": "low-battery-check",
    "type": "switch",
    "property": "telemetry.battery_level",
    "rules": [
      {
        "t": "lt",
        "v": "20"
      }
    ],
    "x": 500,
    "y": 100
  },
  {
    "id": "send-alert",
    "type": "http request",
    "method": "POST",
    "url": "http://localhost:8080/alerts/webhook",
    "headers": {"Content-Type": "application/json"},
    "x": 700,
    "y": 100
  }
]
```

## Тестирование

### Скрипт нагрузочного тестирования

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
    
    # Создание 10 виртуальных узлов
    threads = []
    for i in range(10):
        node_id = f"test{i:04d}"
        thread = Thread(target=publish_messages, args=(client, node_id, 100))
        threads.append(thread)
        thread.start()
    
    # Ожидание завершения
    for thread in threads:
        thread.join()
    
    client.loop_stop()
    client.disconnect()

if __name__ == "__main__":
    main()
```

### Проверка метрик

```bash
#!/bin/bash
# test_metrics.sh

echo "Проверка доступности сервиса..."
curl -f http://localhost:8100/health || exit 1

echo "Проверка метрик..."
METRICS=$(curl -s http://localhost:8100/metrics)

if echo "$METRICS" | grep -q "meshtastic_"; then
    echo "✅ Метрики Meshtastic найдены"
else
    echo "❌ Метрики Meshtastic не найдены"
    exit 1
fi

echo "Количество активных узлов:"
echo "$METRICS" | grep "meshtastic_node_last_seen_timestamp" | wc -l

echo "Средний уровень батареи:"
echo "$METRICS" | grep "meshtastic_battery_level_percent" | \
    awk '{sum+=$2; count++} END {if(count>0) print sum/count "%"}'
```

## Интеграция с внешними системами

### Webhook для Discord

```python
#!/usr/bin/env python3
# discord_webhook.py

import requests
import json
from flask import Flask, request

app = Flask(__name__)

DISCORD_WEBHOOK_URL = "https://discord.com/api/webhooks/YOUR_WEBHOOK_URL"

@app.route('/discord', methods=['POST'])
def discord_webhook():
    data = request.json
    
    for alert in data.get('alerts', []):
        status = alert.get('status', 'unknown')
        alertname = alert.get('labels', {}).get('alertname', 'Unknown')
        summary = alert.get('annotations', {}).get('summary', 'No summary')
        
        color = 0xff0000 if status == 'firing' else 0x00ff00
        
        discord_data = {
            "embeds": [{
                "title": f"Meshtastic Alert: {alertname}",
                "description": summary,
                "color": color,
                "timestamp": alert.get('startsAt', '')
            }]
        }
        
        requests.post(DISCORD_WEBHOOK_URL, json=discord_data)
    
    return "OK"

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5001)
```

### Интеграция с InfluxDB

```python
#!/usr/bin/env python3
# influxdb_exporter.py

import paho.mqtt.client as mqtt
import json
from influxdb_client import InfluxDBClient, Point
from influxdb_client.client.write_api import SYNCHRONOUS

# InfluxDB конфигурация
INFLUXDB_URL = "http://localhost:8086"
INFLUXDB_TOKEN = "your-token"
INFLUXDB_ORG = "your-org"
INFLUXDB_BUCKET = "meshtastic"

client = InfluxDBClient(url=INFLUXDB_URL, token=INFLUXDB_TOKEN, org=INFLUXDB_ORG)
write_api = client.write_api(write_options=SYNCHRONOUS)

def on_message(client, userdata, msg):
    try:
        topic_parts = msg.topic.split('/')
        if len(topic_parts) >= 5:
            node_id = topic_parts[4].replace('!', '')
            data = json.loads(msg.payload.decode())
            
            point = Point("meshtastic_telemetry") \
                .tag("node_id", node_id) \
                .field("battery_level", data.get("battery_level", 0)) \
                .field("temperature", data.get("temperature", 0)) \
                .field("humidity", data.get("humidity", 0)) \
                .field("pressure", data.get("pressure", 0))
            
            write_api.write(bucket=INFLUXDB_BUCKET, org=INFLUXDB_ORG, record=point)
            
    except Exception as e:
        print(f"Error processing message: {e}")

mqtt_client = mqtt.Client()
mqtt_client.on_message = on_message
mqtt_client.connect("localhost", 1883, 60)
mqtt_client.subscribe("msh/+/+/+/+")
mqtt_client.loop_forever()
```