package tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"
	"meshtastic-exporter/pkg/infrastructure"
)

func TestAlertingIntegration(t *testing.T) {
	// Create MQTT server
	server := mqtt.New(&mqtt.Options{InlineClient: true})
	require.NoError(t, server.AddHook(new(auth.AllowHook), nil))

	tcp := listeners.NewTCP(listeners.Config{ID: "tcp", Address: ":0"})
	require.NoError(t, server.AddListener(tcp))

	go server.Serve()
	defer server.Close()

	// Create LoRa AlertSender
	alerter := infrastructure.NewLoRaAlertSender(server, infrastructure.LoRaConfig{
		DefaultChannel: "LongFast",
		DefaultMode:    "broadcast",
	})

	// Create AlertManager hook
	alertHook := hooks.NewAlertmanagerHook(alerter, hooks.AlertManagerConfig{
		HTTPHost: "localhost",
		HTTPPort: 0, // Random port
		HTTPPath: "/webhook",
	})

	require.NoError(t, alertHook.Init(nil))
	defer alertHook.Shutdown(context.Background())

	// Test alert sending
	alert := domain.Alert{
		Severity:  "critical",
		Message:   "Test alert message",
		Channel:   "LongFast",
		Mode:      "broadcast",
		Timestamp: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := alerter.SendAlert(ctx, alert)
	assert.NoError(t, err)
}

func TestMQTTMessageFlow(t *testing.T) {
	server := mqtt.New(&mqtt.Options{InlineClient: true})
	require.NoError(t, server.AddHook(new(auth.AllowHook), nil))

	// Create hook
	f := factory.NewDefaultFactory()
	meshtasticHook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
		ServerAddr:   ":0",
		TopicPrefix:  "msh/",
		EnableHealth: true,
	}, f)

	require.NoError(t, server.AddHook(meshtasticHook, nil))

	tcp := listeners.NewTCP(listeners.Config{ID: "tcp", Address: ":0"})
	require.NoError(t, server.AddListener(tcp))

	go server.Serve()
	defer server.Close()
	defer meshtasticHook.Shutdown(context.Background())

	// Test message processing
	testMessage := map[string]interface{}{
		"from": 123456789,
		"type": "telemetry",
		"payload": map[string]interface{}{
			"battery_level": 85.5,
			"temperature":   23.4,
		},
	}

	data, _ := json.Marshal(testMessage)

	// Publish message
	server.Publish("msh/2/c/LongFast/!123456789", data, false, 0)

	// Give time for processing
	time.Sleep(100 * time.Millisecond)
}

func TestHealthEndpointIntegration(t *testing.T) {
	f := factory.NewDefaultFactory()
	hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
		ServerAddr:   ":0",
		EnableHealth: true,
	}, f)

	require.NoError(t, hook.Init(nil))
	defer hook.Shutdown(context.Background())

	// Wait for server startup
	time.Sleep(100 * time.Millisecond)

	// Test health endpoint (real test needs actual port)
	// Here we show test structure
	t.Log("Health endpoint integration test structure ready")
}
