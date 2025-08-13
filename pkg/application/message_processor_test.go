package application

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/mocks"
)

func TestMeshtasticProcessor_ProcessMessage_Telemetry(t *testing.T) {
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	payload := []byte(fmt.Sprintf(`{
		"from": 123456789,
		"type": "%s",
		"payload": {
			"battery_level": 85.5,
			"temperature": 23.4,
			"voltage": 4.1
		}
	}`, domain.MessageTypeTelemetry))

	err := processor.ProcessMessage(context.Background(), "msh/test", payload)

	require.NoError(t, err)
	assert.True(t, mockCollector.CollectTelemetryCalled)
}

func TestMeshtasticProcessor_ProcessMessage_NodeInfo(t *testing.T) {
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	payload := []byte(fmt.Sprintf(`{
		"from": 987654321,
		"type": "%s",
		"payload": {
			"longname": "Test Node",
			"shortname": "TN01",
			"hardware": 1.0,
			"role": 2.0
		}
	}`, domain.MessageTypeNodeInfo))

	err := processor.ProcessMessage(context.Background(), "msh/test", payload)

	require.NoError(t, err)
	assert.True(t, mockCollector.CollectNodeInfoCalled)
}

func TestMeshtasticProcessor_ProcessMessage_InvalidJSON(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	payload := []byte(fmt.Sprintf(`{
		"from": 0,
		"type": "%s",
		"payload": {"battery_level": 85.5}
	}`, domain.MessageTypeTelemetry))

	err := processor.ProcessMessage(context.Background(), "msh/test", payload)

	require.Error(t, err)
	assert.False(t, mockCollector.CollectTelemetryCalled)
}

func TestMeshtasticProcessor_LogAllMessages(t *testing.T) {
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, true, "msh/#")

	payload := []byte(fmt.Sprintf(`{"from": 123456789, "type": "%s"}`, domain.MessageTypeTelemetry))

	err := processor.ProcessMessage(context.Background(), "msh/test", payload)

	require.NoError(t, err)
}
