# TLS Configuration

## Overview

MQTT Exporter supports TLS encryption for secure MQTT connections. Both TCP (1883) and TLS (8883) ports can run simultaneously.

## Configuration

Full TLS configuration available in [config.yaml](../../config.yaml).

**Critical parameters:**
- `tls_config.enabled: true` — enable TLS support
- `tls_config.port: 8883` — TLS port (standard MQTT over TLS)
- `cert_file`, `key_file`, `ca_file` — certificate paths

## Certificate Generation

```bash
# Generate CA certificate
openssl genrsa -out ca.key 4096
openssl req -new -x509 -days 365 -key ca.key -out ca.crt

# Generate server certificate
openssl genrsa -out server.key 4096
openssl req -new -key server.key -out server.csr
openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key -out server.crt
```

## Client Connection

```bash
# TLS connection
mosquitto_pub -h localhost -p 8883 --cafile ca.crt -t "test/topic" -m "test message"

# TCP connection (insecure)
mosquitto_pub -h localhost -p 1883 -t "test/topic" -m "test message"
```