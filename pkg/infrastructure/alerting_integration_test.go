package infrastructure

import (
	"context"
	"testing"
	"time"

	"meshtastic-exporter/pkg/domain"

	mqtt "github.com/mochi-mqtt/server/v2"
)

func TestLoRaAlertSender_SendAlert_Integration(t *testing.T) {
	opts := &mqtt.Options{
		InlineClient: true,
	}
	server := mqtt.New(opts)
	config := LoRaConfig{DefaultChannel: "LongFast", DefaultMode: "broadcast"}
	sender := NewLoRaAlertSender(server, config)

	alert := domain.Alert{
		Severity:  "critical",
		Message:   "Test alert",
		Channel:   "LongFast",
		Mode:      "broadcast",
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	err := sender.SendAlert(ctx, alert)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}
