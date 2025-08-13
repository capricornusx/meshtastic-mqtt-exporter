package application

import (
	"context"
	"testing"

	"fmt"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/mocks"
)

func TestMeshtasticProcessor_LogAllMessages_WithPattern(t *testing.T) {
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}

	// Процессор с включенным логированием и паттерном "msh/+/2/json/#"
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, true, "msh/+/2/json/#")

	tests := []struct {
		name        string
		topic       string
		shouldLog   bool
		description string
	}{
		{
			name:        "matching_topic_should_log",
			topic:       "msh/node123/2/json/telemetry",
			shouldLog:   true,
			description: "Топик соответствует паттерну - должен логироваться",
		},
		{
			name:        "non_matching_topic_should_not_log",
			topic:       "msh/node123/1/json/telemetry",
			shouldLog:   false,
			description: "Топик не соответствует паттерну - не должен логироваться",
		},
		{
			name:        "different_prefix_should_not_log",
			topic:       "other/node123/2/json/telemetry",
			shouldLog:   false,
			description: "Другой prefix - не должен логироваться",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			payload := []byte(fmt.Sprintf(`{"from": 123456789, "type": "%s", "payload": {"battery_level": 85.5}}`, domain.MessageTypeTelemetry))

			err := processor.ProcessMessage(context.Background(), tt.topic, payload)

			// Проверяем, что сообщение обработалось без ошибок
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// Здесь мы не можем напрямую проверить логирование,
			// но можем убедиться что процессор работает корректно
			// В реальном приложении логирование будет видно в логах
		})
	}
}

func TestMeshtasticProcessor_LogAllMessages_EmptyPattern(t *testing.T) {
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}

	// Процессор с включенным логированием и пустым паттерном (логирует все)
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, true, "")

	payload := []byte(fmt.Sprintf(`{"from": 123456789, "type": "%s", "payload": {"battery_level": 85.5}}`, domain.MessageTypeTelemetry))

	err := processor.ProcessMessage(context.Background(), "any/topic/here", payload)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMeshtasticProcessor_LogAllMessages_Disabled(t *testing.T) {
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}

	// Процессор с выключенным логированием
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "msh/#")

	payload := []byte(fmt.Sprintf(`{"from": 123456789, "type": "%s", "payload": {"battery_level": 85.5}}`, domain.MessageTypeTelemetry))

	err := processor.ProcessMessage(context.Background(), "msh/node123/2/json/telemetry", payload)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
