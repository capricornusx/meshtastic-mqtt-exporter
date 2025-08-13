package infrastructure

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/domain"
)

func TestNewLoRaAlertSender(t *testing.T) {
	t.Parallel()
	config := LoRaConfig{
		DefaultChannel: "TestChannel",
		DefaultMode:    "direct",
		TargetNodes:    []uint32{123456789, 987654321},
	}

	sender := NewLoRaAlertSender(nil, config)

	require.NotNil(t, sender)
	assert.Equal(t, "TestChannel", sender.defaultChannel)
	assert.Equal(t, "direct", sender.defaultMode)
}

func TestNewLoRaAlertSender_Defaults(t *testing.T) {
	t.Parallel()
	sender := NewLoRaAlertSender(nil, LoRaConfig{})

	assert.Equal(t, "LongFast", sender.defaultChannel)
	assert.Equal(t, "broadcast", sender.defaultMode)
}

func TestSendAlert_Broadcast(t *testing.T) {
	t.Parallel()
	sender := NewLoRaAlertSender(nil, LoRaConfig{})
	alert := domain.Alert{
		Message: "Test alert",
		Mode:    "broadcast",
		Channel: "TestChannel",
	}

	err := sender.SendAlert(context.Background(), alert)

	require.NoError(t, err)
}

func TestSendAlert_Direct(t *testing.T) {
	t.Parallel()
	sender := NewLoRaAlertSender(nil, LoRaConfig{})
	alert := domain.Alert{
		Message:     "Test alert",
		Mode:        "direct",
		Channel:     "TestChannel",
		TargetNodes: []uint32{123456789, 987654321},
	}

	err := sender.SendAlert(context.Background(), alert)

	require.NoError(t, err)
}

func TestSendAlert_UseDefaults(t *testing.T) {
	t.Parallel()
	config := LoRaConfig{
		DefaultChannel: "DefaultChannel",
		DefaultMode:    "broadcast",
		TargetNodes:    []uint32{123456789},
	}
	sender := NewLoRaAlertSender(nil, config)

	alert := domain.Alert{Message: "Test alert"}

	err := sender.SendAlert(context.Background(), alert)

	require.NoError(t, err)
}

func TestPublishMessage(t *testing.T) {
	t.Parallel()
	sender := NewLoRaAlertSender(nil, LoRaConfig{})
	payload := map[string]interface{}{
		"type":    "text",
		"payload": "test message",
	}

	err := sender.publishMessage("test/topic", payload)

	require.NoError(t, err)
}

func TestPublishMessage_InvalidPayload(t *testing.T) {
	t.Parallel()
	sender := NewLoRaAlertSender(nil, LoRaConfig{})
	payload := map[string]interface{}{
		"invalid": make(chan int),
	}

	err := sender.publishMessage("test/topic", payload)

	require.Error(t, err)
}

func TestGetDefaultMQTTDownlinkTopic(t *testing.T) {
	topic := GetDefaultMQTTDownlinkTopic()

	assert.NotEmpty(t, topic)
	assert.Equal(t, "msh/MY_433/2/json/mqtt/", topic)
}

func TestLoRaAlertSender_sendDirect_EdgeCases(t *testing.T) {
	t.Parallel()
	sender := &LoRaAlertSender{
		mqttDownlinkTopic: "test/topic",
		fromNodeID:        123,
	}

	tests := []struct {
		name        string
		message     string
		targetNodes []uint32
	}{
		{
			name:        "empty_nodes",
			message:     "test message",
			targetNodes: []uint32{},
		},
		{
			name:        "valid_nodes",
			message:     "test message",
			targetNodes: []uint32{456, 789},
		},
		{
			name:        "single_node",
			message:     "test message",
			targetNodes: []uint32{456},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := sender.sendDirect(context.Background(), tt.message, tt.targetNodes)
			assert.NoError(t, err)
		})
	}
}

func TestLoRaAlertSender_publishMessage_EdgeCases(t *testing.T) {
	t.Parallel()
	sender := &LoRaAlertSender{
		mqttDownlinkTopic: "test/topic",
	}

	payload := map[string]interface{}{
		"type":    "sendtext",
		"payload": "test message",
		"from":    123,
		"to":      456,
	}

	err := sender.publishMessage("test/topic", payload)
	assert.NoError(t, err)
}
