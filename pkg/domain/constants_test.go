package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDefaultMQTTTopics(t *testing.T) {
	topics := GetDefaultMQTTTopics()

	assert.NotEmpty(t, topics)
	assert.Contains(t, topics, "msh/+/+/json/+/+")
	assert.Contains(t, topics, "msh/2/json/+/+")

	// Проверяем что возвращается копия, а не оригинальный слайс
	originalLen := len(topics)
	_ = append(topics, "test/topic")

	newTopics := GetDefaultMQTTTopics()
	assert.Equal(t, originalLen, len(newTopics))
	assert.NotContains(t, newTopics, "test/topic")
}

func TestConstants(t *testing.T) {
	// Проверяем что все константы определены
	assert.NotEmpty(t, DefaultTopicPrefix)
	assert.NotEmpty(t, DefaultPrometheusHost)
	assert.NotZero(t, DefaultPrometheusPort)
	assert.NotEmpty(t, DefaultMetricsPath)
	assert.NotEmpty(t, DefaultAlertsPath)
	assert.NotZero(t, DefaultTimeout)
	assert.NotZero(t, DefaultMetricsTTL)
	assert.NotZero(t, DefaultStateSaveInterval)
	assert.NotZero(t, StateFilePermissions)
	assert.NotZero(t, ShutdownTimeoutDivider)
}

func TestMessageTypes(t *testing.T) {
	// Проверяем что все типы сообщений определены
	assert.NotEmpty(t, MessageTypeTelemetry)
	assert.NotEmpty(t, MessageTypeNodeInfo)
	assert.NotEmpty(t, MessageTypeText)
	assert.NotEmpty(t, MessageTypePosition)
	assert.NotEmpty(t, MessageTypeWaypoint)
	assert.NotEmpty(t, MessageTypeNeighborInfo)
}

func TestMetricNames(t *testing.T) {
	// Проверяем что все имена метрик определены
	assert.NotEmpty(t, MetricMessagesTotal)
	assert.NotEmpty(t, MetricBatteryLevel)
	assert.NotEmpty(t, MetricVoltage)
	assert.NotEmpty(t, MetricTemperature)
	assert.NotEmpty(t, MetricHumidity)
	assert.NotEmpty(t, MetricPressure)
	assert.NotEmpty(t, MetricChannelUtil)
	assert.NotEmpty(t, MetricAirUtilTx)
	assert.NotEmpty(t, MetricUptime)
	assert.NotEmpty(t, MetricRSSI)
	assert.NotEmpty(t, MetricSNR)
	assert.NotEmpty(t, MetricNodeLastSeen)
	assert.NotEmpty(t, MetricNodeInfo)
	assert.NotEmpty(t, MetricExporterInfo)

	// Проверяем что метрики имеют правильный префикс
	assert.Contains(t, MetricMessagesTotal, "meshtastic_")
	assert.Contains(t, MetricBatteryLevel, "meshtastic_")
	assert.Contains(t, MetricTemperature, "meshtastic_")
}

func TestTelemetryTypes(t *testing.T) {
	// Проверяем что все типы телеметрии определены
	assert.NotEmpty(t, TelemetryTypeDevice)
	assert.NotEmpty(t, TelemetryTypeEnvironment)
	assert.NotEmpty(t, TelemetryTypePower)
}

func TestDefaultValues(t *testing.T) {
	// Проверяем разумность значений по умолчанию
	assert.Equal(t, "msh/", DefaultTopicPrefix)
	assert.Equal(t, "localhost", DefaultPrometheusHost)
	assert.Equal(t, 8100, DefaultPrometheusPort)
	assert.Equal(t, "/metrics", DefaultMetricsPath)
	assert.Equal(t, "/alerts", DefaultAlertsPath)

	// Проверяем временные интервалы
	assert.True(t, DefaultTimeout > 0)
	assert.True(t, DefaultMetricsTTL > DefaultTimeout)
	assert.True(t, DefaultStateSaveInterval > 0)
	assert.True(t, DefaultStateSaveInterval < DefaultMetricsTTL)
}
