package infrastructure

import (
	"context"

	"meshtastic-exporter/pkg/domain"
)

type HTTPServer struct {
	unified *UnifiedServer
}

func NewHTTPServer(addr string, collector domain.MetricsCollector, alerter domain.AlertSender) *HTTPServer {
	config := UnifiedServerConfig{
		Addr:         addr,
		EnableHealth: true,
		AlertPath:    domain.DefaultAlertsPath,
	}

	return &HTTPServer{
		unified: NewUnifiedServer(config, collector, alerter),
	}
}

func (s *HTTPServer) Start(ctx context.Context) error {
	return s.unified.Start(ctx)
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.unified.Shutdown(ctx)
}
