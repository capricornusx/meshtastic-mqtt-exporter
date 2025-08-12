package infrastructure

import (
	"os"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/domain"
)

func TestPrometheusCollector_SaveAndLoadState(t *testing.T) {
	collector := NewPrometheusCollector()
	tempFile := "test_state.json"
	defer os.Remove(tempFile)

	populateTestData(t, collector)
	saveAndVerifyState(t, collector, tempFile)
	verifyStateRestore(t, tempFile)
}

func populateTestData(t *testing.T, collector *PrometheusCollector) {
	telemetryData := domain.TelemetryData{
		NodeID:       "test-node-123",
		BatteryLevel: floatPtr(85.5),
		Temperature:  floatPtr(23.4),
		Voltage:      floatPtr(4.1),
		Timestamp:    time.Now(),
	}

	nodeInfo := domain.NodeInfo{
		NodeID:    "test-node-123",
		LongName:  "Test Node",
		ShortName: "TN",
		Hardware:  "TBEAM",
		Role:      "CLIENT",
		Timestamp: time.Now(),
	}

	require.NoError(t, collector.CollectTelemetry(telemetryData))
	require.NoError(t, collector.CollectNodeInfo(nodeInfo))
}

func saveAndVerifyState(t *testing.T, collector *PrometheusCollector, tempFile string) {
	require.NoError(t, collector.SaveState(tempFile))
	_, err := os.Stat(tempFile)
	require.NoError(t, err, "State file was not created")
}

func verifyStateRestore(t *testing.T, tempFile string) {
	newCollector := NewPrometheusCollector()
	require.NoError(t, newCollector.LoadState(tempFile))

	metricFamilies, err := newCollector.GetRegistry().Gather()
	require.NoError(t, err)

	verifyRestoredMetrics(t, metricFamilies)
}

func verifyRestoredMetrics(t *testing.T, metricFamilies []*dto.MetricFamily) {
	foundBattery := false
	foundTemperature := false
	foundNodeInfo := false

	for _, mf := range metricFamilies {
		switch mf.GetName() {
		case "meshtastic_battery_level_percent":
			foundBattery = true
			if len(mf.GetMetric()) > 0 {
				assert.Equal(t, 85.5, mf.GetMetric()[0].GetGauge().GetValue())
			}
		case "meshtastic_temperature_celsius":
			foundTemperature = true
			if len(mf.GetMetric()) > 0 {
				assert.Equal(t, 23.4, mf.GetMetric()[0].GetGauge().GetValue())
			}
		case "meshtastic_node_info":
			foundNodeInfo = true
		}
	}

	assert.True(t, foundBattery, "Battery metric not found after state restore")
	assert.True(t, foundTemperature, "Temperature metric not found after state restore")
	assert.True(t, foundNodeInfo, "Node info metric not found after state restore")
}

func TestPrometheusCollector_LoadState_FileNotExists(t *testing.T) {
	collector := NewPrometheusCollector()
	err := collector.LoadState("non_existent_file.json")
	if err != nil {
		t.Errorf("LoadState should not fail when file doesn't exist, got: %v", err)
	}
}

func TestPrometheusCollector_SaveState_EmptyFilename(t *testing.T) {
	collector := NewPrometheusCollector()
	err := collector.SaveState("")
	if err != nil {
		t.Errorf("SaveState should not fail with empty filename, got: %v", err)
	}
}
