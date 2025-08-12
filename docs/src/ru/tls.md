# TLS поддержка

MQTT экспортер поддерживает одновременную работу TCP и TLS портов.

## Конфигурация

```yaml
mqtt:
  port: 1883  # TCP порт (всегда активен)
  tls_config:
    enabled: true  # Включить TLS порт
    port: 8883     # TLS порт
    cert_file: "certs/server.crt"
    key_file: "certs/server.key"
```

## Генерация сертификатов

```bash
./scripts/generate-certs.sh
```

## Подключение клиентов

- **TCP**: `mqtt://localhost:1883`
- **TLS**: `mqtts://localhost:8883`
