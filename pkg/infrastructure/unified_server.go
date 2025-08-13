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

	"meshtastic-exporter/pkg/adapters"
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
	config      UnifiedServerConfig
	collector   domain.MetricsCollector
	alerter     domain.AlertSender
	alertConfig domain.AlertManagerConfig
	server      *http.Server
	logger      zerolog.Logger
	mu          sync.RWMutex
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

func NewUnifiedServerWithAlertConfig(config UnifiedServerConfig, collector domain.MetricsCollector, alerter domain.AlertSender, alertConfig domain.AlertManagerConfig) *UnifiedServer {
	if config.AlertPath == "" {
		config.AlertPath = domain.DefaultAlertsPath
	}

	return &UnifiedServer{
		config:      config,
		collector:   collector,
		alerter:     alerter,
		alertConfig: alertConfig,
		logger:      logger.ComponentLogger("unified-server"),
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

	alert := domain.Alert{
		Severity:  item.Labels["severity"],
		Message:   msg,
		Timestamp: time.Now(),
	}

	// Применяем routing настройки
	if s.alertConfig != nil {
		s.applyRoutingConfig(&alert)
	}

	return alert
}

func (s *UnifiedServer) applyRoutingConfig(alert *domain.Alert) {
	routing := s.alertConfig.GetRouting()
	if routing == nil {
		return
	}

	// Приводим к типу AlertRoutingConfig
	if routingConfig, ok := routing.(adapters.AlertRoutingConfig); ok {
		var route *adapters.AlertRouteConfig

		// Выбираем маршрут по severity
		switch alert.Severity {
		case "critical":
			route = routingConfig.Critical
		case "warning":
			route = routingConfig.Warning
		case "info":
			route = routingConfig.Info
		default:
			route = routingConfig.Default
		}

		// Если маршрут не найден, используем default
		if route == nil {
			route = routingConfig.Default
		}

		// Применяем настройки маршрута
		if route != nil {
			alert.Mode = route.Mode
			alert.TargetNodes = route.TargetNodes

			// Для broadcast с show_on_sender добавляем отправителя в targets
			if route.Mode == "broadcast" && route.ShowOnSender {
				if fromNodeID := s.alertConfig.GetFromNodeID(); fromNodeID != 0 {
					alert.TargetNodes = append(alert.TargetNodes, fromNodeID)
					alert.Mode = "mixed" // Особый режим: broadcast + direct на отправителя
				}
			}

			s.logger.Debug().Str("severity", alert.Severity).Str("mode", alert.Mode).Ints("targets", intSliceFromUint32(route.TargetNodes)).Bool("show_on_sender", route.ShowOnSender).Msg("applied routing config")
		}
	}
}

func intSliceFromUint32(nodes []uint32) []int {
	result := make([]int, len(nodes))
	for i, node := range nodes {
		result[i] = int(node)
	}
	return result
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
