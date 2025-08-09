package main

import (
	"os"
	"testing"

	"meshtastic-exporter/pkg/exporter"
)

func TestStandaloneConfigLoading(t *testing.T) {
	configContent := `
mqtt:
  host: mqtt.example.com
  port: 8883
  username: testuser
  password: testpass
  tls: true
prometheus:
  enabled: true
  port: 9090
  host: 0.0.0.0
`
	tmpFile, err := os.CreateTemp("", "standalone-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	config, err := exporter.LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.MQTT.Host != "mqtt.example.com" {
		t.Errorf("Expected host mqtt.example.com, got %s", config.MQTT.Host)
	}
	if config.MQTT.Port != 8883 {
		t.Errorf("Expected port 8883, got %d", config.MQTT.Port)
	}
	if config.MQTT.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", config.MQTT.Username)
	}
	if !config.MQTT.TLS {
		t.Error("Expected TLS to be enabled")
	}
	if config.Prometheus.Host != "0.0.0.0" {
		t.Errorf("Expected prometheus host 0.0.0.0, got %s", config.Prometheus.Host)
	}
}

func TestStandaloneDefaults(t *testing.T) {
	config, err := exporter.LoadConfig("non-existent.yaml")
	if err != nil {
		t.Fatalf("Expected no error with missing config, got: %v", err)
	}

	// Verify defaults are set correctly for standalone mode
	if config.MQTT.Host != "localhost" {
		t.Errorf("Expected default host localhost, got %s", config.MQTT.Host)
	}
	if config.MQTT.Port != 1883 {
		t.Errorf("Expected default port 1883, got %d", config.MQTT.Port)
	}
	if config.Prometheus.Enabled != true {
		t.Error("Expected prometheus to be enabled by default")
	}
}

func TestStandaloneExporterCreation(t *testing.T) {
	config := exporter.Config{}
	config.MQTT.Host = "localhost"
	config.MQTT.Port = 1883
	config.Prometheus.Enabled = false

	exp := exporter.New(config)
	if exp == nil {
		t.Fatal("Expected exporter to be created")
	}

	exp.Init()
}
