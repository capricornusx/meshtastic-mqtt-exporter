package infrastructure

import (
	"context"
	"testing"

	"meshtastic-exporter/pkg/domain"
)

func TestNewLoRaAlertSender(t *testing.T) {
	config := LoRaConfig{
		DefaultChannel: "TestChannel",
		DefaultMode:    "direct",
		TargetNodes:    []string{"node1", "node2"},
	}

	sender := NewLoRaAlertSender(nil, config)
	if sender == nil {
		t.Fatal("Expected sender to be created")
	}
	if sender.defaultChannel != "TestChannel" {
		t.Errorf("Expected channel 'TestChannel', got '%s'", sender.defaultChannel)
	}
	if sender.defaultMode != "direct" {
		t.Errorf("Expected mode 'direct', got '%s'", sender.defaultMode)
	}
}

func TestNewLoRaAlertSender_Defaults(t *testing.T) {
	sender := NewLoRaAlertSender(nil, LoRaConfig{})
	if sender.defaultChannel != "LongFast" {
		t.Errorf("Expected default channel 'LongFast', got '%s'", sender.defaultChannel)
	}
	if sender.defaultMode != "broadcast" {
		t.Errorf("Expected default mode 'broadcast', got '%s'", sender.defaultMode)
	}
}

func TestSendAlert_Broadcast(t *testing.T) {
	sender := NewLoRaAlertSender(nil, LoRaConfig{})
	alert := domain.Alert{
		Message: "Test alert",
		Mode:    "broadcast",
		Channel: "TestChannel",
	}

	err := sender.SendAlert(context.Background(), alert)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestSendAlert_Direct(t *testing.T) {
	sender := NewLoRaAlertSender(nil, LoRaConfig{})
	alert := domain.Alert{
		Message:     "Test alert",
		Mode:        "direct",
		Channel:     "TestChannel",
		TargetNodes: []string{"node1", "node2"},
	}

	err := sender.SendAlert(context.Background(), alert)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestSendAlert_UseDefaults(t *testing.T) {
	config := LoRaConfig{
		DefaultChannel: "DefaultChannel",
		DefaultMode:    "broadcast",
		TargetNodes:    []string{"default1"},
	}
	sender := NewLoRaAlertSender(nil, config)

	alert := domain.Alert{Message: "Test alert"}

	err := sender.SendAlert(context.Background(), alert)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestPublishMessage(t *testing.T) {
	sender := NewLoRaAlertSender(nil, LoRaConfig{})
	payload := map[string]interface{}{
		"type":    "text",
		"payload": "test message",
	}

	err := sender.publishMessage("test/topic", payload)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestPublishMessage_InvalidPayload(t *testing.T) {
	sender := NewLoRaAlertSender(nil, LoRaConfig{})
	payload := map[string]interface{}{
		"invalid": make(chan int),
	}

	err := sender.publishMessage("test/topic", payload)
	if err == nil {
		t.Error("Expected error for invalid payload")
	}
}
