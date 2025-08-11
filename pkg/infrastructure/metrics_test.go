package infrastructure

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"

	"meshtastic-exporter/pkg/domain"
)

func TestPrometheusCollector_CollectTelemetry(t *testing.T) {
	collector := NewPrometheusCollector()

	data := domain.TelemetryData{
		NodeID:       "123456789",
		BatteryLevel: floatPtr(85.5),
		Temperature:  floatPtr(23.4),
		Voltage:      floatPtr(4.1),
		Timestamp:    time.Now(),
	}

	err := collector.CollectTelemetry(data)

	assert.NoError(t, err)

	// Check that metrics are set
	batteryMetric := testutil.ToFloat64(collector.batteryLevel.WithLabelValues("123456789"))
	assert.Equal(t, 85.5, batteryMetric)

	tempMetric := testutil.ToFloat64(collector.temperature.WithLabelValues("123456789"))
	assert.Equal(t, 23.4, tempMetric)

	voltageMetric := testutil.ToFloat64(collector.voltage.WithLabelValues("123456789"))
	assert.Equal(t, 4.1, voltageMetric)
}

func TestPrometheusCollector_CollectNodeInfo(t *testing.T) {
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

	assert.NoError(t, err)

	// Check that node_info metric is set
	nodeInfoMetric := testutil.ToFloat64(collector.nodeHardware.WithLabelValues("987654321", "Test Node", "TN01", "1", "2"))
	assert.Equal(t, 1.0, nodeInfoMetric)
}

func TestPrometheusCollector_GetRegistry(t *testing.T) {
	collector := NewPrometheusCollector()

	registry := collector.GetRegistry()

	assert.NotNil(t, registry)
	assert.IsType(t, &prometheus.Registry{}, registry)
}

func TestPrometheusCollector_CollectTelemetry_PartialData(t *testing.T) {
	collector := NewPrometheusCollector()

	// Data with only battery_level
	data := domain.TelemetryData{
		NodeID:       "123456789",
		BatteryLevel: floatPtr(75.0),
		// Other fields are nil
		Timestamp: time.Now(),
	}

	err := collector.CollectTelemetry(data)

	assert.NoError(t, err)

	// Check that only battery_level is set
	batteryMetric := testutil.ToFloat64(collector.batteryLevel.WithLabelValues("123456789"))
	assert.Equal(t, 75.0, batteryMetric)
}

func floatPtr(f float64) *float64 {
	return &f
}
