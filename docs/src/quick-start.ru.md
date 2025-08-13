# Быстрый старт

## Установка

```bash
# Скачать бинарник
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-linux-amd64
chmod +x mqtt-exporter-linux-amd64

# Скачать конфигурацию
wget https://raw.githubusercontent.com/capricornusx/meshtastic-mqtt-exporter/main/config.yaml

# Запустить
./mqtt-exporter-linux-amd64 --config config.yaml
```

## Проверка

```bash
# Метрики
curl http://localhost:8100/metrics

# Состояние
curl http://localhost:8100/health

# Отладка
./mqtt-exporter-linux-amd64 --config config.yaml --log-level debug
```

## Настройка Meshtastic

```bash
meshtastic --set mqtt.enabled true
meshtastic --set mqtt.address your-server.com
meshtastic --set mqtt.username meshtastic
meshtastic --set mqtt.password password
```

Устройства публикуют в:
- `msh/2/c/LongFast/!<node_id>` — сообщения
- `msh/2/e/LongFast/!<node_id>` — телеметрия

## Docker

```bash
# Отдельный контейнер
docker run -p 1883:1883 -p 8100:8100 -v $(pwd)/config.yaml:/config.yaml \
  ghcr.io/capricornusx/meshtastic-mqtt-exporter:latest --config /config.yaml

# Полный стек
cd docs/stack
docker-compose up -d
```

## Сборка из исходников

```bash
git clone https://github.com/capricornusx/meshtastic-mqtt-exporter.git
cd meshtastic-mqtt-exporter
make build
```

## Troubleshooting

**Метрики не появляются:**
- Проверьте паттерн топиков в конфиге
- Убедитесь, что устройства публикуют данные

**MQTT не работает:**
- Проверьте файрвол: `sudo ufw allow 1883`
- Проверьте порт: `netstat -tlnp | grep 1883`