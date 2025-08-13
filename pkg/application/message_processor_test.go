package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/mocks"
)

func TestMeshtasticProcessor_ProcessMessage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		payload        string
		expectError    bool
		expectTelCall  bool
		expectNodeCall bool
	}{
		{
			name: "valid_telemetry",
			payload: fmt.Sprintf(`{
				"from": 123456789,
				"type": "%s",
				"payload": {
					"battery_level": 85.5,
					"temperature": 23.4,
					"voltage": 4.1
				}
			}`, domain.MessageTypeTelemetry),
			expectTelCall: true,
		},
		{
			name: "valid_nodeinfo",
			payload: fmt.Sprintf(`{
				"from": 987654321,
				"type": "%s",
				"payload": {
					"longname": "Test Node",
					"shortname": "TN01",
					"hardware": 1.0,
					"role": 2.0
				}
			}`, domain.MessageTypeNodeInfo),
			expectNodeCall: true,
		},
		{
			name:    "invalid_json",
			payload: `invalid json`,
		},
		{
			name:        "zero_from_node",
			payload:     fmt.Sprintf(`{"from": 0, "type": "%s", "payload": {"battery_level": 85.5}}`, domain.MessageTypeTelemetry),
			expectError: true,
		},
		{
			name:          "max_uint32_node_id",
			payload:       `{"from": 4294967295, "type": "telemetry", "payload": {"battery_level": 50.0}}`,
			expectTelCall: true,
		},
		{
			name:          "negative_battery_level",
			payload:       `{"from": 123456789, "type": "telemetry", "payload": {"battery_level": -10.5}}`,
			expectTelCall: true,
		},
		{
			name:          "extreme_temperature",
			payload:       `{"from": 123456789, "type": "telemetry", "payload": {"temperature": -273.15}}`,
			expectTelCall: true,
		},
		{
			name:          "null_payload",
			payload:       `{"from": 123456789, "type": "telemetry", "payload": null}`,
			expectTelCall: true,
		},
		{
			name:          "empty_payload",
			payload:       `{"from": 123456789, "type": "telemetry", "payload": {}}`,
			expectTelCall: true,
		},
		{
			name:        "string_node_id",
			payload:     `{"from": "invalid", "type": "telemetry", "payload": {"battery_level": 50.0}}`,
			expectError: true,
		},
		{
			name:    "missing_type",
			payload: `{"from": 123456789, "payload": {"battery_level": 50.0}}`,
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

			assert.Equal(t, tt.expectTelCall, mockCollector.CollectTelemetryCalled)
			assert.Equal(t, tt.expectNodeCall, mockCollector.CollectNodeInfoCalled)
		})
	}
}

func TestMeshtasticProcessor_LogAllMessages(t *testing.T) {
	t.Parallel()
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, true, "msh/#")

	payload := []byte(fmt.Sprintf(`{"from": 123456789, "type": "%s"}`, domain.MessageTypeTelemetry))

	err := processor.ProcessMessage(context.Background(), "msh/test", payload)

	require.NoError(t, err)
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

	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() {
			done <- processor.ProcessMessage(context.Background(), "msh/test", []byte(payload))
		}()
	}

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
			expectError: true,
		},
		{
			name:        "invalid_value",
			payload:     `{"from": 123456789, "type": "telemetry", "payload": {"battery_level": }}`,
			expectError: true,
		},
		{
			name:        "nan_value",
			payload:     `{"from": 123456789, "type": "telemetry", "payload": {"battery_level": NaN}}`,
			expectError: true,
		},
		{
			name:    "boolean_instead_of_number",
			payload: `{"from": 123456789, "type": "telemetry", "payload": {"battery_level": true}}`,
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

func TestMeshtasticProcessor_RSSI_SNR_Extraction(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		payload string
		expRSSI *float64
		expSNR  *float64
	}{
		{
			name:    "rssi_snr_in_root",
			payload: `{"from": 123, "type": "telemetry", "payload": {"battery_level": 85}, "rssi": -95.5, "snr": 12.3}`,
			expRSSI: &[]float64{-95.5}[0],
			expSNR:  &[]float64{12.3}[0],
		},
		{
			name:    "rssi_snr_in_payload",
			payload: `{"from": 123, "type": "telemetry", "payload": {"battery_level": 85, "rssi": -88.2, "snr": 8.7}}`,
			expRSSI: &[]float64{-88.2}[0],
			expSNR:  &[]float64{8.7}[0],
		},
		{
			name:    "no_rssi_snr",
			payload: `{"from": 123, "type": "telemetry", "payload": {"battery_level": 85}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockCollector := &mocks.MockMetricsCollector{}
			processor := NewMeshtasticProcessor(mockCollector, nil, false, "")

			err := processor.ProcessMessage(context.Background(), "msh/test", []byte(tt.payload))
			require.NoError(t, err)

			if len(mockCollector.TelemetryData) > 0 {
				data := mockCollector.TelemetryData[0]
				if tt.expRSSI != nil {
					assert.Equal(t, *tt.expRSSI, *data.RSSI)
				} else {
					assert.Nil(t, data.RSSI)
				}
				if tt.expSNR != nil {
					assert.Equal(t, *tt.expSNR, *data.SNR)
				} else {
					assert.Nil(t, data.SNR)
				}
			}
		})
	}
}

func TestExtractTelemetryFields(t *testing.T) {
	t.Parallel()
	collector := &mocks.MockMetricsCollector{}
	processor := NewMeshtasticProcessor(collector, nil, false, "msh/#")

	payload := map[string]interface{}{
		"battery_level":       85.5,
		"voltage":             3.7,
		"temperature":         23.4,
		"relative_humidity":   65.2,
		"barometric_pressure": 1013.25,
		"channel_utilization": 12.5,
		"air_util_tx":         8.3,
		"uptime_seconds":      3600.0,
		"rssi":                -95.2,
		"snr":                 -8.5,
	}

	data := domain.TelemetryData{NodeID: "123", Timestamp: time.Now()}
	processor.extractTelemetryFields(&data, payload)

	assert.Equal(t, 85.5, *data.BatteryLevel)
	assert.Equal(t, -95.2, *data.RSSI)
	assert.Equal(t, -8.5, *data.SNR)
}

func TestExtractBasicFields(t *testing.T) {
	t.Parallel()
	processor := NewMeshtasticProcessor(nil, nil, false, "")

	payload := map[string]interface{}{
		"battery_level":  75.0,
		"voltage":        3.8,
		"uptime_seconds": 7200.0,
	}

	data := domain.TelemetryData{}
	processor.extractBasicFields(&data, payload)

	assert.Equal(t, 75.0, *data.BatteryLevel)
	assert.Equal(t, 3.8, *data.Voltage)
	assert.Equal(t, 7200.0, *data.UptimeSeconds)
}

func TestExtractEnvironmentalFields(t *testing.T) {
	t.Parallel()
	processor := NewMeshtasticProcessor(nil, nil, false, "")

	payload := map[string]interface{}{
		"temperature":         25.5,
		"relative_humidity":   60.0,
		"barometric_pressure": 1015.0,
	}

	data := domain.TelemetryData{}
	processor.extractEnvironmentalFields(&data, payload)

	assert.Equal(t, 25.5, *data.Temperature)
	assert.Equal(t, 60.0, *data.RelativeHumidity)
	assert.Equal(t, 1015.0, *data.BarometricPressure)
}

func TestExtractNetworkFields(t *testing.T) {
	t.Parallel()
	processor := NewMeshtasticProcessor(nil, nil, false, "")

	payload := map[string]interface{}{
		"channel_utilization": 15.0,
		"air_util_tx":         10.0,
		"rssi":                -90.0,
		"snr":                 -5.0,
	}

	data := domain.TelemetryData{}
	processor.extractNetworkFields(&data, payload)

	assert.Equal(t, 15.0, *data.ChannelUtilization)
	assert.Equal(t, 10.0, *data.AirUtilTx)
	assert.Equal(t, -90.0, *data.RSSI)
	assert.Equal(t, -5.0, *data.SNR)
}
