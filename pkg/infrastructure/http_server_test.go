package infrastructure

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/mocks"
)

func TestNewHTTPServer(t *testing.T) {
	t.Parallel()
	collector := &mocks.MockMetricsCollector{}
	alerter := &mocks.MockAlertSender{}

	server := NewHTTPServer(":8080", collector, alerter)
	assert.NotNil(t, server)
	assert.NotNil(t, server.unified)
}

func TestHTTPServer_Start(t *testing.T) {
	t.Parallel()
	collector := &mocks.MockMetricsCollector{}
	alerter := &mocks.MockAlertSender{}

	server := NewHTTPServer(":0", collector, alerter)
	ctx := context.Background()

	err := server.Start(ctx)
	require.NoError(t, err)

	defer server.Shutdown(ctx)
	time.Sleep(10 * time.Millisecond)
}

func TestHTTPServer_Shutdown(t *testing.T) {
	t.Parallel()
	collector := &mocks.MockMetricsCollector{}
	alerter := &mocks.MockAlertSender{}

	server := NewHTTPServer(":0", collector, alerter)
	ctx := context.Background()

	server.Start(ctx)
	time.Sleep(10 * time.Millisecond)

	err := server.Shutdown(ctx)
	require.NoError(t, err)
}

func TestHTTPServer_ShutdownWithoutStart(t *testing.T) {
	t.Parallel()
	collector := &mocks.MockMetricsCollector{}
	alerter := &mocks.MockAlertSender{}

	server := NewHTTPServer(":0", collector, alerter)
	ctx := context.Background()

	err := server.Shutdown(ctx)
	require.NoError(t, err)
}
