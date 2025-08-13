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
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}

	config := UnifiedServerConfig{
		Addr:         ":0",
		EnableHealth: true,
		AlertPath:    "/alerts",
	}

	server := NewUnifiedServer(config, mockCollector, mockAlerter)
	ctx := context.Background()

	err := server.Start(ctx)
	require.NoError(t, err)

	err = server.Shutdown(ctx)
	require.NoError(t, err)
}

func TestUnifiedServer_HealthHandler(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

func TestUnifiedServer_AlertWebhookHandler_InvalidJSON(t *testing.T) {
	t.Parallel()
	collector := &mocks.MockMetricsCollector{}
	alerter := &mocks.MockAlertSender{}

	config := UnifiedServerConfig{
		Addr:         "localhost:0",
		EnableHealth: true,
		AlertPath:    "/alerts/webhook",
	}

	server := NewUnifiedServer(config, collector, alerter)

	req := httptest.NewRequest(http.MethodPost, "/alerts/webhook", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.alertWebhookHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUnifiedServer_AlertWebhookHandler_EmptyAlerts(t *testing.T) {
	t.Parallel()
	collector := &mocks.MockMetricsCollector{}
	alerter := &mocks.MockAlertSender{}

	config := UnifiedServerConfig{
		Addr:         "localhost:0",
		EnableHealth: true,
		AlertPath:    "/alerts/webhook",
	}

	server := NewUnifiedServer(config, collector, alerter)

	alertPayload := map[string]interface{}{
		"alerts": []interface{}{},
	}

	jsonData, _ := json.Marshal(alertPayload)
	req := httptest.NewRequest(http.MethodPost, "/alerts/webhook", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.alertWebhookHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUnifiedServer_ConvertToAlert(t *testing.T) {
	t.Parallel()
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
		{
			name: "missing_fields",
			item: AlertItem{
				Status:      "firing",
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
			},
			expected: "🚨 firing: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			alert := server.convertToAlert(tt.item)
			assert.Equal(t, tt.expected, alert.Message)
			assert.Equal(t, tt.item.Labels["severity"], alert.Severity)
			assert.True(t, time.Since(alert.Timestamp) < time.Second)
		})
	}
}

func TestUnifiedServer_Start_InvalidAddress(t *testing.T) {
	t.Parallel()
	collector := &mocks.MockMetricsCollector{}
	alerter := &mocks.MockAlertSender{}

	config := UnifiedServerConfig{
		Addr:         "invalid:address:format",
		EnableHealth: true,
		AlertPath:    "/alerts/webhook",
	}

	server := NewUnifiedServer(config, collector, alerter)

	ctx := context.Background()
	err := server.Start(ctx)

	assert.NoError(t, err)
}
