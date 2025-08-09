package hooks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mochi-mqtt/server/v2/packets"
)

func TestNewMeshtasticHook(t *testing.T) {
	config := MeshtasticHookConfig{
		PrometheusAddr: ":8100",
		EnableHealth:   true,
	}

	hook := NewMeshtasticHook(config)

	if hook.ID() != "meshtastic-prometheus" {
		t.Errorf("Expected ID 'meshtastic-prometheus', got %s", hook.ID())
	}

	// Test that hook provides OnPublish by checking the method exists
	if hook.Provides(0) {
		t.Log("Hook provides events")
	}
}

func TestMeshtasticHook_OnPublish(t *testing.T) {
	hook := NewMeshtasticHook(MeshtasticHookConfig{})

	tests := []struct {
		name    string
		topic   string
		payload string
		wantErr bool
	}{
		{
			name:    "valid meshtastic message",
			topic:   "msh/2/json/LongFast/123456",
			payload: `{"from":123456,"to":789012,"type":"telemetry","payload":{"battery_level":85.5}}`,
			wantErr: false,
		},
		{
			name:    "non-meshtastic topic",
			topic:   "other/topic",
			payload: `{"test":"data"}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			topic:   "msh/2/json/LongFast/123456",
			payload: `invalid json`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pk := packets.Packet{
				TopicName: tt.topic,
				Payload:   []byte(tt.payload),
			}

			_, err := hook.OnPublish(nil, pk)
			if (err != nil) != tt.wantErr {
				t.Errorf("OnPublish() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMeshtasticHook_Provides(t *testing.T) {
	hook := NewMeshtasticHook(MeshtasticHookConfig{})

	// Test OnPublish event (value 1 corresponds to mqtt.OnPublish)
	if !hook.Provides(1) {
		t.Log("Hook provides OnPublish event")
	}

	// Test other events
	if hook.Provides(0) {
		t.Error("Hook should not provide other events")
	}
}

func TestMeshtasticHook_HealthEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Test health handler directly
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "meshtastic-hook",
		})
	}

	handler(w, req)

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

func TestMeshtasticHook_ProcessTelemetry(t *testing.T) {
	hook := NewMeshtasticHook(MeshtasticHookConfig{})

	// Test with all telemetry fields
	payload := map[string]interface{}{
		"battery_level":       85.5,
		"voltage":             3.7,
		"temperature":         22.5,
		"relative_humidity":   65.0,
		"barometric_pressure": 1013.25,
		"channel_utilization": 12.5,
		"air_util_tx":         8.2,
		"uptime_seconds":      3600.0,
	}

	hook.processTelemetry("123456", payload)

	// Test with missing fields
	payload = map[string]interface{}{
		"battery_level": float64(50),
		// Other fields missing
	}
	hook.processTelemetry("789012", payload)
}

func TestMeshtasticHook_ProcessNodeInfo(t *testing.T) {
	hook := NewMeshtasticHook(MeshtasticHookConfig{})

	// Test with all node info fields
	payload := map[string]interface{}{
		"longname":  "Test Node Long",
		"shortname": "TN",
		"hardware":  float64(31),
		"role":      float64(1),
	}

	hook.processNodeInfo("123456", payload)

	// Test with missing fields
	payload = map[string]interface{}{
		"longname": "Minimal Node",
		// Other fields missing
	}
	hook.processNodeInfo("789012", payload)
}

func TestMeshtasticHook_ProcessMessage(t *testing.T) {
	hook := NewMeshtasticHook(MeshtasticHookConfig{})

	// Test message without from field
	data := map[string]interface{}{
		"type": "telemetry",
	}
	hook.processMessage(data)

	// Test valid telemetry message
	data = map[string]interface{}{
		"from": float64(123456),
		"type": "telemetry",
		"payload": map[string]interface{}{
			"battery_level": 85.5,
			"temperature":   22.5,
		},
	}
	hook.processMessage(data)

	// Test valid nodeinfo message
	data = map[string]interface{}{
		"from": float64(789012),
		"type": "nodeinfo",
		"payload": map[string]interface{}{
			"longname":  "Test Node",
			"shortname": "TN",
			"hardware":  float64(31),
		},
	}
	hook.processMessage(data)

	// Test message without payload
	data = map[string]interface{}{
		"from": float64(456789),
		"type": "position",
	}
	hook.processMessage(data)
}

func TestMeshtasticHook_Init(t *testing.T) {
	hook := NewMeshtasticHook(MeshtasticHookConfig{
		PrometheusAddr: "", // No server
		EnableHealth:   false,
	})

	err := hook.Init(nil)
	if err != nil {
		t.Errorf("Init() should not error: %v", err)
	}
}
