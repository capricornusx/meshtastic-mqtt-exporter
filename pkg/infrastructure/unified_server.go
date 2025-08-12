package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/logger"
	"meshtastic-exporter/pkg/middleware"
)

type UnifiedServerConfig struct {
	Addr         string
	EnableHealth bool
	AlertPath    string
}

type UnifiedServer struct {
	config    UnifiedServerConfig
	collector domain.MetricsCollector
	alerter   domain.AlertSender
	server    *http.Server
	logger    zerolog.Logger
	mu        sync.RWMutex
}

type AlertPayload struct {
	Alerts []AlertItem `json:"alerts"`
}

type AlertItem struct {
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

func NewUnifiedServer(config UnifiedServerConfig, collector domain.MetricsCollector, alerter domain.AlertSender) *UnifiedServer {
	if config.AlertPath == "" {
		config.AlertPath = domain.DefaultAlertsPath
	}

	return &UnifiedServer{
		config:    config,
		collector: collector,
		alerter:   alerter,
		logger:    logger.ComponentLogger("unified-server"),
	}
}

func (s *UnifiedServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	if s.collector != nil {
		mux.Handle(domain.DefaultMetricsPath, promhttp.HandlerFor(s.collector.GetRegistry(), promhttp.HandlerOpts{}))
	}

	if s.config.EnableHealth {
		mux.HandleFunc(domain.DefaultHealthPath, s.healthHandler)
	}

	if s.alerter != nil {
		mux.HandleFunc(s.config.AlertPath, s.alertWebhookHandler)
	}

	handler := middleware.ChainMiddleware(
		middleware.RecoveryMiddleware(s.logger),
		middleware.TimeoutMiddleware(domain.DefaultTimeout),
	)(mux)

	server := &http.Server{
		Addr:              s.config.Addr,
		Handler:           handler,
		ReadTimeout:       domain.DefaultReadTimeout,
		WriteTimeout:      domain.DefaultWriteTimeout,
		ReadHeaderTimeout: domain.DefaultHeaderTimeout,
		IdleTimeout:       domain.DefaultIdleTimeout,
	}

	s.mu.Lock()
	s.server = server
	s.mu.Unlock()

	go func() {
		s.logger.Info().Str("address", s.config.Addr).Msg("unified server starting")

		listener, err := net.Listen("tcp", s.config.Addr)
		if err != nil {
			s.logger.Error().Err(err).Msg("failed to create listener")
			return
		}

		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error().Err(err).Msg("unified server error")
		}
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), domain.DefaultTimeout)
		defer cancel()
		if err := s.Shutdown(shutdownCtx); err != nil {
			s.logger.Error().Err(err).Msg("server shutdown error")
		}
	}()

	return nil
}

func (s *UnifiedServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"service":"meshtastic-exporter","status":"ok","timestamp":%d}`, time.Now().Unix())
}

func (s *UnifiedServer) alertWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload AlertPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), domain.DefaultTimeout)
	defer cancel()

	for _, alertItem := range payload.Alerts {
		alert := s.convertToAlert(alertItem)
		if err := s.alerter.SendAlert(ctx, alert); err != nil {
			s.logger.Error().Err(err).Msg("failed to send alert")
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *UnifiedServer) convertToAlert(item AlertItem) domain.Alert {
	emoji := "🚨"
	if item.Status == "resolved" {
		emoji = "✅"
	}

	msg := fmt.Sprintf("%s %s: %s", emoji, item.Status, item.Labels["alertname"])
	if summary := item.Annotations["summary"]; summary != "" && len(msg)+len(summary) < 200 {
		msg += " - " + summary
	}

	return domain.Alert{
		Severity:  item.Labels["severity"],
		Message:   msg,
		Timestamp: time.Now(),
	}
}

func (s *UnifiedServer) Shutdown(ctx context.Context) error {
	s.mu.RLock()
	server := s.server
	s.mu.RUnlock()

	if server != nil {
		return server.Shutdown(ctx)
	}
	return nil
}
