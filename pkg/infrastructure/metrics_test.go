package infrastructure

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/domain"
)

func TestPrometheusCollector_CollectTelemetry(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	data := domain.TelemetryData{
		NodeID:       "123456789",
		BatteryLevel: floatPtr(85.5),
		Temperature:  floatPtr(23.4),
		Voltage:      floatPtr(4.1),
		Timestamp:    time.Now(),
	}

	err := collector.CollectTelemetry(data)

	require.NoError(t, err)

	batteryMetric := testutil.ToFloat64(collector.batteryLevel.WithLabelValues("123456789"))
	assert.Equal(t, 85.5, batteryMetric)

	tempMetric := testutil.ToFloat64(collector.temperature.WithLabelValues("123456789"))
	assert.Equal(t, 23.4, tempMetric)

	voltageMetric := testutil.ToFloat64(collector.voltage.WithLabelValues("123456789"))
	assert.Equal(t, 4.1, voltageMetric)
}

func TestPrometheusCollector_CollectNodeInfo(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	info := domain.NodeInfo{
		NodeID:    "987654321",
		LongName:  "Test Node",
		ShortName: "TN01",
		Hardware:  "1",
		Role:      "2",
		Timestamp: time.Now(),
	}

	err := collector.CollectNodeInfo(info)

	require.NoError(t, err)

	nodeInfoMetric := testutil.ToFloat64(collector.nodeHardware.WithLabelValues("987654321", "Test Node", "TN01", "1", "2"))
	assert.Equal(t, 1.0, nodeInfoMetric)
}

func TestPrometheusCollector_GetRegistry(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	registry := collector.GetRegistry()

	assert.NotNil(t, registry)
	assert.IsType(t, &prometheus.Registry{}, registry)
}

func TestPrometheusCollector_CollectTelemetry_PartialData(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	data := domain.TelemetryData{
		NodeID:       "123456789",
		BatteryLevel: floatPtr(75.0),
		Timestamp:    time.Now(),
	}

	err := collector.CollectTelemetry(data)

	require.NoError(t, err)

	batteryMetric := testutil.ToFloat64(collector.batteryLevel.WithLabelValues("123456789"))
	assert.Equal(t, 75.0, batteryMetric)
}

func TestPrometheusCollector_AllTelemetryFields(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	data := domain.TelemetryData{
		NodeID:             "123",
		BatteryLevel:       floatPtr(85.5),
		Voltage:            floatPtr(3.7),
		Temperature:        floatPtr(23.4),
		RelativeHumidity:   floatPtr(65.2),
		BarometricPressure: floatPtr(1013.25),
		ChannelUtilization: floatPtr(12.5),
		AirUtilTx:          floatPtr(8.3),
		UptimeSeconds:      floatPtr(3600.0),
		RSSI:               floatPtr(-95.2),
		SNR:                floatPtr(-8.5),
		Timestamp:          time.Now(),
	}

	err := collector.CollectTelemetry(data)
	require.NoError(t, err)

	registry := collector.GetRegistry()
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	metricNames := make(map[string]bool)
	for _, mf := range metricFamilies {
		metricNames[mf.GetName()] = true
	}

	expectedMetrics := []string{
		domain.MetricBatteryLevel,
		domain.MetricTemperature,
		domain.MetricRSSI,
		domain.MetricSNR,
		domain.MetricVoltage,
		domain.MetricHumidity,
		domain.MetricPressure,
	}

	for _, metric := range expectedMetrics {
		assert.True(t, metricNames[metric], "Metric %s not found", metric)
	}
}

func TestSetBasicMetrics(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	data := domain.TelemetryData{
		NodeID:        "123",
		BatteryLevel:  floatPtr(75.0),
		Voltage:       floatPtr(3.8),
		UptimeSeconds: floatPtr(7200.0),
	}

	collector.setBasicMetrics(data)

	batteryMetric := testutil.ToFloat64(collector.batteryLevel.WithLabelValues("123"))
	assert.Equal(t, 75.0, batteryMetric)

	voltageMetric := testutil.ToFloat64(collector.voltage.WithLabelValues("123"))
	assert.Equal(t, 3.8, voltageMetric)
}

func TestSetEnvironmentalMetrics(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	data := domain.TelemetryData{
		NodeID:             "123",
		Temperature:        floatPtr(25.5),
		RelativeHumidity:   floatPtr(60.0),
		BarometricPressure: floatPtr(1015.0),
	}

	collector.setEnvironmentalMetrics(data)

	tempMetric := testutil.ToFloat64(collector.temperature.WithLabelValues("123"))
	assert.Equal(t, 25.5, tempMetric)

	humidityMetric := testutil.ToFloat64(collector.humidity.WithLabelValues("123"))
	assert.Equal(t, 60.0, humidityMetric)

	pressureMetric := testutil.ToFloat64(collector.pressure.WithLabelValues("123"))
	assert.Equal(t, 1015.0, pressureMetric)
}

func TestSetNetworkMetrics(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	data := domain.TelemetryData{
		NodeID:             "123",
		ChannelUtilization: floatPtr(15.0),
		AirUtilTx:          floatPtr(10.0),
		RSSI:               floatPtr(-90.0),
		SNR:                floatPtr(-5.0),
	}

	collector.setNetworkMetrics(data)

	rssiMetric := testutil.ToFloat64(collector.rssi.WithLabelValues("123"))
	assert.Equal(t, -90.0, rssiMetric)

	snrMetric := testutil.ToFloat64(collector.snr.WithLabelValues("123"))
	assert.Equal(t, -5.0, snrMetric)
}

func floatPtr(f float64) *float64 {
	return &f
}
