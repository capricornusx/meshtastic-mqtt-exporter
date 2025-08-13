package application

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/mocks"
)

func TestMeshtasticProcessor_RoundingMetrics(t *testing.T) {
	t.Parallel()
	collector := &mocks.MockMetricsCollector{}
	processor := NewMeshtasticProcessor(collector, nil, false, "")

	message := domain.MeshtasticMessage{
		From: 123456789,
		Type: "telemetry",
		Payload: map[string]interface{}{
			"voltage":             3.14159265,
			"temperature":         25.987654321,
			"relative_humidity":   67.123456789,
			"barometric_pressure": 1013.25987654,
		},
	}

	payload, _ := json.Marshal(message)
	err := processor.ProcessMessage(context.Background(), "msh/test/json/telemetry", payload)

	require.NoError(t, err)
	require.Len(t, collector.TelemetryData, 1)

	data := collector.TelemetryData[0]
	assert.Equal(t, 3.14, *data.Voltage)
	assert.Equal(t, 25.98, *data.Temperature)
	assert.Equal(t, 67.12, *data.RelativeHumidity)
	assert.Equal(t, 1013.25, *data.BarometricPressure)
}

func TestRoundToTwoDecimals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    float64
		expected float64
	}{
		{3.14159265, 3.14},
		{25.987654321, 25.99},
		{67.123456789, 67.12},
		{1013.25987654, 1013.26},
		{0.999, 1.0},
		{0.994, 0.99},
		{0.995, 1.0},
	}

	for _, test := range tests {
		result := roundToTwoDecimals(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func TestTruncateToTwoDecimals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    float64
		expected float64
	}{
		{3.14159265, 3.14},
		{25.987654321, 25.98},
		{67.123456789, 67.12},
		{1013.25987654, 1013.25},
		{0.999, 0.99},
		{0.994, 0.99},
		{0.995, 0.99},
		{-3.14159265, -3.14},
		{-25.987654321, -25.98},
	}

	for _, test := range tests {
		result := truncateToTwoDecimals(test.input)
		assert.Equal(t, test.expected, result)
	}
}
