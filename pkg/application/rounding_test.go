package application

import (
	"context"
	"encoding/json"
	"testing"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/mocks"
)

func TestMeshtasticProcessor_RoundingMetrics(t *testing.T) {
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

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(collector.TelemetryData) != 1 {
		t.Fatalf("Expected 1 telemetry data, got %d", len(collector.TelemetryData))
	}

	data := collector.TelemetryData[0]
	assertMetricValue(t, "voltage", data.Voltage, 3.14)
	assertMetricValue(t, "temperature", data.Temperature, 25.99)
	assertMetricValue(t, "humidity", data.RelativeHumidity, 67.12)
	assertMetricValue(t, "pressure", data.BarometricPressure, 1013.26)
}

func assertMetricValue(t *testing.T, name string, actual *float64, expected float64) {
	if actual == nil || *actual != expected {
		t.Errorf("Expected %s %g, got %v", name, expected, actual)
	}
}

func TestRoundToTwoDecimals(t *testing.T) {
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
		if result != test.expected {
			t.Errorf("roundToTwoDecimals(%f) = %f, expected %f", test.input, result, test.expected)
		}
	}
}
