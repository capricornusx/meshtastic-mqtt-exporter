#!/bin/env bash

# Включаем TLS в конфигурации
sed -i 's/enabled: false/enabled: true/' config.yaml

echo "Запуск MQTT сервера с TCP + TLS..."
./mqtt-exporter --config config.yaml &
SERVER_PID=$!

sleep 2

echo "Тестирование TCP подключения (порт 1883)..."
mosquitto_pub -h localhost -p 1883 -t "msh/test/2/json/tcp" -m '{"test": "tcp works"}' -u admin -P admin

echo "Тестирование TLS подключения с ca.crt (порт 8883)..."
mosquitto_pub -h localhost -p 8883 --cafile certs/ca.crt -t "msh/test/2/json/tls" -m '{"test": "tls with ca works"}' -u admin -P admin

echo "Тестирование TLS подключения без ca.crt (--insecure)..."
mosquitto_pub -h localhost -p 8883 --insecure -t "msh/test/2/json/tls" -m '{"test": "tls insecure works"}' -u admin -P admin

echo "Остановка сервера..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

# Возвращаем TLS в выключенное состояние
sed -i 's/enabled: true/enabled: false/' config.yaml

echo "Тест завершен"