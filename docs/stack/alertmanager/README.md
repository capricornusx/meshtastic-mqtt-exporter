# AlertManager Configuration Examples

Примеры конфигурации для интеграции с AlertManager.

## Файлы

- `alertmanager.yml` - конфигурация AlertManager
- `meshtastic-alerts.yml` - правила Prometheus

## Установка

```bash
# Скопировать конфигурацию
cp alertmanager.yml /etc/alertmanager/
cp meshtastic-alerts.yml /etc/prometheus/rules/

# Перезапустить сервисы
sudo systemctl restart alertmanager prometheus
```

Подробная документация: [../ALERTMANAGER.md](../en/ALERTMANAGER.md)