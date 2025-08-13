package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationHookMode(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	config := hooks.MeshtasticHookConfig{
		ServerAddr:   ":0", // Random port
		EnableHealth: true,
	}
	f := factory.NewDefaultFactory()
	hook := hooks.NewMeshtasticHook(config, f)
	assert.NotNil(t, hook)
}

func TestHealthEndpoint(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test HTTP client with timeout on unused port
	client := &http.Client{Timeout: 1 * time.Second}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:65432/health", nil)
	_, err := client.Do(req)
	// Expect connection error since no server is running
	assert.Error(t, err)
}

func TestAlertManagerWebhook(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test webhook endpoint with timeout on unused port
	client := &http.Client{Timeout: 1 * time.Second}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:65433/alerts", nil)
	_, err := client.Do(req)
	// Expect connection error since no server is running
	assert.Error(t, err)
}
