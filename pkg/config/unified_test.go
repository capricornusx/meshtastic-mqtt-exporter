package config

import (
	"os"
	"testing"
	"time"

	"fmt"

	"meshtastic-exporter/pkg/domain"
)

// Используем константы из domain пакета вместо дублирования

func TestLoadUnifiedConfig_Success(t *testing.T) {
	t.Parallel()
	configContent := `
logging:
  level: "debug"

mqtt:
  host: "test.example.com"
  port: 1884
  tls: true
  allow_anonymous: false
  username: "testuser"
  password: "testpass"
  users:
    - username: "user1"
      password: "pass1"

hook:
  listen: "0.0.0.0:8100"
  prometheus:
    path: "/metrics"
    metrics_ttl: "1h"
    topic:
      pattern: "test/#"
  alertmanager:
    path: "/test/webhook"
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	config, err := LoadUnifiedConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	mqttConfig := config.GetMQTTConfig()
	if mqttConfig.GetHost() != "test.example.com" {
		t.Errorf("Expected host 'test.example.com', got '%s'", mqttConfig.GetHost())
	}
	if mqttConfig.GetPort() != 1884 {
		t.Errorf("Expected port 1884, got %d", mqttConfig.GetPort())
	}
}

func TestLoadUnifiedConfig_FileNotExists(t *testing.T) {
	t.Parallel()
	config, err := LoadUnifiedConfig("nonexistent.yaml")
	if err != nil {
		t.Fatalf("Expected no error for missing file, got %v", err)
	}

	mqttConfig := config.GetMQTTConfig()
	if mqttConfig.GetHost() != domain.DefaultPrometheusHost {
		t.Errorf("Expected default host '%s', got '%s'", domain.DefaultPrometheusHost, mqttConfig.GetHost())
	}
}

func TestLoadUnifiedConfig_InvalidYAML(t *testing.T) {
	t.Parallel()
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString("invalid: yaml: content:"); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	_, err = LoadUnifiedConfig(tmpFile.Name())
	if err == nil {
		t.Fatal("Expected error for invalid YAML")
	}
}

func TestSetDefaults(t *testing.T) {
	t.Parallel()
	config := &UnifiedConfig{}
	setDefaults(config)

	if config.MQTT.Host != domain.DefaultPrometheusHost {
		t.Errorf("Expected default MQTT host '%s', got '%s'", domain.DefaultPrometheusHost, config.MQTT.Host)
	}
	if config.MQTT.Port != 1883 {
		t.Errorf("Expected default MQTT port 1883, got %d", config.MQTT.Port)
	}

	expectedListen := fmt.Sprintf("%s:%d", domain.DefaultPrometheusHost, domain.DefaultPrometheusPort)
	if config.Hook.Listen != expectedListen {
		t.Errorf("Expected default listen '%s', got '%s'", expectedListen, config.Hook.Listen)
	}
	if config.Hook.Prometheus.Topic.Pattern != domain.DefaultTopicPrefix {
		t.Errorf("Expected default topic pattern '%s', got '%s'", domain.DefaultTopicPrefix, config.Hook.Prometheus.Topic.Pattern)
	}
	if config.Logging.Level != "info" {
		t.Errorf("Expected default logging level 'info', got '%s'", config.Logging.Level)
	}
	if config.Hook.Prometheus.Topic.LogAllMessages != false {
		t.Errorf("Expected default log_all_messages 'false', got '%v'", config.Hook.Prometheus.Topic.LogAllMessages)
	}
}

func TestConvertToAdapter_InvalidMetricsTTL(t *testing.T) {
	t.Parallel()
	config := &UnifiedConfig{}
	setDefaults(config)
	config.Hook.Prometheus.MetricsTTL = "invalid-duration"

	adapter, err := convertToAdapter(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	prometheusConfig := adapter.GetPrometheusConfig()
	if prometheusConfig.GetMetricsTTL() != domain.DefaultMetricsTTL {
		t.Errorf("Expected default TTL %v, got %v", domain.DefaultMetricsTTL, prometheusConfig.GetMetricsTTL())
	}
}

func TestConvertToAdapter_WithUsers(t *testing.T) {
	t.Parallel()
	config := &UnifiedConfig{}
	setDefaults(config)
	config.MQTT.Username = "mainuser"
	config.MQTT.Password = "mainpass"
	config.MQTT.Users = []struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	}{
		{Username: "user1", Password: "pass1"},
		{Username: "user2", Password: "pass2"},
	}

	adapter, err := convertToAdapter(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	mqttConfig := adapter.GetMQTTConfig()
	users := mqttConfig.GetUsers()
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}
}

func TestConvertToAdapter_ValidMetricsTTL(t *testing.T) {
	t.Parallel()
	config := &UnifiedConfig{}
	setDefaults(config)
	config.Hook.Prometheus.MetricsTTL = "2h"

	adapter, err := convertToAdapter(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	prometheusConfig := adapter.GetPrometheusConfig()
	expected := 2 * time.Hour
	if prometheusConfig.GetMetricsTTL() != expected {
		t.Errorf("Expected TTL %v, got %v", expected, prometheusConfig.GetMetricsTTL())
	}
}

func TestConvertToAdapter_LoggingLevel(t *testing.T) {
	t.Parallel()
	config := &UnifiedConfig{}
	setDefaults(config)
	config.Logging.Level = "debug"

	adapter, err := convertToAdapter(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if adapter == nil {
		t.Fatal("Expected adapter to be created")
	}
}

func TestConvertToAdapter_LogAllMessages(t *testing.T) {
	t.Parallel()
	config := &UnifiedConfig{}
	setDefaults(config)
	config.Hook.Prometheus.Topic.LogAllMessages = true

	adapter, err := convertToAdapter(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	prometheusConfig := adapter.GetPrometheusConfig()
	if !prometheusConfig.GetLogAllMessages() {
		t.Error("Expected log_all_messages to be true")
	}
}
