package hooks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/mocks"
)

func TestNewAlertmanagerHook(t *testing.T) {
	mockAlerter := &mocks.MockAlertSender{}
	hook := NewAlertmanagerHook(mockAlerter, AlertManagerConfig{
		HTTPHost: "localhost",
		HTTPPort: 8080,
		HTTPPath: "/webhook",
	})

	assert.Equal(t, "alertmanager-lora", hook.ID())
	assert.Equal(t, "localhost", hook.config.HTTPHost)
	assert.Equal(t, 8080, hook.config.HTTPPort)
	assert.Equal(t, "/webhook", hook.config.HTTPPath)
	assert.False(t, hook.Provides(0)) // HTTP-only hook
}

func TestAlertmanagerDefaults(t *testing.T) {
	mockAlerter := &mocks.MockAlertSender{}
	hook := NewAlertmanagerHook(mockAlerter, AlertManagerConfig{})

	assert.Equal(t, "localhost", hook.config.HTTPHost)
	assert.Equal(t, "/alerts/webhook", hook.config.HTTPPath)
	assert.Equal(t, 8080, hook.config.HTTPPort)
}

func TestAlertmanagerHook_Init(t *testing.T) {
	mockAlerter := &mocks.MockAlertSender{}
	hook := NewAlertmanagerHook(mockAlerter, AlertManagerConfig{
		HTTPPort: 0, // Use random port for test
	})

	err := hook.Init(nil)
	require.NoError(t, err)

	// Cleanup
	ctx := context.Background()
	err = hook.Shutdown(ctx)
	require.NoError(t, err)
}

func TestAlertmanagerHook_Shutdown(t *testing.T) {
	mockAlerter := &mocks.MockAlertSender{}
	hook := NewAlertmanagerHook(mockAlerter, AlertManagerConfig{})

	err := hook.Shutdown(context.Background())
	require.NoError(t, err)
}
