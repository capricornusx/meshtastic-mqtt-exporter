package application

import (
	"context"
	"testing"

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

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !mockCollector.CollectTelemetryCalled {
		t.Error("Expected CollectTelemetry to be called")
	}
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

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !mockCollector.CollectNodeInfoCalled {
		t.Error("Expected CollectNodeInfo to be called")
	}
}

func TestMeshtasticProcessor_ProcessMessage_InvalidJSON(t *testing.T) {
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	// Не-JSON сообщения теперь игнорируются
	payload := []byte(`invalid json`)

	err := processor.ProcessMessage(context.Background(), "msh/test", payload)

	if err != nil {
		t.Errorf("Expected no error for non-JSON message, got %v", err)
	}
	if mockCollector.CollectTelemetryCalled {
		t.Error("Expected CollectTelemetry not to be called")
	}
	if mockCollector.CollectNodeInfoCalled {
		t.Error("Expected CollectNodeInfo not to be called")
	}
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

	if err == nil {
		t.Error("Expected error for zero from node")
	}
	if mockCollector.CollectTelemetryCalled {
		t.Error("Expected CollectTelemetry not to be called")
	}
}

func TestMeshtasticProcessor_LogAllMessages(t *testing.T) {
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, true, "msh/#")

	payload := []byte(`{"from": 123456789, "type": "telemetry"}`)

	err := processor.ProcessMessage(context.Background(), "msh/test", payload)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
