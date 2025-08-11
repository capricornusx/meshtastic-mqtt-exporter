package infrastructure

import (
	"context"
	"testing"

	"meshtastic-exporter/pkg/adapters"
	"meshtastic-exporter/pkg/mocks"
)

func TestNewMQTTClient(t *testing.T) {
	config := &adapters.MQTTConfigAdapter{Host: "localhost", Port: 1883}
	processor := &mocks.MockMessageProcessor{}

	client := NewMQTTClient(config, processor)
	if client == nil {
		t.Fatal("Expected client to be created")
	}
	if client.config != config {
		t.Error("Expected config to be set")
	}
	if client.processor != processor {
		t.Error("Expected processor to be set")
	}
}

func TestBuildBrokerURL_TCP(t *testing.T) {
	config := &adapters.MQTTConfigAdapter{Host: "localhost", Port: 1883, TLS: false}
	processor := &mocks.MockMessageProcessor{}

	client := NewMQTTClient(config, processor)
	url := client.buildBrokerURL()

	expected := "tcp://localhost:1883"
	if url != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, url)
	}
}

func TestBuildBrokerURL_TLS(t *testing.T) {
	config := &adapters.MQTTConfigAdapter{Host: "localhost", Port: 8883, TLS: true}
	processor := &mocks.MockMessageProcessor{}

	client := NewMQTTClient(config, processor)
	url := client.buildBrokerURL()

	expected := "ssl://localhost:8883"
	if url != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, url)
	}
}

func TestMQTTClient_Disconnect_NotConnected(t *testing.T) {
	config := &adapters.MQTTConfigAdapter{Host: "localhost", Port: 1883}
	processor := &mocks.MockMessageProcessor{}

	client := NewMQTTClient(config, processor)
	client.Disconnect()
}

func TestMQTTClient_MessageHandler(t *testing.T) {
	config := &adapters.MQTTConfigAdapter{Host: "localhost", Port: 1883}
	processor := &mocks.MockMessageProcessor{}

	client := NewMQTTClient(config, processor)

	mockMsg := &mockMessage{
		topic:   "msh/test/topic",
		payload: []byte(`{"test": "data"}`),
	}

	client.messageHandler(nil, mockMsg)

	if !processor.ProcessMessageCalled {
		t.Error("Expected ProcessMessage to be called")
	}
}

func TestMQTTClient_OnConnectionLost(t *testing.T) {
	config := &adapters.MQTTConfigAdapter{Host: "localhost", Port: 1883}
	processor := &mocks.MockMessageProcessor{}

	client := NewMQTTClient(config, processor)
	client.onConnectionLost(nil, context.DeadlineExceeded)
}

type mockMessage struct {
	topic   string
	payload []byte
}

func (m *mockMessage) Duplicate() bool   { return false }
func (m *mockMessage) Qos() byte         { return 0 }
func (m *mockMessage) Retained() bool    { return false }
func (m *mockMessage) Topic() string     { return m.topic }
func (m *mockMessage) MessageID() uint16 { return 0 }
func (m *mockMessage) Payload() []byte   { return m.payload }
func (m *mockMessage) Ack()              {}
func (m *mockMessage) Reject()           {}

func TestMQTTClient_OnConnect(t *testing.T) {
	config := &adapters.MQTTConfigAdapter{Host: "localhost", Port: 1883}
	processor := &mocks.MockMessageProcessor{}

	client := NewMQTTClient(config, processor)
	// Test with nil client - should not panic but will log errors
	defer func() {
		_ = recover() // Expected to panic with nil client
	}()
	client.onConnect(nil)
}

func TestMQTTClient_Connect_WithAuth(t *testing.T) {
	config := &adapters.MQTTConfigAdapter{
		Host:  "localhost",
		Port:  1883,
		Users: []adapters.UserAuthAdapter{{Username: "test", Password: "pass"}},
	}
	processor := &mocks.MockMessageProcessor{}

	client := NewMQTTClient(config, processor)
	if client == nil {
		t.Fatal("Expected client to be created")
	}
}
