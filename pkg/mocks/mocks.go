package mocks

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"meshtastic-exporter/pkg/domain"
)

type MockMetricsCollector struct {
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
	m.CollectTelemetryCalled = true
	m.TelemetryData = append(m.TelemetryData, data)
	return nil
}

func (m *MockMetricsCollector) CollectNodeInfo(info domain.NodeInfo) error {
	m.CollectNodeInfoCalled = true
	m.NodeInfoData = append(m.NodeInfoData, info)
	return nil
}

func (m *MockMetricsCollector) GetRegistry() *prometheus.Registry {
	if m.Registry == nil {
		m.Registry = prometheus.NewRegistry()
	}
	return m.Registry
}

func (m *MockMetricsCollector) SaveState(filename string) error {
	m.SaveStateCalled = true
	m.LastStateFile = filename
	return nil
}

func (m *MockMetricsCollector) LoadState(filename string) error {
	m.LoadStateCalled = true
	m.LastStateFile = filename
	return nil
}

type MockAlertSender struct {
	SendAlertCalled bool
	LastAlert       domain.Alert
}

func (m *MockAlertSender) SendAlert(ctx context.Context, alert domain.Alert) error {
	m.SendAlertCalled = true
	m.LastAlert = alert
	return nil
}

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
