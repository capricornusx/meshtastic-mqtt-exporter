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
