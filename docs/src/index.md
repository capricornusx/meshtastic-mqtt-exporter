# Meshtastic MQTT Exporter

Экспорт телеметрии Meshtastic устройств в метрики Prometheus с интеграцией AlertManager для отправки алертов в LoRa сеть.

## Возможности

- **Встроенный MQTT брокер** на основе mochi-mqtt
- **Поддержка TLS/QoS/Retention**
- **Prometheus метрики**: Батарея, температура, влажность, давление, качество сигнала
- **AlertManager интеграция**: Отправка алертов в LoRa mesh сеть
- **Персистентность состояния**: Сохранение/восстановление метрик между перезапусками

## Быстрый старт

### Docker Compose (полный стек)

```bash
# Полный стек мониторинга
cd docs/stack
docker-compose up -d
```

### Отдельный бинарник

```bash
# Скачать бинарник
wget https://github.com/capricornusx/meshtastic-mqtt-exporter/releases/latest/download/mqtt-exporter-linux-amd64

# Запустить embedded режим
./mqtt-exporter-linux-amd64 --config config.yaml

# Проверить метрики
curl http://localhost:8100/metrics
```
