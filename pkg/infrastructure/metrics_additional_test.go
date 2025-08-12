package infrastructure

import (
	"testing"
	"time"

	"meshtastic-exporter/pkg/domain"
)

func TestSetTelemetryMetrics(t *testing.T) {
	collector := NewPrometheusCollector()

	data := domain.TelemetryData{
		NodeID:             "123",
		BatteryLevel:       floatPtr(85.5),
		Voltage:            floatPtr(3.7),
		Temperature:        floatPtr(23.4),
		RelativeHumidity:   floatPtr(65.2),
		BarometricPressure: floatPtr(1013.25),
		ChannelUtilization: floatPtr(12.5),
		AirUtilTx:          floatPtr(8.3),
		UptimeSeconds:      floatPtr(3600.0),
		RSSI:               floatPtr(-95.2),
		SNR:                floatPtr(-8.5),
		Timestamp:          time.Now(),
	}

	collector.setTelemetryMetrics(data)

	// Проверяем, что метрики установлены (через registry)
	registry := collector.GetRegistry()
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	foundRSSI := false
	foundSNR := false
	for _, mf := range metricFamilies {
		if mf.GetName() == domain.MetricRSSI {
			foundRSSI = true
		}
		if mf.GetName() == domain.MetricSNR {
			foundSNR = true
		}
	}

	if !foundRSSI {
		t.Error("RSSI metric not found")
	}
	if !foundSNR {
		t.Error("SNR metric not found")
	}
}

func TestSetBasicMetrics(t *testing.T) {
	collector := NewPrometheusCollector()

	data := domain.TelemetryData{
		NodeID:        "123",
		BatteryLevel:  floatPtr(75.0),
		Voltage:       floatPtr(3.8),
		UptimeSeconds: floatPtr(7200.0),
	}

	collector.setBasicMetrics(data)

	registry := collector.GetRegistry()
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	foundBattery := false
	for _, mf := range metricFamilies {
		if mf.GetName() == domain.MetricBatteryLevel {
			foundBattery = true
			break
		}
	}

	if !foundBattery {
		t.Error("Battery metric not found")
	}
}

func TestSetEnvironmentalMetrics(t *testing.T) {
	collector := NewPrometheusCollector()

	data := domain.TelemetryData{
		NodeID:             "123",
		Temperature:        floatPtr(25.5),
		RelativeHumidity:   floatPtr(60.0),
		BarometricPressure: floatPtr(1015.0),
	}

	collector.setEnvironmentalMetrics(data)

	registry := collector.GetRegistry()
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	foundTemp := false
	for _, mf := range metricFamilies {
		if mf.GetName() == domain.MetricTemperature {
			foundTemp = true
			break
		}
	}

	if !foundTemp {
		t.Error("Temperature metric not found")
	}
}

func TestSetNetworkMetrics(t *testing.T) {
	collector := NewPrometheusCollector()

	data := domain.TelemetryData{
		NodeID:             "123",
		ChannelUtilization: floatPtr(15.0),
		AirUtilTx:          floatPtr(10.0),
		RSSI:               floatPtr(-90.0),
		SNR:                floatPtr(-5.0),
	}

	collector.setNetworkMetrics(data)

	registry := collector.GetRegistry()
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	foundRSSI := false
	foundSNR := false
	for _, mf := range metricFamilies {
		if mf.GetName() == domain.MetricRSSI {
			foundRSSI = true
		}
		if mf.GetName() == domain.MetricSNR {
			foundSNR = true
		}
	}

	if !foundRSSI {
		t.Error("RSSI metric not found")
	}
	if !foundSNR {
		t.Error("SNR metric not found")
	}
}
