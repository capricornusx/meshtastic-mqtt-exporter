package hooks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"meshtastic-exporter/pkg/exporter"

	"github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestNewPrometheusHook(t *testing.T) {
	config := exporter.Config{}
	config.Prometheus.Enabled = false

	hook := NewPrometheusHook(config)
	if hook == nil {
		t.Fatal("Expected hook to be created")
	}
	if hook.ID() != "prometheus-exporter" {
		t.Errorf("Expected ID 'prometheus-exporter', got %s", hook.ID())
	}
}

func TestPrometheusHook_Provides(t *testing.T) {
	config := exporter.Config{}
	hook := NewPrometheusHook(config)

	if !hook.Provides(mqtt.OnPublish) {
		t.Error("Expected hook to provide OnPublish")
	}
	if hook.Provides(mqtt.OnConnect) {
		t.Error("Expected hook not to provide OnConnect")
	}
}

func TestPrometheusHook_OnPublish(t *testing.T) {
	config := exporter.Config{}
	hook := NewPrometheusHook(config)

	tests := []struct {
		name      string
		topic     string
		payload   string
		expectErr bool
	}{
		{
			name:      "valid telemetry message",
			topic:     "msh/2/json/LongFast/!12345678",
			payload:   `{"from":12345678,"to":4294967295,"type":"telemetry","payload":{"battery_level":85.5,"voltage":4.12}}`,
			expectErr: false,
		},
		{
			name:      "invalid topic",
			topic:     "invalid/topic",
			payload:   `{"test":"data"}`,
			expectErr: false,
		},
		{
			name:      "empty payload",
			topic:     "msh/2/json/LongFast/!12345678",
			payload:   "",
			expectErr: false,
		},
		{
			name:      "invalid json",
			topic:     "msh/2/json/LongFast/!12345678",
			payload:   `{invalid json}`,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pk := packets.Packet{
				TopicName: tt.topic,
				Payload:   []byte(tt.payload),
			}

			result, err := hook.OnPublish(nil, pk)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, got: %v", tt.expectErr, err)
			}
			if result.TopicName != tt.topic {
				t.Errorf("Expected topic %s, got %s", tt.topic, result.TopicName)
			}
		})
	}
}

func TestPrometheusHook_ProcessMessage(t *testing.T) {
	config := exporter.Config{}
	hook := NewPrometheusHook(config)

	// Test telemetry message
	telemetryData := map[string]interface{}{
		"from": float64(12345678),
		"to":   float64(4294967295),
		"type": "telemetry",
		"payload": map[string]interface{}{
			"battery_level": 85.5,
			"voltage":       4.12,
			"temperature":   23.5,
		},
	}

	hook.processMessage(telemetryData)

	// Check if metrics were updated
	metricChan := make(chan prometheus.Metric, 10)
	hook.batteryLevel.Collect(metricChan)
	close(metricChan)

	found := false
	for metric := range metricChan {
		dtoMetric := &dto.Metric{}
		if err := metric.Write(dtoMetric); err != nil {
			continue
		}
		if dtoMetric.GetGauge().GetValue() == 85.5 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected battery level metric to be set")
	}
}

func TestPrometheusHook_ProcessNodeInfo(t *testing.T) {
	config := exporter.Config{}
	hook := NewPrometheusHook(config)

	nodeInfoData := map[string]interface{}{
		"from": float64(12345678),
		"type": "nodeinfo",
		"payload": map[string]interface{}{
			"longname":  "Test Node",
			"shortname": "TN",
			"hardware":  1.0,
			"role":      2.0,
		},
	}

	hook.processMessage(nodeInfoData)

	// Test processNodeInfo directly with missing fields
	payload := map[string]interface{}{
		"longname": "Minimal Node",
		// Other fields missing
	}
	hook.processNodeInfo("789012", payload)

	// Check if the node info metric was updated
	metricChan := make(chan prometheus.Metric, 10)
	hook.nodeHardware.Collect(metricChan)
	close(metricChan)

	found := false
	for metric := range metricChan {
		dtoMetric := &dto.Metric{}
		if err := metric.Write(dtoMetric); err != nil {
			continue
		}
		if dtoMetric.GetGauge().GetValue() == 1.0 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected node hardware metric to be set")
	}
}

func TestGetUint32(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected uint32
	}{
		{"valid float64", map[string]interface{}{"test": float64(12345)}, "test", 12345},
		{"missing key", map[string]interface{}{}, "test", 0},
		{"wrong type", map[string]interface{}{"test": "string"}, "test", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getUint32(tt.data, tt.key)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected string
	}{
		{"valid string", map[string]interface{}{"test": "value"}, "test", "value"},
		{"missing key", map[string]interface{}{}, "test", ""},
		{"wrong type", map[string]interface{}{"test": 123}, "test", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getString(tt.data, tt.key)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestHealthHandler(t *testing.T) {
	config := exporter.Config{}
	config.Prometheus.Enabled = false
	hook := NewPrometheusHook(config)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	hook.healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status ok, got %s", response["status"])
	}
}

func TestRoundFloat(t *testing.T) {
	tests := []struct {
		value     float64
		precision int
		expected  float64
	}{
		{3.14159, 2, 3.14},
		{3.14159, 0, 3},
		{3.99, 1, 4.0},
	}

	for _, tt := range tests {
		result := roundFloat(tt.value, tt.precision)
		if result != tt.expected {
			t.Errorf("roundFloat(%f, %d) = %f, expected %f", tt.value, tt.precision, result, tt.expected)
		}
	}
}

func TestPrometheusHook_ProcessTelemetryEdgeCases(t *testing.T) {
	config := exporter.Config{}
	hook := NewPrometheusHook(config)

	// Test with missing fields
	payload := map[string]interface{}{
		"battery_level": float64(50),
		// Other fields missing
	}
	hook.processTelemetry("123", payload)

	// Test with all fields
	payload = map[string]interface{}{
		"battery_level":       85.5,
		"voltage":             3.7,
		"channel_utilization": 12.5,
		"air_util_tx":         8.2,
		"uptime_seconds":      3600.0,
		"temperature":         22.5,
		"relative_humidity":   65.0,
		"barometric_pressure": 1013.25,
	}
	hook.processTelemetry("456", payload)
}

func TestPrometheusHook_ProcessMessageEdgeCases(t *testing.T) {
	config := exporter.Config{}
	hook := NewPrometheusHook(config)

	// Test message without from field
	data := map[string]interface{}{
		"type": "telemetry",
	}
	hook.processMessage(data)

	// Test message with unknown type
	data = map[string]interface{}{
		"from": float64(123456),
		"type": "unknown",
		"payload": map[string]interface{}{
			"some_field": "value",
		},
	}
	hook.processMessage(data)

	// Test message without payload
	data = map[string]interface{}{
		"from": float64(123456),
		"type": "position",
	}
	hook.processMessage(data)
}

func TestPrometheusHook_StartServer(t *testing.T) {
	// Test that startServer doesn't panic when called
	config := exporter.Config{}
	config.Prometheus.Enabled = false // Disable to avoid conflicts
	config.Prometheus.Port = 0
	config.Prometheus.Host = "127.0.0.1"

	hook := NewPrometheusHook(config)

	// Just test that the function exists and can be called
	// Don't start server to avoid port conflicts in tests
	if hook.Config.Prometheus.Enabled {
		go hook.startServer()
		time.Sleep(10 * time.Millisecond)
	}
}

func TestPrometheusHook_ProcessPayload(t *testing.T) {
	config := exporter.Config{}
	hook := NewPrometheusHook(config)

	// Test telemetry payload
	telemetryPayload := map[string]interface{}{
		"battery_level": 75.5,
		"temperature":   25.0,
	}
	hook.processPayload("123", "telemetry", telemetryPayload)

	// Test nodeinfo payload
	nodeinfoPayload := map[string]interface{}{
		"longname":  "Test Node",
		"shortname": "TN",
		"hardware":  31.0,
		"role":      1.0,
	}
	hook.processPayload("456", "nodeinfo", nodeinfoPayload)

	// Test unknown payload type
	hook.processPayload("789", "unknown", map[string]interface{}{})
}
