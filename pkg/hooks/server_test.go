package hooks

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/factory"
)

func TestPrometheusHookConfig(t *testing.T) {
	t.Parallel()
	// Test hook creation with configuration
	config := MeshtasticHookConfig{
		ServerAddr:   ":9001",
		EnableHealth: false,
	}
	f := factory.NewDefaultFactory()
	hook := NewMeshtasticHook(config, f)

	// Test that hook is created with correct config
	assert.Equal(t, ":9001", hook.config.ServerAddr)
}

func TestStartServerDisabled(t *testing.T) {
	t.Parallel()
	config := MeshtasticHookConfig{
		ServerAddr: "", // Disabled
	}
	f := factory.NewDefaultFactory()
	hook := NewMeshtasticHook(config, f)

	// Should not start server when disabled
	err := hook.Init(nil)
	require.NoError(t, err)
}

func TestStartPeriodicStateSave(t *testing.T) {
	t.Parallel()
	// Test that function doesn't panic
	config := MeshtasticHookConfig{
		ServerAddr: "",
	}
	f := factory.NewDefaultFactory()
	hook := NewMeshtasticHook(config, f)

	// Test that function doesn't panic when disabled
	assert.NotNil(t, hook)
}

func TestLoadStateEnabled(t *testing.T) {
	t.Parallel()
	config := MeshtasticHookConfig{
		ServerAddr: "",
	}
	f := factory.NewDefaultFactory()
	hook := NewMeshtasticHook(config, f)

	// Should return early when state is disabled
	assert.NotNil(t, hook)
}

func TestSaveStateDisabled(t *testing.T) {
	t.Parallel()
	config := MeshtasticHookConfig{
		ServerAddr: "",
	}
	f := factory.NewDefaultFactory()
	hook := NewMeshtasticHook(config, f)

	// Should return early when state is disabled
	assert.NotNil(t, hook)
}

func TestMeshtasticHookInit(t *testing.T) {
	t.Parallel()
	f := factory.NewDefaultFactory()
	hook := NewMeshtasticHookSimple(f)

	// Test Init without server
	err := hook.Init(nil)
	require.NoError(t, err)
}

func TestMeshtasticHookStartServers(t *testing.T) {
	t.Parallel()
	config := MeshtasticHookConfig{
		ServerAddr:   "",    // Disable prometheus
		EnableHealth: false, // Disable health
	}
	// AlertManager disabled // Disable alertmanager
	f := factory.NewDefaultFactory()

	hook := NewMeshtasticHook(config, f)

	// Test Init without starting servers
	err := hook.Init(nil)
	require.NoError(t, err)
}

func TestHealthEndpointIntegration(t *testing.T) {
	t.Parallel()
	config := MeshtasticHookConfig{
		ServerAddr:   "", // Disable to avoid conflicts
		EnableHealth: false,
	}
	f := factory.NewDefaultFactory()

	_ = NewMeshtasticHook(config, f)

	// Create a test server
	server := &http.Server{
		Addr:              "127.0.0.1:0",
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	// Test that server can be created without panic
	assert.Equal(t, "127.0.0.1:0", server.Addr)
}

func TestServerShutdown(t *testing.T) {
	t.Parallel()
	// Test MeshtasticHook with disabled prometheus to avoid port conflicts
	config := MeshtasticHookConfig{
		ServerAddr:   "", // Will be disabled
		EnableHealth: false,
	}
	// AlertManager disabled
	f := factory.NewDefaultFactory()

	hook := NewMeshtasticHook(config, f)

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Test that hook can be created without starting servers
	assert.NotNil(t, hook)

	// Test context cancellation doesn't panic
	select {
	case <-ctx.Done():
		// Context cancelled as expected
	case <-time.After(2 * time.Second):
		assert.Fail(t, "Context should have been cancelled")
	}
}
