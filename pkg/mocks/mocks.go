package mocks

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"meshtastic-exporter/pkg/domain"
)

// MockMetricsCollector базовый mock для MetricsCollector
type MockMetricsCollector struct {
	mu                     sync.RWMutex
	CollectTelemetryCalled bool
	CollectNodeInfoCalled  bool
	SaveStateCalled        bool
	LoadStateCalled        bool
	Registry               *prometheus.Registry
	TelemetryData          []domain.TelemetryData
	NodeInfoData           []domain.NodeInfo
	LastStateFile          string
}

func (m *MockMetricsCollector) CollectTelemetry(data domain.TelemetryData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CollectTelemetryCalled = true
	m.TelemetryData = append(m.TelemetryData, data)
	return nil
}

func (m *MockMetricsCollector) CollectNodeInfo(info domain.NodeInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CollectNodeInfoCalled = true
	m.NodeInfoData = append(m.NodeInfoData, info)
	return nil
}

func (m *MockMetricsCollector) GetRegistry() *prometheus.Registry {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Registry == nil {
		m.Registry = prometheus.NewRegistry()
	}
	return m.Registry
}

func (m *MockMetricsCollector) SaveState(filename string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SaveStateCalled = true
	m.LastStateFile = filename
	return nil
}

func (m *MockMetricsCollector) LoadState(filename string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LoadStateCalled = true
	m.LastStateFile = filename
	return nil
}

// MockMetricsCollectorWithErrors расширенный mock с возможностью симуляции ошибок
type MockMetricsCollectorWithErrors struct {
	MockMetricsCollector
	mu                    sync.RWMutex
	ShouldFailTelemetry   bool
	ShouldFailNodeInfo    bool
	ShouldFailSaveState   bool
	ShouldFailLoadState   bool
	TelemetryErrorMessage string
	NodeInfoErrorMessage  string
	SaveStateErrorMessage string
	LoadStateErrorMessage string
}

func (m *MockMetricsCollectorWithErrors) CollectTelemetry(data domain.TelemetryData) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.ShouldFailTelemetry {
		msg := m.TelemetryErrorMessage
		if msg == "" {
			msg = "telemetry collection failed"
		}
		return errors.New(msg)
	}
	return m.MockMetricsCollector.CollectTelemetry(data)
}

func (m *MockMetricsCollectorWithErrors) CollectNodeInfo(info domain.NodeInfo) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.ShouldFailNodeInfo {
		msg := m.NodeInfoErrorMessage
		if msg == "" {
			msg = "node info collection failed"
		}
		return errors.New(msg)
	}
	return m.MockMetricsCollector.CollectNodeInfo(info)
}

func (m *MockMetricsCollectorWithErrors) SaveState(filename string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.ShouldFailSaveState {
		msg := m.SaveStateErrorMessage
		if msg == "" {
			msg = "save state failed"
		}
		return errors.New(msg)
	}
	return m.MockMetricsCollector.SaveState(filename)
}

func (m *MockMetricsCollectorWithErrors) LoadState(filename string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.ShouldFailLoadState {
		msg := m.LoadStateErrorMessage
		if msg == "" {
			msg = "load state failed"
		}
		return errors.New(msg)
	}
	return m.MockMetricsCollector.LoadState(filename)
}

func (m *MockMetricsCollectorWithErrors) SetTelemetryError(shouldFail bool, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ShouldFailTelemetry = shouldFail
	m.TelemetryErrorMessage = message
}

func (m *MockMetricsCollectorWithErrors) SetNodeInfoError(shouldFail bool, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ShouldFailNodeInfo = shouldFail
	m.NodeInfoErrorMessage = message
}

// MockAlertSender базовый mock для AlertSender
type MockAlertSender struct {
	SendAlertCalled bool
	LastAlert       domain.Alert
}

func (m *MockAlertSender) SendAlert(ctx context.Context, alert domain.Alert) error {
	m.SendAlertCalled = true
	m.LastAlert = alert
	return nil
}

// MockAlertSenderWithErrors расширенный mock для AlertSender с ошибками
type MockAlertSenderWithErrors struct {
	MockAlertSender
	mu               sync.RWMutex
	ShouldFail       bool
	ErrorMessage     string
	SendDelay        time.Duration
	AlertsSent       []domain.Alert
	MaxAlertsToStore int
}

func (m *MockAlertSenderWithErrors) SendAlert(ctx context.Context, alert domain.Alert) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Симулируем задержку
	if m.SendDelay > 0 {
		select {
		case <-time.After(m.SendDelay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if m.ShouldFail {
		msg := m.ErrorMessage
		if msg == "" {
			msg = "alert sending failed"
		}
		return errors.New(msg)
	}

	// Сохраняем алерт
	m.AlertsSent = append(m.AlertsSent, alert)

	// Ограничиваем количество сохраненных алертов
	if m.MaxAlertsToStore > 0 && len(m.AlertsSent) > m.MaxAlertsToStore {
		m.AlertsSent = m.AlertsSent[len(m.AlertsSent)-m.MaxAlertsToStore:]
	}

	m.SendAlertCalled = true
	m.LastAlert = alert
	return nil
}

func (m *MockAlertSenderWithErrors) SetError(shouldFail bool, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ShouldFail = shouldFail
	m.ErrorMessage = message
}

func (m *MockAlertSenderWithErrors) SetDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SendDelay = delay
}

func (m *MockAlertSenderWithErrors) GetAlertsSent() []domain.Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]domain.Alert, len(m.AlertsSent))
	copy(result, m.AlertsSent)
	return result
}

func (m *MockAlertSenderWithErrors) ClearAlerts() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AlertsSent = nil
}

// MockMessageProcessor базовый mock для MessageProcessor
type MockMessageProcessor struct {
	ProcessMessageCalled bool
	LastTopic            string
	LastPayload          []byte
}

func (m *MockMessageProcessor) ProcessMessage(ctx context.Context, topic string, payload []byte) error {
	m.ProcessMessageCalled = true
	m.LastTopic = topic
	m.LastPayload = payload
	return nil
}

// ProcessedMessage информация об обработанном сообщении
type ProcessedMessage struct {
	Topic   string
	Payload []byte
	Time    time.Time
}

// MockMessageProcessorWithErrors расширенный mock для MessageProcessor
type MockMessageProcessorWithErrors struct {
	MockMessageProcessor
	mu                 sync.RWMutex
	ShouldFail         bool
	ErrorMessage       string
	ProcessDelay       time.Duration
	MessagesProcessed  []ProcessedMessage
	MaxMessagesToStore int
}

func (m *MockMessageProcessorWithErrors) ProcessMessage(ctx context.Context, topic string, payload []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Симулируем задержку обработки
	if m.ProcessDelay > 0 {
		select {
		case <-time.After(m.ProcessDelay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if m.ShouldFail {
		msg := m.ErrorMessage
		if msg == "" {
			msg = "message processing failed"
		}
		return errors.New(msg)
	}

	// Сохраняем обработанное сообщение
	m.MessagesProcessed = append(m.MessagesProcessed, ProcessedMessage{
		Topic:   topic,
		Payload: make([]byte, len(payload)),
		Time:    time.Now(),
	})
	copy(m.MessagesProcessed[len(m.MessagesProcessed)-1].Payload, payload)

	// Ограничиваем количество сохраненных сообщений
	if m.MaxMessagesToStore > 0 && len(m.MessagesProcessed) > m.MaxMessagesToStore {
		m.MessagesProcessed = m.MessagesProcessed[len(m.MessagesProcessed)-m.MaxMessagesToStore:]
	}

	m.ProcessMessageCalled = true
	m.LastTopic = topic
	m.LastPayload = make([]byte, len(payload))
	copy(m.LastPayload, payload)

	return nil
}

func (m *MockMessageProcessorWithErrors) SetError(shouldFail bool, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ShouldFail = shouldFail
	m.ErrorMessage = message
}

func (m *MockMessageProcessorWithErrors) SetDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ProcessDelay = delay
}

func (m *MockMessageProcessorWithErrors) GetMessagesProcessed() []ProcessedMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]ProcessedMessage, len(m.MessagesProcessed))
	copy(result, m.MessagesProcessed)
	return result
}

func (m *MockMessageProcessorWithErrors) ClearMessages() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesProcessed = nil
}
