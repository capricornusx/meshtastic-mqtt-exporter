package application

import (
	"testing"
	"time"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/mocks"
)

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

	if data.BatteryLevel == nil || *data.BatteryLevel != 85.5 {
		t.Error("BatteryLevel not extracted correctly")
	}
	if data.RSSI == nil || *data.RSSI != -95.2 {
		t.Error("RSSI not extracted correctly")
	}
	if data.SNR == nil || *data.SNR != -8.5 {
		t.Error("SNR not extracted correctly")
	}
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

	if data.BatteryLevel == nil || *data.BatteryLevel != 75.0 {
		t.Error("BatteryLevel not extracted")
	}
	if data.Voltage == nil || *data.Voltage != 3.8 {
		t.Error("Voltage not extracted")
	}
	if data.UptimeSeconds == nil || *data.UptimeSeconds != 7200.0 {
		t.Error("UptimeSeconds not extracted")
	}
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

	if data.Temperature == nil || *data.Temperature != 25.5 {
		t.Error("Temperature not extracted")
	}
	if data.RelativeHumidity == nil || *data.RelativeHumidity != 60.0 {
		t.Error("RelativeHumidity not extracted")
	}
	if data.BarometricPressure == nil || *data.BarometricPressure != 1015.0 {
		t.Error("BarometricPressure not extracted")
	}
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

	if data.ChannelUtilization == nil || *data.ChannelUtilization != 15.0 {
		t.Error("ChannelUtilization not extracted")
	}
	if data.AirUtilTx == nil || *data.AirUtilTx != 10.0 {
		t.Error("AirUtilTx not extracted")
	}
	if data.RSSI == nil || *data.RSSI != -90.0 {
		t.Error("RSSI not extracted")
	}
	if data.SNR == nil || *data.SNR != -5.0 {
		t.Error("SNR not extracted")
	}
}
