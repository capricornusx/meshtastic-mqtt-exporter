package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/mocks"
)

func TestMeshtasticProcessor_ProcessMessage_TextMessage(t *testing.T) {
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	payload := []byte(`{
		"from": 123456789,
		"type": "text",
		"payload": {
			"text": "Hello mesh network!",
			"channel": 0
		}
	}`)

	err := processor.ProcessMessage(context.Background(), "msh/test", payload)

	require.NoError(t, err)
	assert.True(t, mockCollector.UpdateNodeLastSeenCalled)
}

func TestMeshtasticProcessor_ProcessMessage_Position(t *testing.T) {
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	payload := []byte(`{
		"from": 123456789,
		"type": "position",
		"payload": {
			"latitude_i": 559748544,
			"longitude_i": 373418112,
			"altitude": 150,
			"sats_in_view": 8,
			"precision_bits": 32
		}
	}`)

	err := processor.ProcessMessage(context.Background(), "msh/test", payload)

	require.NoError(t, err)
	assert.True(t, mockCollector.UpdateNodeLastSeenCalled)
}

func TestMeshtasticProcessor_ProcessMessage_Waypoint(t *testing.T) {
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	payload := []byte(`{
		"from": 123456789,
		"type": "waypoint",
		"payload": {
			"id": 1,
			"latitude_i": 559748544,
			"longitude_i": 373418112,
			"expire": 0,
			"locked_to": 0,
			"name": "Test Waypoint",
			"description": "Test waypoint description",
			"icon": 1
		}
	}`)

	err := processor.ProcessMessage(context.Background(), "msh/test", payload)

	require.NoError(t, err)
	assert.True(t, mockCollector.UpdateNodeLastSeenCalled)
}

func TestMeshtasticProcessor_ProcessMessage_NeighborInfo(t *testing.T) {
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	payload := []byte(`{
		"from": 123456789,
		"type": "neighborinfo",
		"payload": {
			"node_id": 987654321,
			"snr": 12.5,
			"last_rx_time": 1640995200,
			"node_broadcast_interval_secs": 900
		}
	}`)

	err := processor.ProcessMessage(context.Background(), "msh/test", payload)

	require.NoError(t, err)
	assert.True(t, mockCollector.UpdateNodeLastSeenCalled)
}

func TestMeshtasticProcessor_DetermineTelemetryType(t *testing.T) {
	t.Parallel()
	processor := &MeshtasticProcessor{}

	testCases := []struct {
		name     string
		payload  map[string]interface{}
		expected string
	}{
		{
			name:     "environment metrics - temperature",
			payload:  map[string]interface{}{"temperature": 25.5},
			expected: domain.TelemetryTypeEnvironment,
		},
		{
			name:     "environment metrics - humidity",
			payload:  map[string]interface{}{"relative_humidity": 65.2},
			expected: domain.TelemetryTypeEnvironment,
		},
		{
			name:     "environment metrics - pressure",
			payload:  map[string]interface{}{"barometric_pressure": 1013.25},
			expected: domain.TelemetryTypeEnvironment,
		},
		{
			name:     "power metrics - ch1 voltage",
			payload:  map[string]interface{}{"ch1_voltage": 12.5},
			expected: domain.TelemetryTypePower,
		},
		{
			name:     "power metrics - ch1 current",
			payload:  map[string]interface{}{"ch1_current": 2.1},
			expected: domain.TelemetryTypePower,
		},
		{
			name:     "device metrics - battery",
			payload:  map[string]interface{}{"battery_level": 85.5},
			expected: domain.TelemetryTypeDevice,
		},
		{
			name:     "device metrics - voltage",
			payload:  map[string]interface{}{"voltage": 4.1},
			expected: domain.TelemetryTypeDevice,
		},
		{
			name:     "device metrics - empty",
			payload:  map[string]interface{}{},
			expected: domain.TelemetryTypeDevice,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := processor.determineTelemetryType(tc.payload)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMeshtasticProcessor_ProcessMessage_TelemetryWithType(t *testing.T) {
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "")

	// Environment telemetry
	envPayload := []byte(`{
		"from": 123456789,
		"type": "telemetry",
		"payload": {
			"temperature": 23.4,
			"relative_humidity": 65.2,
			"barometric_pressure": 1013.25
		}
	}`)

	err := processor.ProcessMessage(context.Background(), "msh/test", envPayload)
	require.NoError(t, err)
	assert.True(t, mockCollector.CollectTelemetryCalled)

	// Power telemetry
	powerPayload := []byte(`{
		"from": 123456789,
		"type": "telemetry",
		"payload": {
			"ch1_voltage": 12.5,
			"ch1_current": 2.1,
			"ch2_voltage": 5.0
		}
	}`)

	err = processor.ProcessMessage(context.Background(), "msh/test", powerPayload)
	require.NoError(t, err)
}
