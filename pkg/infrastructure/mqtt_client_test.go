package infrastructure

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

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
	config := &adapters.MQTTConfigAdapter{
		Host:      "localhost",
		Port:      1883,
		TLSConfig: adapters.TLSConfigAdapter{Enabled: false},
	}
	processor := &mocks.MockMessageProcessor{}

	client := NewMQTTClient(config, processor)
	url := client.buildBrokerURL()

	expected := "tcp://localhost:1883"
	if url != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, url)
	}
}

func TestBuildBrokerURL_TLS(t *testing.T) {
	config := &adapters.MQTTConfigAdapter{
		Host:      "localhost",
		Port:      1883,
		TLSConfig: adapters.TLSConfigAdapter{Enabled: true, Port: 8883},
	}
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

func TestMQTTClient_Connect_WithTLS(t *testing.T) {
	config := &adapters.MQTTConfigAdapter{
		Host: "localhost",
		Port: 1883,
		TLSConfig: adapters.TLSConfigAdapter{
			Enabled:            true,
			Port:               8883,
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
		},
	}
	processor := &mocks.MockMessageProcessor{}

	client := NewMQTTClient(config, processor)
	// Тест создания клиента с TLS конфигурацией
	if client == nil {
		t.Fatal("Expected client to be created")
	}
}

func TestMQTTClient_Connect_FailedConnection(t *testing.T) {
	// Тест попытки подключения к несуществующему брокеру
	config := &adapters.MQTTConfigAdapter{
		Host:     "nonexistent.host",
		Port:     1883,
		ClientID: "test-client",
		Topics:   []string{"test/topic"},
	}
	processor := &mocks.MockMessageProcessor{}

	client := NewMQTTClient(config, processor)

	// Попытка подключения должна вернуть ошибку
	err := client.Connect()
	if err == nil {
		t.Error("Expected connection to fail for nonexistent host")
	}
}

func TestMQTTClient_MessageHandler_WithError(t *testing.T) {
	config := &adapters.MQTTConfigAdapter{Host: "localhost", Port: 1883}
	processor := &mocks.MockMessageProcessorWithErrors{}
	processor.SetError(true, "processing failed")

	client := NewMQTTClient(config, processor)

	mockMsg := &mockMessage{
		topic:   "msh/test/topic",
		payload: []byte(`{"test": "data"}`),
	}

	// Тест обработки сообщения с ошибкой - просто проверяем, что не паникуем
	client.messageHandler(nil, mockMsg)
}

func TestMQTTClient_MessageHandler_WithTimeout(t *testing.T) {
	config := &adapters.MQTTConfigAdapter{Host: "localhost", Port: 1883}
	processor := &mocks.MockMessageProcessorWithErrors{}
	processor.SetDelay(time.Millisecond * 100) // Короткая задержка

	client := NewMQTTClient(config, processor)

	mockMsg := &mockMessage{
		topic:   "msh/test/topic",
		payload: []byte(`{"test": "data"}`),
	}

	// Тест обработки сообщения с задержкой
	client.messageHandler(nil, mockMsg)
}

func TestMQTTClient_Disconnect_Connected(t *testing.T) {
	config := &adapters.MQTTConfigAdapter{Host: "localhost", Port: 1883}
	processor := &mocks.MockMessageProcessor{}

	client := NewMQTTClient(config, processor)

	// Создаем мок клиента
	mockClient := &mockMQTTClient{connected: true}
	client.client = mockClient

	client.Disconnect()

	if !mockClient.disconnectCalled {
		t.Error("Expected Disconnect to be called")
	}
}

func TestMQTTClient_OnConnect_WithTopics(t *testing.T) {
	config := &adapters.MQTTConfigAdapter{
		Host:   "localhost",
		Port:   1883,
		Topics: []string{"msh/test/+", "msh/another/+"},
	}
	processor := &mocks.MockMessageProcessor{}

	client := NewMQTTClient(config, processor)

	// Создаем мок клиента с успешной подпиской
	mockClient := &mockMQTTClient{subscribeSuccess: true}
	client.onConnect(mockClient)

	if len(mockClient.subscribedTopics) != 2 {
		t.Errorf("Expected 2 subscribed topics, got %d", len(mockClient.subscribedTopics))
	}
}

func TestMQTTClient_OnConnect_WithSubscribeError(t *testing.T) {
	config := &adapters.MQTTConfigAdapter{
		Host:   "localhost",
		Port:   1883,
		Topics: []string{"msh/test/+"},
	}
	processor := &mocks.MockMessageProcessor{}

	client := NewMQTTClient(config, processor)

	// Создаем мок клиента с ошибкой подписки
	mockClient := &mockMQTTClient{subscribeSuccess: false}
	client.onConnect(mockClient)

	if len(mockClient.subscribedTopics) != 1 {
		t.Errorf("Expected 1 attempted subscription, got %d", len(mockClient.subscribedTopics))
	}
}

// Мок MQTT клиента для тестирования
type mockMQTTClient struct {
	connected        bool
	disconnectCalled bool
	subscribeSuccess bool
	subscribedTopics []string
}

func (m *mockMQTTClient) IsConnected() bool {
	return m.connected
}

func (m *mockMQTTClient) IsConnectionOpen() bool {
	return m.connected
}

func (m *mockMQTTClient) Connect() mqtt.Token {
	return &mockToken{err: nil}
}

func (m *mockMQTTClient) Disconnect(quiesce uint) {
	m.disconnectCalled = true
	m.connected = false
}

func (m *mockMQTTClient) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	return &mockToken{err: nil}
}

func (m *mockMQTTClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	m.subscribedTopics = append(m.subscribedTopics, topic)
	if m.subscribeSuccess {
		return &mockToken{err: nil}
	}
	return &mockToken{err: context.DeadlineExceeded}
}

func (m *mockMQTTClient) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	return &mockToken{err: nil}
}

func (m *mockMQTTClient) Unsubscribe(topics ...string) mqtt.Token {
	return &mockToken{err: nil}
}

func (m *mockMQTTClient) AddRoute(topic string, callback mqtt.MessageHandler) {}

func (m *mockMQTTClient) OptionsReader() mqtt.ClientOptionsReader {
	return mqtt.ClientOptionsReader{}
}

// Мок токена для MQTT операций
type mockToken struct {
	err error
}

func (m *mockToken) Wait() bool {
	return true
}

func (m *mockToken) WaitTimeout(time.Duration) bool {
	return true
}

func (m *mockToken) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func (m *mockToken) Error() error {
	return m.err
}
