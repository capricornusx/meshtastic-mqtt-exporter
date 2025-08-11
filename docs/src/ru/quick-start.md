# Быстрый старт

## Установка

### Скачать готовый бинарник

```bash
# Linux AMD64
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-linux-amd64

# Linux ARM64
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-linux-arm64

# macOS
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-darwin-amd64

chmod +x mqtt-exporter-*
```

### Сборка из исходников

```bash
git clone https://github.com/capricornusx/meshtastic-mqtt-exporter.git
cd meshtastic-mqtt-exporter
make build
```

## Конфигурация

Создайте файл `config.yaml`:

```yaml
logging:
  level: "info"  # debug, info, warn, error, fatal

mqtt:
  host: 0.0.0.0
  port: 1883
  allow_anonymous: true
  debug:
    log_all_messages: false  # Логировать все MQTT сообщения

prometheus:
  enabled: true
  port: 8100
  topic:
    pattern: "msh/#"  # Паттерн MQTT топиков с поддержкой wildcards

alertmanager:
  enabled: false
```

## Запуск

### Embedded режим

```bash
./mqtt-exporter-embedded --config config.yaml
```

### Standalone режим

```bash
# Для подключения к внешнему MQTT брокеру
./mqtt-exporter-standalone --config config.yaml
```

## Проверка работы

### Метрики Prometheus

```bash
curl http://localhost:8100/metrics
```

### Health check

```bash
curl http://localhost:8100/health
```

### Логи

```bash
# Embedded режим с отладкой
./mqtt-exporter-embedded --config config.yaml --log-level debug
```

## Интеграция с Meshtastic

### Настройка устройства

1. Подключитесь к устройству через Meshtastic CLI или приложение
2. Настройте MQTT:

```bash
meshtastic --set mqtt.enabled true
meshtastic --set mqtt.address your-mqtt-server.com
meshtastic --set mqtt.username your-username
meshtastic --set mqtt.password your-password
meshtastic --set mqtt.encryption_enabled false
```

### Проверка топиков

Устройства должны публиковать в топики:
- `msh/2/c/LongFast/!<node_id>` — сообщения
- `msh/2/e/LongFast/!<node_id>` — телеметрия

## Docker

```bash
# Запуск с Docker
docker run -p 1883:1883 -p 8100:8100 -v $(pwd)/config.yaml:/config.yaml \
  ghcr.io/capricornusx/meshtastic-mqtt-exporter:latest --config /config.yaml
```

## Systemd сервис

```bash
# Копирование файлов
sudo cp mqtt-exporter-embedded /usr/local/bin/
sudo cp config.yaml /etc/mqtt-exporter/

# Создание сервиса
sudo cp docs/mqtt-exporter-embedded.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable mqtt-exporter-embedded
sudo systemctl start mqtt-exporter-embedded
```

## Troubleshooting

### Метрики не появляются

1. Проверьте префикс топиков в конфигурации
2. Убедитесь, что устройства публикуют данные
3. Проверьте логи: `journalctl -u mqtt-exporter-embedded -f`

### MQTT подключение не работает

1. Проверьте настройки брандмауэра: `sudo ufw allow 1883`
2. Проверьте привязку к интерфейсу: `netstat -tlnp | grep 1883`
3. Проверьте учетные данные в конфигурации

### Prometheus не собирает метрики

1. Добавьте job в `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'meshtastic'
    static_configs:
      - targets: ['localhost:8100']
```

2. Перезапустите Prometheus: `sudo systemctl restart prometheus`

