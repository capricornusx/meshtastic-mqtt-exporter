# Systemd развертывание

## Автоматическая установка

```bash
#!/bin/bash
# install.sh

set -e

# Переменные
USER="mqtt-exporter"
INSTALL_DIR="/opt/mqtt-exporter"
CONFIG_DIR="/etc/mqtt-exporter"
DATA_DIR="/var/lib/mqtt-exporter"
LOG_DIR="/var/log/mqtt-exporter"

# Создание пользователя
echo "Создание системного пользователя..."
sudo useradd --system --no-create-home --shell /bin/false $USER || true

# Создание директорий
echo "Создание директорий..."
sudo mkdir -p $INSTALL_DIR $CONFIG_DIR $DATA_DIR $LOG_DIR
sudo chown $USER:$USER $DATA_DIR $LOG_DIR

# Копирование файлов
echo "Копирование файлов..."
sudo cp mqtt-exporter-embedded $INSTALL_DIR/
sudo cp config.yaml $CONFIG_DIR/
sudo chmod 755 $INSTALL_DIR/mqtt-exporter-embedded

# Создание сервиса
echo "Создание systemd сервиса..."
cat << 'EOF' | sudo tee /etc/systemd/system/mqtt-exporter.service
[Unit]
Description=Meshtastic MQTT Exporter
Documentation=https://github.com/capricornusx/meshtastic-mqtt-exporter
After=network.target
Wants=network.target

[Service]
Type=simple
User=mqtt-exporter
Group=mqtt-exporter
ExecStart=/opt/mqtt-exporter/mqtt-exporter-embedded --config /etc/mqtt-exporter/config.yaml
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=5
WorkingDirectory=/var/lib/mqtt-exporter

# Логирование
StandardOutput=journal
StandardError=journal
SyslogIdentifier=mqtt-exporter

# Безопасность
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
PrivateDevices=yes
ProtectKernelTunables=yes
ProtectKernelModules=yes
ProtectControlGroups=yes
ReadWritePaths=/var/lib/mqtt-exporter /var/log/mqtt-exporter
RestrictSUIDSGID=yes
RestrictRealtime=yes
RestrictNamespaces=yes
LockPersonality=yes
MemoryDenyWriteExecute=yes

# Ограничения ресурсов
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF

# Перезагрузка systemd
echo "Перезагрузка systemd..."
sudo systemctl daemon-reload

# Включение и запуск сервиса
echo "Включение и запуск сервиса..."
sudo systemctl enable mqtt-exporter
sudo systemctl start mqtt-exporter

echo "Установка завершена!"
echo "Проверьте статус: sudo systemctl status mqtt-exporter"
echo "Просмотр логов: sudo journalctl -u mqtt-exporter -f"
```

## Ручная установка

### 1. Создание пользователя

```bash
sudo useradd --system --no-create-home --shell /bin/false mqtt-exporter
```

### 2. Создание директорий

```bash
sudo mkdir -p /opt/mqtt-exporter /etc/mqtt-exporter /var/lib/mqtt-exporter /var/log/mqtt-exporter
sudo chown mqtt-exporter:mqtt-exporter /var/lib/mqtt-exporter /var/log/mqtt-exporter
```

### 3. Установка бинарника

```bash
sudo cp mqtt-exporter-embedded /opt/mqtt-exporter/
sudo chown root:root /opt/mqtt-exporter/mqtt-exporter-embedded
sudo chmod 755 /opt/mqtt-exporter/mqtt-exporter-embedded
```

### 4. Конфигурация

```bash
sudo cp config.yaml /etc/mqtt-exporter/
sudo chown root:root /etc/mqtt-exporter/config.yaml
sudo chmod 644 /etc/mqtt-exporter/config.yaml
```

### 5. Systemd сервис

```ini
# /etc/systemd/system/mqtt-exporter.service
[Unit]
Description=Meshtastic MQTT Exporter
Documentation=https://github.com/capricornusx/meshtastic-mqtt-exporter
After=network.target
Wants=network.target

[Service]
Type=simple
User=mqtt-exporter
Group=mqtt-exporter
ExecStart=/opt/mqtt-exporter/mqtt-exporter-embedded --config /etc/mqtt-exporter/config.yaml
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=5
WorkingDirectory=/var/lib/mqtt-exporter

# Логирование
StandardOutput=journal
StandardError=journal
SyslogIdentifier=mqtt-exporter

# Безопасность
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
PrivateDevices=yes
ProtectKernelTunables=yes
ProtectKernelModules=yes
ProtectControlGroups=yes
ReadWritePaths=/var/lib/mqtt-exporter /var/log/mqtt-exporter
RestrictSUIDSGID=yes
RestrictRealtime=yes
RestrictNamespaces=yes
LockPersonality=yes
MemoryDenyWriteExecute=yes

# Ограничения ресурсов
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
```

## Управление сервисом

### Основные команды

```bash
# Включение автозапуска
sudo systemctl enable mqtt-exporter

# Запуск сервиса
sudo systemctl start mqtt-exporter

# Остановка сервиса
sudo systemctl stop mqtt-exporter

# Перезапуск сервиса
sudo systemctl restart mqtt-exporter

# Перезагрузка конфигурации
sudo systemctl reload mqtt-exporter

# Статус сервиса
sudo systemctl status mqtt-exporter

# Отключение автозапуска
sudo systemctl disable mqtt-exporter
```

### Просмотр логов

```bash
# Все логи
sudo journalctl -u mqtt-exporter

# Последние логи
sudo journalctl -u mqtt-exporter -n 50

# Следить за логами
sudo journalctl -u mqtt-exporter -f

# Логи за сегодня
sudo journalctl -u mqtt-exporter --since today

# Логи с определенного времени
sudo journalctl -u mqtt-exporter --since "2024-01-01 10:00:00"
```

## Мониторинг

### Health check скрипт

```bash
#!/bin/bash
# /usr/local/bin/mqtt-exporter-health.sh

HEALTH_URL="http://localhost:8101/health"
METRICS_URL="http://localhost:8101/metrics"

# Проверка health endpoint
if ! curl -f -s $HEALTH_URL > /dev/null; then
    echo "Health check failed"
    exit 1
fi

# Проверка метрик
if ! curl -f -s $METRICS_URL | grep -q "meshtastic_"; then
    echo "Metrics check failed"
    exit 1
fi

echo "Service is healthy"
exit 0
```

### Cron мониторинг

```bash
# Добавить в crontab
# crontab -e

# Проверка каждые 5 минут
*/5 * * * * /usr/local/bin/mqtt-exporter-health.sh || systemctl restart mqtt-exporter
```

## Логирование

### Настройка rsyslog

```bash
# /etc/rsyslog.d/mqtt-exporter.conf
if $programname == 'mqtt-exporter' then /var/log/mqtt-exporter/mqtt-exporter.log
& stop
```

### Ротация логов

```bash
# /etc/logrotate.d/mqtt-exporter
/var/log/mqtt-exporter/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 mqtt-exporter mqtt-exporter
    postrotate
        systemctl reload mqtt-exporter
    endscript
}
```

## Обновление

### Скрипт обновления

```bash
#!/bin/bash
# update.sh

set -e

INSTALL_DIR="/opt/mqtt-exporter"
BACKUP_DIR="/opt/mqtt-exporter/backup"
NEW_BINARY="mqtt-exporter-embedded"

# Создание резервной копии
echo "Создание резервной копии..."
sudo mkdir -p $BACKUP_DIR
sudo cp $INSTALL_DIR/mqtt-exporter-embedded $BACKUP_DIR/mqtt-exporter-embedded.$(date +%Y%m%d_%H%M%S)

# Остановка сервиса
echo "Остановка сервиса..."
sudo systemctl stop mqtt-exporter

# Обновление бинарника
echo "Обновление бинарника..."
sudo cp $NEW_BINARY $INSTALL_DIR/mqtt-exporter-embedded
sudo chmod 755 $INSTALL_DIR/mqtt-exporter-embedded

# Запуск сервиса
echo "Запуск сервиса..."
sudo systemctl start mqtt-exporter

# Проверка статуса
echo "Проверка статуса..."
sleep 5
sudo systemctl status mqtt-exporter

echo "Обновление завершено!"
```

## Troubleshooting

### Проблемы с запуском

```bash
# Проверка синтаксиса сервиса
sudo systemd-analyze verify /etc/systemd/system/mqtt-exporter.service

# Проверка зависимостей
sudo systemctl list-dependencies mqtt-exporter

# Проверка конфигурации
sudo -u mqtt-exporter /opt/mqtt-exporter/mqtt-exporter-embedded --config /etc/mqtt-exporter/config.yaml --validate
```

### Проблемы с правами

```bash
# Проверка владельца файлов
ls -la /opt/mqtt-exporter/
ls -la /etc/mqtt-exporter/
ls -la /var/lib/mqtt-exporter/

# Исправление прав
sudo chown -R mqtt-exporter:mqtt-exporter /var/lib/mqtt-exporter
sudo chown -R mqtt-exporter:mqtt-exporter /var/log/mqtt-exporter
```

### Проблемы с сетью

```bash
# Проверка портов
sudo netstat -tlnp | grep mqtt-exporter
sudo ss -tlnp | grep mqtt-exporter

# Проверка брандмауэра
sudo ufw status
sudo iptables -L

# Тестирование подключения
curl http://localhost:8101/health
curl http://localhost:8101/metrics
```

## Удаление

```bash
#!/bin/bash
# uninstall.sh

# Остановка и отключение сервиса
sudo systemctl stop mqtt-exporter
sudo systemctl disable mqtt-exporter

# Удаление файлов сервиса
sudo rm -f /etc/systemd/system/mqtt-exporter.service
sudo systemctl daemon-reload

# Удаление файлов приложения
sudo rm -rf /opt/mqtt-exporter
sudo rm -rf /etc/mqtt-exporter

# Удаление данных (опционально)
# sudo rm -rf /var/lib/mqtt-exporter
# sudo rm -rf /var/log/mqtt-exporter

# Удаление пользователя
sudo userdel mqtt-exporter

echo "Удаление завершено!"
```