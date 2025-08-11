# API Documentation

## Endpoints

### Prometheus Metrics
- **GET** `/metrics` - Возвращает метрики в формате Prometheus
- **GET** `/health` - Health check endpoint

### AlertManager Webhook  
- **POST** `/alerts/webhook` - Принимает алерты от AlertManager

## OpenAPI Specification

Полная спецификация API доступна в файле [api/openapi.yaml](../../api/openapi.yaml).

Для просмотра используйте:
```bash
# Swagger UI
docker run -p 8080:8080 -e SWAGGER_JSON=/api/openapi.yaml -v $(pwd)/api:/api swaggerapi/swagger-ui

# Redoc
npx redoc-cli serve api/openapi.yaml
```

## Примеры использования

### Получение метрик
```bash
curl http://localhost:8100/metrics
```

### Отправка алерта
```bash
curl -X POST http://localhost:8080/alerts/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "alerts": [{
      "status": "firing",
      "labels": {
        "alertname": "NodeDown",
        "severity": "critical"
      },
      "annotations": {
        "summary": "Node is offline"
      }
    }]
  }'
```