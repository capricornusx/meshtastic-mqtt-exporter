# Развертывание

## Docker

### Готовый образ

```bash
docker run -p 1883:1883 -p 8100:8100 -v $(pwd)/config.yaml:/config.yaml \
  ghcr.io/capricornusx/meshtastic-mqtt-exporter:latest --config /config.yaml
```

### Docker Compose (полный стек)

```bash
cd docs/stack
docker-compose up -d
```

Включает: MQTT Exporter, Prometheus, Grafana, AlertManager

## Systemd

### Установка сервиса

```bash
# Скачать бинарник
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-linux-amd64
sudo mv mqtt-exporter-linux-amd64 /usr/local/bin/mqtt-exporter
sudo chmod +x /usr/local/bin/mqtt-exporter

# Создать конфигурацию
sudo mkdir -p /etc/mqtt-exporter
sudo wget -O /etc/mqtt-exporter/config.yaml \
  https://raw.githubusercontent.com/capricornusx/meshtastic-mqtt-exporter/main/config.yaml
```

### Systemd unit

```ini
[Unit]
Description=Meshtastic MQTT Exporter
After=network.target

[Service]
Type=simple
User=mqtt-exporter
Group=mqtt-exporter
ExecStart=/usr/local/bin/mqtt-exporter --config /etc/mqtt-exporter/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Запуск

```bash
# Создать пользователя
sudo useradd -r -s /bin/false mqtt-exporter

# Установить и запустить сервис
sudo systemctl daemon-reload
sudo systemctl enable mqtt-exporter
sudo systemctl start mqtt-exporter

# Проверить статус
sudo systemctl status mqtt-exporter
sudo journalctl -u mqtt-exporter -f
```

## Бинарник

### Установка

```bash
# Linux AMD64
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-linux-amd64
chmod +x mqtt-exporter-linux-amd64

# Запуск
./mqtt-exporter-linux-amd64 --config config.yaml
```

### Сборка из исходников

```bash
git clone https://github.com/capricornusx/meshtastic-mqtt-exporter.git
cd meshtastic-mqtt-exporter
make build
```

## Проверка работы

```bash
# Метрики
curl http://localhost:8100/metrics

# Health check
curl http://localhost:8100/health

# Логи
journalctl -u mqtt-exporter -f
```
