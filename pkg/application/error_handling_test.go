package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/mocks"
)

func TestMeshtasticProcessor_ErrorHandling(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		telemetryError        bool
		nodeInfoError         bool
		payload               string
		expectProcessingError bool
	}{
		{
			name:                  "telemetry_collection_error",
			telemetryError:        true,
			payload:               `{"from": 123456789, "type": "telemetry", "payload": {"battery_level": 85.5}}`,
			expectProcessingError: true,
		},
		{
			name:                  "nodeinfo_collection_error",
			nodeInfoError:         true,
			payload:               `{"from": 123456789, "type": "nodeinfo", "payload": {"longname": "Test"}}`,
			expectProcessingError: true,
		},
		{
			name:                  "no_errors",
			payload:               `{"from": 123456789, "type": "telemetry", "payload": {"battery_level": 85.5}}`,
			expectProcessingError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockCollector := &mocks.MockMetricsCollectorWithErrors{}
			mockAlerter := &mocks.MockAlertSenderWithErrors{}

			if tt.telemetryError {
				mockCollector.SetTelemetryError(true, "telemetry failed")
			}
			if tt.nodeInfoError {
				mockCollector.SetNodeInfoError(true, "nodeinfo failed")
			}

			processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

			err := processor.ProcessMessage(context.Background(), "msh/test", []byte(tt.payload))

			if tt.expectProcessingError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

const lowBatteryPayload = `{"from": 123456789, "type": "telemetry", "payload": {"battery_level": 5.0}}`

func TestMeshtasticProcessor_AlertingErrors(t *testing.T) {
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollectorWithErrors{}
	mockAlerter := &mocks.MockAlertSenderWithErrors{}

	// Настраиваем alerter на ошибку
	mockAlerter.SetError(true, "alert sending failed")

	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	// Отправляем сообщение, которое может вызвать алерт
	payload := lowBatteryPayload

	err := processor.ProcessMessage(context.Background(), "msh/test", []byte(payload))

	// Ошибка алертинга не должна прерывать обработку основного сообщения
	require.NoError(t, err)
	assert.True(t, mockCollector.CollectTelemetryCalled)
}

func TestMeshtasticProcessor_ContextCancellation(t *testing.T) {
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollectorWithErrors{}
	mockAlerter := &mocks.MockAlertSenderWithErrors{}

	// Устанавливаем задержку для симуляции долгой обработки
	mockAlerter.SetDelay(1 * time.Second)

	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	payload := lowBatteryPayload

	start := time.Now()
	err := processor.ProcessMessage(ctx, "msh/test", []byte(payload))
	duration := time.Since(start)

	// Проверяем, что обработка завершилась быстро из-за отмены контекста
	assert.Less(t, duration, 500*time.Millisecond)

	// Основная обработка должна пройти успешно
	require.NoError(t, err)
	assert.True(t, mockCollector.CollectTelemetryCalled)
}

func TestMeshtasticProcessor_ConcurrentErrorHandling(t *testing.T) {
	mockCollector := &mocks.MockMetricsCollectorWithErrors{}
	mockAlerter := &mocks.MockAlertSenderWithErrors{}

	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	const numGoroutines = 20
	const numMessages = 10

	// Канал для сбора ошибок
	errors := make(chan error, numGoroutines*numMessages)

	// Запускаем множество горутин
	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			for j := 0; j < numMessages; j++ {
				// Каждая 5-я операция будет с ошибкой
				if (routineID*numMessages+j)%5 == 0 {
					mockCollector.SetTelemetryError(true, "concurrent error")
				} else {
					mockCollector.SetTelemetryError(false, "")
				}

				payload := fmt.Sprintf(`{"from": %d, "type": "telemetry", "payload": {"battery_level": %d}}`,
					routineID*1000+j, j%100)

				err := processor.ProcessMessage(context.Background(), "msh/test", []byte(payload))
				errors <- err
			}
		}(i)
	}

	// Собираем результаты
	var errorCount int
	for i := 0; i < numGoroutines*numMessages; i++ {
		err := <-errors
		if err != nil {
			errorCount++
		}
	}

	// Проверяем, что примерно 20% операций завершились с ошибкой (допускаем погрешность)
	expectedErrors := numGoroutines * numMessages / 5
	assert.InDelta(t, expectedErrors, errorCount, float64(expectedErrors)*0.3) // 30% погрешность для concurrent операций
}
