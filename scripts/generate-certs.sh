#!/bin/bash

# Создание директории для сертификатов
mkdir -p certs

# Генерация приватного ключа CA
openssl genrsa -out certs/ca.key 4096

# Генерация сертификата CA
openssl req -new -x509 -days 365 -key certs/ca.key -out certs/ca.crt -subj "/C=RU/ST=Moscow/L=Moscow/O=MeshtasticExporter/CN=ca"

# Генерация приватного ключа сервера
openssl genrsa -out certs/server.key 4096

# Генерация запроса на сертификат сервера
openssl req -new -key certs/server.key -out certs/server.csr -subj "/C=RU/ST=Moscow/L=Moscow/O=MeshtasticExporter/CN=localhost"

# Подписание сертификата сервера CA
openssl x509 -req -days 365 -in certs/server.csr -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -out certs/server.crt

# Удаление временного файла
rm certs/server.csr

echo "Сертификаты созданы в директории certs/"
echo "CA: certs/ca.crt"
echo "Server cert: certs/server.crt"
echo "Server key: certs/server.key"