package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/mocks"
)

func TestMeshtasticProcessor_EdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		payload     string
		expectError bool
		expectCall  bool
	}{
		{
			name:        "max_uint32_node_id",
			payload:     `{"from": 4294967295, "type": "telemetry", "payload": {"battery_level": 50.0}}`,
			expectError: false,
			expectCall:  true,
		},
		{
			name:        "negative_battery_level",
			payload:     `{"from": 123456789, "type": "telemetry", "payload": {"battery_level": -10.5}}`,
			expectError: false,
			expectCall:  true,
		},
		{
			name:        "extreme_temperature",
			payload:     `{"from": 123456789, "type": "telemetry", "payload": {"temperature": -273.15}}`,
			expectError: false,
			expectCall:  true,
		},
		{
			name:        "infinity_value",
			payload:     `{"from": 123456789, "type": "telemetry", "payload": {"voltage": "Infinity"}}`,
			expectError: false,
			expectCall:  true,
		},
		{
			name:        "null_payload",
			payload:     `{"from": 123456789, "type": "telemetry", "payload": null}`,
			expectError: false,
			expectCall:  true,
		},
		{
			name:        "empty_payload",
			payload:     `{"from": 123456789, "type": "telemetry", "payload": {}}`,
			expectError: false,
			expectCall:  true,
		},
		{
			name:        "string_node_id",
			payload:     `{"from": "invalid", "type": "telemetry", "payload": {"battery_level": 50.0}}`,
			expectError: true,
			expectCall:  false,
		},
		{
			name:        "missing_type",
			payload:     `{"from": 123456789, "payload": {"battery_level": 50.0}}`,
			expectError: false,
			expectCall:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockCollector := &mocks.MockMetricsCollector{}
			mockAlerter := &mocks.MockAlertSender{}
			processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

			err := processor.ProcessMessage(context.Background(), "msh/test", []byte(tt.payload))

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectCall, mockCollector.CollectTelemetryCalled)
		})
	}
}

func TestMeshtasticProcessor_ExtremeValues(t *testing.T) {
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	payload := `{
		"from": 123456789,
		"type": "telemetry",
		"payload": {
			"battery_level": 999999.99,
			"temperature": -999.99,
			"voltage": 0.0,
			"relative_humidity": 150.0,
			"barometric_pressure": -100.0
		}
	}`

	err := processor.ProcessMessage(context.Background(), "msh/test", []byte(payload))

	require.NoError(t, err)
	assert.True(t, mockCollector.CollectTelemetryCalled)
	assert.Len(t, mockCollector.TelemetryData, 1)

	data := mockCollector.TelemetryData[0]
	assert.Equal(t, "123456789", data.NodeID)
	assert.Equal(t, 999999.99, *data.BatteryLevel)
	assert.Equal(t, -999.99, *data.Temperature)
	assert.Equal(t, 0.0, *data.Voltage)
}

func TestMeshtasticProcessor_ConcurrentProcessing(t *testing.T) {
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	payload := `{"from": 123456789, "type": "telemetry", "payload": {"battery_level": 85.5}}`

	// Запускаем 10 горутин одновременно
	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() {
			done <- processor.ProcessMessage(context.Background(), "msh/test", []byte(payload))
		}()
	}

	// Проверяем, что все завершились без ошибок
	for i := 0; i < 10; i++ {
		err := <-done
		require.NoError(t, err)
	}

	assert.True(t, mockCollector.CollectTelemetryCalled)
}

func TestMeshtasticProcessor_LargePayload(t *testing.T) {
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	// Создаем большой payload с множеством полей
	largePayload := `{
		"from": 123456789,
		"type": "telemetry",
		"payload": {
			"battery_level": 85.5,
			"temperature": 23.4,
			"voltage": 4.1,
			"relative_humidity": 65.2,
			"barometric_pressure": 1013.25,
			"channel_utilization": 12.5,
			"air_util_tx": 8.3,
			"uptime_seconds": 86400,
			"extra_field_1": 1.0,
			"extra_field_2": 2.0,
			"extra_field_3": 3.0
		},
		"rssi": -85.5,
		"snr": 12.3
	}`

	err := processor.ProcessMessage(context.Background(), "msh/test", []byte(largePayload))

	require.NoError(t, err)
	assert.True(t, mockCollector.CollectTelemetryCalled)
}

func TestMeshtasticProcessor_MalformedJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		payload     string
		expectError bool
	}{
		{
			name:        "missing_closing_brace",
			payload:     `{"from": 123456789, "type": "telemetry", "payload": {"battery_level": 85.5}`,
			expectError: true, // Invalid JSON causes validation error
		},
		{
			name:        "invalid_value",
			payload:     `{"from": 123456789, "type": "telemetry", "payload": {"battery_level": }}`,
			expectError: true, // Invalid JSON causes validation error
		},
		{
			name:        "nan_value",
			payload:     `{"from": 123456789, "type": "telemetry", "payload": {"battery_level": NaN}}`,
			expectError: true, // Invalid JSON causes validation error
		},
		{
			name:        "boolean_instead_of_number",
			payload:     `{"from": 123456789, "type": "telemetry", "payload": {"battery_level": true}}`,
			expectError: false, // Valid JSON, but wrong type - should be processed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockCollector := &mocks.MockMetricsCollector{}
			mockAlerter := &mocks.MockAlertSender{}
			processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

			err := processor.ProcessMessage(context.Background(), "msh/test", []byte(tt.payload))

			if tt.expectError {
				require.Error(t, err)
				assert.False(t, mockCollector.CollectTelemetryCalled)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
