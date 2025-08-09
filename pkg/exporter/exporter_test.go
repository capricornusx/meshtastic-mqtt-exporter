package exporter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected Config
	}{
		{
			name: "minimal config",
			content: `mqtt:
  host: test.local
  port: 1884
prometheus:
  enabled: false`,
			expected: Config{
				MQTT: struct {
					Host           string     `yaml:"host"`
					Port           int        `yaml:"port"`
					Username       string     `yaml:"username"`
					Password       string     `yaml:"password"`
					TLS            bool       `yaml:"tls"`
					AllowAnonymous bool       `yaml:"allow_anonymous"`
					Users          []MQTTUser `yaml:"users,omitempty"`
					Broker         struct {
						MaxInflight     int  `yaml:"max_inflight"`
						MaxQueued       int  `yaml:"max_queued"`
						RetainAvailable bool `yaml:"retain_available"`
						MaxPacketSize   int  `yaml:"max_packet_size"`
						KeepAlive       int  `yaml:"keep_alive"`
					} `yaml:"broker,omitempty"`
				}{Host: "test.local", Port: 1884, Broker: struct {
					MaxInflight     int  `yaml:"max_inflight"`
					MaxQueued       int  `yaml:"max_queued"`
					RetainAvailable bool `yaml:"retain_available"`
					MaxPacketSize   int  `yaml:"max_packet_size"`
					KeepAlive       int  `yaml:"keep_alive"`
				}{MaxInflight: 50, MaxQueued: 1000, RetainAvailable: true, MaxPacketSize: 131072, KeepAlive: 60}},
				Prometheus: struct {
					Enabled bool   `yaml:"enabled"`
					Port    int    `yaml:"port"`
					Host    string `yaml:"host"`
				}{Enabled: false, Port: 8000, Host: "127.0.0.1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "config-*.yaml")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatal(err)
			}
			tmpFile.Close()

			config, err := LoadConfig(tmpFile.Name())
			if err != nil {
				t.Fatalf("LoadConfig() error = %v", err)
			}

			if config.MQTT.Host != tt.expected.MQTT.Host {
				t.Errorf("Host = %v, want %v", config.MQTT.Host, tt.expected.MQTT.Host)
			}
			if config.MQTT.Port != tt.expected.MQTT.Port {
				t.Errorf("Port = %v, want %v", config.MQTT.Port, tt.expected.MQTT.Port)
			}
		})
	}
}

func TestProcessMessage(t *testing.T) {
	config := Config{}
	config.Prometheus.Enabled = false
	exporter := New(config)
	exporter.Init()

	tests := []struct {
		name string
		data map[string]interface{}
	}{
		{
			name: "telemetry message",
			data: map[string]interface{}{
				"from": float64(123456),
				"to":   float64(789012),
				"type": "telemetry",
				"payload": map[string]interface{}{
					"battery_level": float64(85.5),
					"voltage":       float64(3.7),
					"temperature":   float64(22.5),
				},
			},
		},
		{
			name: "nodeinfo message",
			data: map[string]interface{}{
				"from": float64(123456),
				"type": "nodeinfo",
				"payload": map[string]interface{}{
					"longname":  "Test Node",
					"shortname": "TN",
					"hardware":  float64(31),
					"role":      float64(1),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			exporter.processMessage(tt.data)
		})
	}
}

func TestStateManagement(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "state-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	config := Config{}
	config.State.Enabled = true
	config.State.File = tmpFile.Name()
	config.Prometheus.Enabled = false

	exporter := New(config)
	exporter.Init()

	state := State{
		Nodes: map[string]NodeState{
			"123456": {
				BatteryLevel: 85.5,
				Temperature:  22.5,
			},
		},
		Timestamp: time.Now(),
	}

	data, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(tmpFile.Name(), data, 0600)

	exporter.loadState()
}

func TestUtilityFunctions(t *testing.T) {
	data := map[string]interface{}{
		"uint_val":   float64(123),
		"string_val": "test",
		"wrong_type": 123,
	}

	if getUint32(data, "uint_val") != 123 {
		t.Error("getUint32 failed")
	}
	if getUint32(data, "missing") != 0 {
		t.Error("getUint32 should return 0 for missing key")
	}
	if getString(data, "string_val") != "test" {
		t.Error("getString failed")
	}
	if getString(data, "wrong_type") != "" {
		t.Error("getString should return empty for wrong type")
	}

	if roundFloat(3.14159, 2) != 3.14 {
		t.Error("roundFloat failed")
	}
}

func TestConfigDefaults(t *testing.T) {
	config, err := LoadConfig("non-existent-file.yaml")
	if err != nil {
		t.Fatalf("LoadConfig() should not error on missing file: %v", err)
	}
	if config.MQTT.Host != "localhost" {
		t.Errorf("Default host = %v, want localhost", config.MQTT.Host)
	}
	if config.MQTT.Port != 1883 {
		t.Errorf("Default port = %v, want 1883", config.MQTT.Port)
	}
}

func TestHealthHandler(t *testing.T) {
	config := Config{}
	exporter := New(config)
	exporter.Init()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	exporter.healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %s", response["status"])
	}
}

func TestProcessTelemetryEdgeCases(t *testing.T) {
	config := Config{}
	exporter := New(config)
	exporter.Init()

	// Test with missing fields
	payload := map[string]interface{}{
		"battery_level": float64(50),
		// Missing other fields
	}
	exporter.processTelemetry("123", payload)

	// Test with all fields
	payload = map[string]interface{}{
		"battery_level":       float64(85.5),
		"voltage":             float64(3.7),
		"channel_utilization": float64(12.5),
		"air_util_tx":         float64(8.2),
		"uptime_seconds":      float64(3600),
		"temperature":         float64(22.5),
		"relative_humidity":   float64(65.0),
		"barometric_pressure": float64(1013.25),
	}
	exporter.processTelemetry("123", payload)
}

func TestProcessMessageEdgeCases(t *testing.T) {
	config := Config{}
	exporter := New(config)
	exporter.Init()

	// Test message without from field
	data := map[string]interface{}{
		"type": "telemetry",
	}
	exporter.processMessage(data)

	// Test message with RSSI and SNR
	data = map[string]interface{}{
		"from": float64(123456),
		"to":   float64(789012),
		"type": "telemetry",
		"rssi": float64(-85.5),
		"snr":  float64(12.3),
		"payload": map[string]interface{}{
			"battery_level": float64(75),
		},
	}
	exporter.processMessage(data)

	// Test message without payload
	data = map[string]interface{}{
		"from": float64(123456),
		"type": "position",
	}
	exporter.processMessage(data)
}

func TestStateFileOperations(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "state-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	config := Config{}
	config.State.Enabled = true
	config.State.File = tmpFile.Name()

	exporter := New(config)
	exporter.Init()

	// Set some metric values
	exporter.batteryLevel.WithLabelValues("123").Set(85.5)
	exporter.voltage.WithLabelValues("123").Set(3.7)
	exporter.temperature.WithLabelValues("123").Set(22.5)
	exporter.uptime.WithLabelValues("123").Set(3600)

	// Test save state
	exporter.saveState()

	// Verify file exists and has content
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("State file should not be empty")
	}

	// Test load state with corrupted file
	os.WriteFile(tmpFile.Name(), []byte("invalid json"), 0600)
	exporter.loadState() // Should not panic
}

func TestMQTTConnectionHandlers(t *testing.T) {
	config := Config{}
	exporter := New(config)
	exporter.Init()

	// Test connection lost handler
	exporter.onConnectionLost(nil, nil)

	// Test message handler with invalid JSON
	msg := &mockMessage{
		payload: []byte("invalid json"),
	}
	exporter.messageHandler(nil, msg)

	// Test message handler with valid JSON
	msg = &mockMessage{
		payload: []byte(`{"from":123456,"type":"telemetry","payload":{"battery_level":85.5}}`),
	}
	exporter.messageHandler(nil, msg)
}

type mockMessage struct {
	payload []byte
}

func (m *mockMessage) Duplicate() bool   { return false }
func (m *mockMessage) Qos() byte         { return 0 }
func (m *mockMessage) Retained() bool    { return false }
func (m *mockMessage) Topic() string     { return "test/topic" }
func (m *mockMessage) MessageID() uint16 { return 1 }
func (m *mockMessage) Payload() []byte   { return m.payload }
func (m *mockMessage) Ack()              {}

func TestExtractMetricValues(t *testing.T) {
	config := Config{}
	exporter := New(config)
	exporter.Init()

	// Set some values
	exporter.batteryLevel.WithLabelValues("123").Set(85.5)
	exporter.voltage.WithLabelValues("123").Set(3.7)
	exporter.temperature.WithLabelValues("456").Set(22.5)
	exporter.uptime.WithLabelValues("456").Set(3600)

	nodes := exporter.extractMetricValues()
	if len(nodes) == 0 {
		t.Error("Expected extracted nodes")
	}
}
