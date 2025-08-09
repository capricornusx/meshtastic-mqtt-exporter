package main

import (
	"os"
	"testing"

	"meshtastic-exporter/pkg/exporter"
)

func TestConfigLoading(t *testing.T) {
	// Create a temporary config file
	configContent := `
mqtt:
  host: localhost
  port: 1883
  allow_anonymous: true
prometheus:
  enabled: false
`
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Test config loading
	config, err := exporter.LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.MQTT.Host != "localhost" {
		t.Errorf("Expected host localhost, got %s", config.MQTT.Host)
	}
	if config.MQTT.Port != 1883 {
		t.Errorf("Expected port 1883, got %d", config.MQTT.Port)
	}
	if !config.MQTT.AllowAnonymous {
		t.Error("Expected allow_anonymous to be true")
	}
}

func TestConfigDefaults(t *testing.T) {
	// Test with non-existent config file
	config, err := exporter.LoadConfig("non-existent-file.yaml")
	if err != nil {
		t.Fatalf("Expected no error with missing config, got: %v", err)
	}

	// Check defaults
	if config.MQTT.Host != "localhost" {
		t.Errorf("Expected default host localhost, got %s", config.MQTT.Host)
	}
	if config.MQTT.Port != 1883 {
		t.Errorf("Expected default port 1883, got %d", config.MQTT.Port)
	}
	if config.Prometheus.Port != 8000 {
		t.Errorf("Expected default prometheus port 8000, got %d", config.Prometheus.Port)
	}
}

func TestConfigLoadingError(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// Write invalid YAML
	if _, err := tmpFile.WriteString("invalid: yaml: content: ["); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	_, err = exporter.LoadConfig(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestConfigWithUsers(t *testing.T) {
	configContent := `
mqtt:
  host: test.local
  port: 1884
  users:
    - username: user1
      password: pass1
    - username: user2
      password: pass2
prometheus:
  enabled: true
  port: 8100
state:
  enabled: true
  file: test_state.json
`
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
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

	if len(config.MQTT.Users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(config.MQTT.Users))
	}
	if config.MQTT.Users[0].Username != "user1" {
		t.Errorf("Expected username user1, got %s", config.MQTT.Users[0].Username)
	}
	if config.State.Enabled != true {
		t.Error("Expected state to be enabled")
	}
	if config.State.File != "test_state.json" {
		t.Errorf("Expected state file test_state.json, got %s", config.State.File)
	}
}
