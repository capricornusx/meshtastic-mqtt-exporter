package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/mocks"
)

func TestMeshtasticProcessor_ProcessMessage_Telemetry(t *testing.T) {
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	payload := []byte(`{
		"from": 123456789,
		"type": "telemetry",
		"payload": {
			"battery_level": 85.5,
			"temperature": 23.4,
			"voltage": 4.1
		}
	}`)

	err := processor.ProcessMessage(context.Background(), "msh/test", payload)

	require.NoError(t, err)
	assert.True(t, mockCollector.CollectTelemetryCalled)
}

func TestMeshtasticProcessor_ProcessMessage_NodeInfo(t *testing.T) {
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	payload := []byte(`{
		"from": 987654321,
		"type": "nodeinfo",
		"payload": {
			"longname": "Test Node",
			"shortname": "TN01",
			"hardware": 1.0,
			"role": 2.0
		}
	}`)

	err := processor.ProcessMessage(context.Background(), "msh/test", payload)

	require.NoError(t, err)
	assert.True(t, mockCollector.CollectNodeInfoCalled)
}

func TestMeshtasticProcessor_ProcessMessage_InvalidJSON(t *testing.T) {
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	// Не-JSON сообщения теперь игнорируются
	payload := []byte(`invalid json`)

	err := processor.ProcessMessage(context.Background(), "msh/test", payload)

	require.NoError(t, err)
	assert.False(t, mockCollector.CollectTelemetryCalled)
	assert.False(t, mockCollector.CollectNodeInfoCalled)
}

func TestMeshtasticProcessor_ProcessMessage_ZeroFromNode(t *testing.T) {
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	payload := []byte(`{
		"from": 0,
		"type": "telemetry",
		"payload": {"battery_level": 85.5}
	}`)

	err := processor.ProcessMessage(context.Background(), "msh/test", payload)

	require.Error(t, err)
	assert.False(t, mockCollector.CollectTelemetryCalled)
}

func TestMeshtasticProcessor_LogAllMessages(t *testing.T) {
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, true, "msh/#")

	payload := []byte(`{"from": 123456789, "type": "telemetry"}`)

	err := processor.ProcessMessage(context.Background(), "msh/test", payload)

	require.NoError(t, err)
}
