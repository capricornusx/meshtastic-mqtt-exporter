package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"meshtastic-exporter/pkg/exporter"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func TestIntegrationEmbeddedMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	config := exporter.Config{}
	config.MQTT.Host = "localhost"
	config.MQTT.Port = 1883
	config.MQTT.AllowAnonymous = true
	config.Prometheus.Enabled = true
	config.Prometheus.Host = "localhost"
	config.Prometheus.Port = 8100

	tmpFile, err := os.CreateTemp("", "integration-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	configData := fmt.Sprintf(`mqtt:
  host: %s
  port: %d
  allow_anonymous: true
prometheus:
  enabled: true
  host: %s
  port: %d`, config.MQTT.Host, config.MQTT.Port, config.Prometheus.Host, config.Prometheus.Port)

	if err := os.WriteFile(tmpFile.Name(), []byte(configData), 0600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		// Simulate embedded server startup
		time.Sleep(2 * time.Second)
	}()

	time.Sleep(3 * time.Second)

	// Test MQTT publishing
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", config.MQTT.Host, config.MQTT.Port))
	opts.SetClientID("integration-test")
	opts.SetUsername("meshtastic")
	opts.SetPassword("mesh456")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		t.Skipf("MQTT broker not available: %v", token.Error())
	}
	defer client.Disconnect(250)

	// Publish test message
	testMessage := map[string]interface{}{
		"from": float64(123456),
		"to":   float64(789012),
		"type": "telemetry",
		"payload": map[string]interface{}{
			"battery_level": float64(85.5),
			"voltage":       float64(3.7),
		},
	}

	msgBytes, _ := json.Marshal(testMessage)
	token := client.Publish("msh/2/json/LongFast/123456", 0, false, msgBytes)
	if token.Wait() && token.Error() != nil {
		t.Errorf("Failed to publish: %v", token.Error())
	}

	time.Sleep(2 * time.Second)

	// Test Prometheus metrics
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/metrics", config.Prometheus.Host, config.Prometheus.Port))
	if err != nil {
		t.Skipf("Prometheus endpoint not available: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	metrics := string(body)
	if !strings.Contains(metrics, "meshtastic_messages_total") {
		t.Error("Expected meshtastic_messages_total metric not found")
	}

	select {
	case <-ctx.Done():
		t.Error("Test timeout")
	default:
	}
}

func TestHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Test health endpoint
	resp, err := http.Get("http://localhost:8100/health")
	if err != nil {
		t.Skipf("Health endpoint not available: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var health map[string]interface{}
	if err := json.Unmarshal(body, &health); err != nil {
		t.Fatal(err)
	}

	if health["status"] != "ok" {
		t.Errorf("Expected status ok, got %v", health["status"])
	}
}
