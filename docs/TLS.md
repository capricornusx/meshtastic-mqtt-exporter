# TLS поддержка

MQTT экспортер поддерживает TLS шифрование для безопасной передачи данных.

## Конфигурация

```yaml
mqtt:
  host: 0.0.0.0
  port: 1883  # TCP порт (всегда активен)
  tls_config:
    enabled: true  # Включить TLS listener
    port: 8883     # TLS порт
    cert_file: "certs/server.crt"
    key_file: "certs/server.key"
    ca_file: "certs/ca.crt"  # Опционально
```

## Режимы работы

- **TCP только**: `tls_config.enabled: false` - работает только порт 1883
- **TCP + TLS**: `tls_config.enabled: true` - работают оба порта 1883 и 8883

## Генерация сертификатов

Для тестирования используйте скрипт:

```bash
./scripts/generate-certs.sh
```

Это создаст:
- `certs/ca.crt` - Сертификат CA
- `certs/server.crt` - Сертификат сервера
- `certs/server.key` - Приватный ключ сервера

## Тестирование

```bash
# Запуск с TLS
./mqtt-exporter --config config.yaml

# Тест подключения
mosquitto_pub -h localhost -p 8883 \
  --cafile certs/ca.crt \
  -t "msh/test/2/json/test" \
  -m '{"test": "message"}' \
  -u admin -P admin
```

## Производственное использование

Для продакшена используйте сертификаты от доверенного CA:

1. Получите сертификат от CA (Let's Encrypt, внутренний CA)
2. Укажите пути к сертификатам в конфигурации
3. Настройте автообновление сертификатов

## Порты

- **1883** - Стандартный MQTT (без TLS)
- **8883** - Стандартный MQTT с TLS