package domain

import "time"

const (
	DefaultBufferSize = 1024
	MaxMessageSize    = 256

	MetricBatteryLevel = "meshtastic_battery_level_percent"
	MetricVoltage      = "meshtastic_voltage_volts"
	MetricTemperature  = "meshtastic_temperature_celsius"
	MetricHumidity     = "meshtastic_humidity_percent"
	MetricPressure     = "meshtastic_pressure_hpa"
	MetricChannelUtil  = "meshtastic_channel_utilization_percent"
	MetricAirUtilTx    = "meshtastic_air_util_tx_percent"
	MetricUptime       = "meshtastic_uptime_seconds"
	MetricNodeLastSeen = "meshtastic_node_last_seen_timestamp"
	MetricNodeInfo     = "meshtastic_node_info"

	DefaultStateSaveInterval = 5 * time.Minute
	StateFilePermissions     = 0600

	DefaultTimeout       = 30 * time.Second
	DefaultMetricsTTL    = 30 * time.Minute
	DefaultKeepAlive     = 60 * time.Second
	DefaultReadTimeout   = 15 * time.Second
	DefaultWriteTimeout  = 15 * time.Second
	DefaultIdleTimeout   = 60 * time.Second
	DefaultHeaderTimeout = 5 * time.Second

	DefaultTopicPrefix = "msh/"
	DefaultHealthPath  = "/health"
	DefaultMetricsPath = "/metrics"
	DefaultAlertsPath  = "/alerts"

	DefaultPrometheusHost = "localhost"
	DefaultPrometheusPort = 8100
	DefaultAlertPort      = 8080

	DefaultMQTTKeepAlive    = 60 * time.Second
	DefaultMQTTPingTimeout  = 10 * time.Second
	DefaultMQTTConnTimeout  = 30 * time.Second
	DefaultMQTTReconnectInt = 30 * time.Second
	DefaultMQTTDisconnectMs = 250

	MaxTopicLength  = 256
	MaxNodeIDLength = 32

	LoRaBroadcastNodeID = uint32(4294967295)

	ShutdownTimeoutDivider = 3

	MessageTypeTelemetry = "telemetry"
	MessageTypeNodeInfo  = "nodeinfo"
)
