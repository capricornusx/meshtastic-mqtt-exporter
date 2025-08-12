# Полный стек мониторинга Meshtastic

Готовая конфигурация для запуска полного стека мониторинга с MQTT экспортером, Prometheus, AlertManager и Grafana.

## Быстрый запуск

```bash
cd docs/stack
docker-compose up -d
```

## Сервисы

- **MQTT Exporter**: http://localhost:8100/metrics
- **Prometheus**: http://localhost:9090
- **AlertManager**: http://localhost:9093
- **Grafana**: http://localhost:3000 (admin/admin123)

## Структура

```
stack/
├── docker-compose.yml          # Основная конфигурация
├── config.yaml                 # Конфигурация MQTT экспортера
├── prometheus/
│   └── prometheus.yml          # Конфигурация Prometheus
├── alertmanager/
│   ├── alertmanager.yml        # Конфигурация AlertManager
│   └── meshtastic-alerts.yml   # Правила алертов
└── grafana/
    ├── datasources/
    │   └── prometheus.yml      # Datasource конфигурация
    └── dashboards/
        └── dashboard.yml       # Провайдер дашбордов
```

## Использование

1. Скопируйте папку `stack/` в ваш проект
2. Настройте `config.yaml` под ваши нужды
3. Запустите: `docker-compose up -d`
4. Подключите Meshtastic устройства к MQTT брокеру на порту 1883

## Персистентность данных

Все данные сохраняются в Docker volumes:
- `mqtt_data` - состояние MQTT экспортера
- `prometheus_data` - метрики Prometheus
- `alertmanager_data` - состояние AlertManager
- `grafana_data` - дашборды и настройки Grafana
