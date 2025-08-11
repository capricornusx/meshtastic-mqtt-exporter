package standalone

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/errors"
	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/infrastructure"
	"meshtastic-exporter/pkg/logger"
)

type App struct {
	config     domain.Config
	processor  domain.MessageProcessor
	collector  domain.MetricsCollector
	alerter    domain.AlertSender
	mqttClient *infrastructure.MQTTClient
	httpServer *infrastructure.HTTPServer
	logger     zerolog.Logger
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewApp(config domain.Config) *App {
	ctx, cancel := context.WithCancel(context.Background())

	f := factory.NewFactory(config)
	collector := f.CreateMetricsCollectorWithMode("standalone")
	alerter := f.CreateAlertSender()
	processor := f.CreateMessageProcessor()

	return &App{
		config:    config,
		processor: processor,
		collector: collector,
		alerter:   alerter,
		logger:    logger.ComponentLogger("standalone-app"),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (a *App) Run() error {
	if err := a.config.Validate(); err != nil {
		return errors.NewConfigError("invalid configuration", err)
	}

	// Start MQTT client
	mqttConfig := a.config.GetMQTTConfig()
	a.mqttClient = infrastructure.NewMQTTClient(mqttConfig, a.processor)
	if err := a.mqttClient.Connect(); err != nil {
		return errors.NewNetworkError("failed to connect to mqtt", err)
	}

	// Start HTTP server for metrics
	prometheusConfig := a.config.GetPrometheusConfig()
	addr := prometheusConfig.GetListen()

	a.httpServer = infrastructure.NewHTTPServer(addr, a.collector, a.alerter)
	if err := a.httpServer.Start(a.ctx); err != nil {
		return errors.NewNetworkError("failed to start http server", err)
	}

	a.logger.Info().Str("address", addr).Msg("http server started")

	a.logger.Info().Msg("standalone application started")

	// Wait for a shutdown signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	return a.Shutdown()
}

func (a *App) Shutdown() error {
	a.logger.Info().Msg("shutting down")

	a.cancel()

	ctx, cancel := context.WithTimeout(context.Background(), domain.DefaultTimeout/domain.ShutdownTimeoutDivider)
	defer cancel()

	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			a.logger.Error().Err(err).Msg("http server shutdown error")
		}
	}

	if a.mqttClient != nil {
		a.mqttClient.Disconnect()
	}

	a.logger.Info().Msg("shutdown completed")
	return nil
}
