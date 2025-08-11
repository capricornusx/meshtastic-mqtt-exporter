package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/mocks"
)

func TestUnifiedServer_Start(t *testing.T) {
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}

	config := UnifiedServerConfig{
		Addr:         ":0", // Random port
		EnableHealth: true,
		AlertPath:    "/alerts",
	}

	server := NewUnifiedServer(config, mockCollector, mockAlerter)
	ctx := context.Background()

	err := server.Start(ctx)
	require.NoError(t, err)

	// Cleanup
	err = server.Shutdown(ctx)
	require.NoError(t, err)
}

func TestUnifiedServer_HealthHandler(t *testing.T) {
	mockCollector := &mocks.MockMetricsCollector{}

	config := UnifiedServerConfig{
		Addr:         ":0",
		EnableHealth: true,
	}

	server := NewUnifiedServer(config, mockCollector, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	server.healthHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "meshtastic-exporter")
	assert.Contains(t, rec.Body.String(), "ok")
}

func TestUnifiedServer_AlertWebhookHandler(t *testing.T) {
	mockAlerter := &mocks.MockAlertSender{}

	config := UnifiedServerConfig{
		Addr:      ":0",
		AlertPath: "/alerts",
	}

	server := NewUnifiedServer(config, nil, mockAlerter)

	payload := AlertPayload{
		Alerts: []AlertItem{
			{
				Status:      "firing",
				Labels:      map[string]string{"alertname": "TestAlert", "severity": "critical"},
				Annotations: map[string]string{"summary": "Test summary"},
			},
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/alerts", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	server.alertWebhookHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())
	assert.True(t, mockAlerter.SendAlertCalled)
}

func TestUnifiedServer_ConvertToAlert(t *testing.T) {
	server := NewUnifiedServer(UnifiedServerConfig{}, nil, nil)

	tests := []struct {
		name     string
		item     AlertItem
		expected string
	}{
		{
			name: "firing_alert",
			item: AlertItem{
				Status:      "firing",
				Labels:      map[string]string{"alertname": "NodeDown", "severity": "critical"},
				Annotations: map[string]string{"summary": "Node is offline"},
			},
			expected: "🚨 firing: NodeDown - Node is offline",
		},
		{
			name: "resolved_alert",
			item: AlertItem{
				Status:      "resolved",
				Labels:      map[string]string{"alertname": "NodeDown", "severity": "warning"},
				Annotations: map[string]string{"summary": "Node is back online"},
			},
			expected: "✅ resolved: NodeDown - Node is back online",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alert := server.convertToAlert(tt.item)
			assert.Equal(t, tt.expected, alert.Message)
			assert.Equal(t, tt.item.Labels["severity"], alert.Severity)
			assert.True(t, time.Since(alert.Timestamp) < time.Second)
		})
	}
}
