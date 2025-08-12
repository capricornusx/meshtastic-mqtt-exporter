package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/mocks"
)

func TestMeshtasticProcessor_ProcessMessage_WithTopLevelRSSI(t *testing.T) {
	mockCollector := &mocks.MockMetricsCollector{}
	mockAlerter := &mocks.MockAlertSender{}
	processor := NewMeshtasticProcessor(mockCollector, mockAlerter, false, "msh/#")

	// Сырое сообщение с RSSI/SNR на верхнем уровне
	payload := `{
		"channel": 0,
		"from": 4187204132,
		"hop_start": 2,
		"hops_away": 0,
		"id": 3928256623,
		"payload": {
			"air_util_tx": 0.934000015258789,
			"battery_level": 51,
			"channel_utilization": 5.36499977111816,
			"uptime_seconds": 661,
			"voltage": 3.7279999256134
		},
		"rssi": -40,
		"sender": "!f992bd54",
		"snr": 10.5,
		"timestamp": 1754956324,
		"to": 4294967295,
		"type": "telemetry"
	}`

	err := processor.ProcessMessage(context.Background(), "msh/2/json/LongFast/!f992bd54", []byte(payload))

	assert.NoError(t, err)
	assert.True(t, mockCollector.CollectTelemetryCalled)
	assert.Len(t, mockCollector.TelemetryData, 1)

	data := mockCollector.TelemetryData[0]
	assert.Equal(t, "4187204132", data.NodeID)
	assert.NotNil(t, data.RSSI)
	assert.Equal(t, -40.0, *data.RSSI)
	assert.NotNil(t, data.SNR)
	assert.Equal(t, 10.5, *data.SNR)
	assert.NotNil(t, data.BatteryLevel)
	assert.Equal(t, 51.0, *data.BatteryLevel)
}

func TestExtractTopLevelFields(t *testing.T) {
	processor := NewMeshtasticProcessor(nil, nil, false, "msh/#")

	rssi := -40.0
	snr := 10.5
	msg := domain.MeshtasticMessage{
		RSSI: &rssi,
		SNR:  &snr,
	}

	data := &domain.TelemetryData{}
	processor.extractTopLevelFields(data, msg)

	assert.NotNil(t, data.RSSI)
	assert.Equal(t, -40.0, *data.RSSI)
	assert.NotNil(t, data.SNR)
	assert.Equal(t, 10.5, *data.SNR)
}
