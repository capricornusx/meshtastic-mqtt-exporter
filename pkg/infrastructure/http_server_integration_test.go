package infrastructure

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"meshtastic-exporter/pkg/mocks"
)

func TestHTTPServer_StartAndShutdown_Integration(t *testing.T) {
	collector := &mocks.MockMetricsCollector{}
	alerter := &mocks.MockAlertSender{}
	server := NewHTTPServer(":0", collector, alerter)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- server.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		t.Errorf("Server shutdown error: %v", err)
	}

	select {
	case err := <-done:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Errorf("Server start error: %v", err)
		}
	case <-time.After(time.Second):
		t.Error("Server did not stop in time")
	}
}
