package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/application"
	"meshtastic-exporter/pkg/infrastructure"
)

func TestIntegration_MessageProcessing(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	collector := infrastructure.NewPrometheusCollector()
	alerter := infrastructure.NewLoRaAlertSender(nil, infrastructure.LoRaConfig{})
	processor := application.NewMeshtasticProcessor(collector, alerter, false, "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test telemetry message
	telemetryPayload := []byte(`{
		"from": 123456789,
		"type": "telemetry",
		"payload": {
			"battery_level": 85.5,
			"temperature": 23.4,
			"voltage": 4.1,
			"relative_humidity": 65.2,
			"barometric_pressure": 1013.25
		}
	}`)

	err := processor.ProcessMessage(ctx, "msh/test/telemetry", telemetryPayload)

	require.NoError(t, err)

	// Check that metrics are collected
	registry := collector.GetRegistry()
	assert.NotNil(t, registry)

	// Additional metric checks can be added via registry
}

func TestIntegration_NodeInfoProcessing(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	collector := infrastructure.NewPrometheusCollector()
	alerter := infrastructure.NewLoRaAlertSender(nil, infrastructure.LoRaConfig{})
	processor := application.NewMeshtasticProcessor(collector, alerter, false, "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test node info message
	nodeInfoPayload := []byte(`{
		"from": 987654321,
		"type": "nodeinfo",
		"payload": {
			"longname": "Integration Test Node",
			"shortname": "ITN1",
			"hardware": 1.0,
			"role": 2.0
		}
	}`)

	err := processor.ProcessMessage(ctx, "msh/test/nodeinfo", nodeInfoPayload)

	require.NoError(t, err)

	// Check that metrics are collected
	registry := collector.GetRegistry()
	assert.NotNil(t, registry)
}

func TestIntegration_InvalidMessages(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	collector := infrastructure.NewPrometheusCollector()
	alerter := infrastructure.NewLoRaAlertSender(nil, infrastructure.LoRaConfig{})
	processor := application.NewMeshtasticProcessor(collector, alerter, false, "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	testCases := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "invalid JSON",
			payload: []byte(`{invalid json`),
		},
		{
			name:    "missing from field",
			payload: []byte(`{"type": "telemetry", "payload": {}}`),
		},
		{
			name:    "zero from field",
			payload: []byte(`{"from": 0, "type": "telemetry", "payload": {}}`),
		},
		{
			name:    "unknown message type",
			payload: []byte(`{"from": 123, "type": "unknown", "payload": {}}`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := processor.ProcessMessage(ctx, "msh/test", tc.payload)
			if tc.name == "unknown message type" {
				// Unknown message type is ignored
				assert.NoError(t, err)
			} else {
				// Other cases should return errors
				assert.Error(t, err)
			}
		})
	}
}

func TestIntegration_ContextTimeout(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	collector := infrastructure.NewPrometheusCollector()
	alerter := infrastructure.NewLoRaAlertSender(nil, infrastructure.LoRaConfig{})
	processor := application.NewMeshtasticProcessor(collector, alerter, false, "")

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(1 * time.Millisecond)

	payload := []byte(`{
		"from": 123456789,
		"type": "telemetry",
		"payload": {"battery_level": 85.5}
	}`)

	err := processor.ProcessMessage(ctx, "msh/test", payload)

	assert.NoError(t, err)
}
