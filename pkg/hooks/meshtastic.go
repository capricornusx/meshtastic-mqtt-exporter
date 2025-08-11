package hooks

import (
	"context"
	"fmt"
	"strings"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/rs/zerolog"

	"meshtastic-exporter/pkg/domain"
	apperrors "meshtastic-exporter/pkg/errors"
	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/infrastructure"
	"meshtastic-exporter/pkg/logger"
)

type MeshtasticHookConfig struct {
	ServerAddr   string
	EnableHealth bool
	TopicPrefix  string
	MetricsTTL   time.Duration
	AlertPath    string
}

type MeshtasticHook struct {
	mqtt.HookBase

	processor domain.MessageProcessor
	collector domain.MetricsCollector
	alerter   domain.AlertSender

	config   MeshtasticHookConfig
	logger   zerolog.Logger
	server   *infrastructure.UnifiedServer
	factory  *factory.Factory
	stopSave chan struct{}
}

func NewMeshtasticHook(cfg MeshtasticHookConfig, f *factory.Factory) *MeshtasticHook {
	if cfg.TopicPrefix == "" {
		cfg.TopicPrefix = domain.DefaultTopicPrefix
	}
	if cfg.MetricsTTL == 0 {
		cfg.MetricsTTL = domain.DefaultMetricsTTL
	}
	if cfg.AlertPath == "" {
		cfg.AlertPath = domain.DefaultAlertsPath
	}

	if f == nil {
		return nil
	}

	collector := f.CreateMetricsCollector()
	alerter := f.CreateAlertSender()
	processor := f.CreateMessageProcessor()

	return &MeshtasticHook{
		processor: processor,
		collector: collector,
		alerter:   alerter,
		config:    cfg,
		logger:    logger.ComponentLogger("meshtastic-hook"),
		factory:   f,
		stopSave:  make(chan struct{}),
	}
}

func NewMeshtasticHookSimple(f *factory.Factory) *MeshtasticHook {
	config := MeshtasticHookConfig{
		ServerAddr:   fmt.Sprintf("%s:%d", domain.DefaultPrometheusHost, domain.DefaultPrometheusPort),
		EnableHealth: true,
		TopicPrefix:  domain.DefaultTopicPrefix,
		AlertPath:    domain.DefaultAlertsPath,
	}

	if f != nil {
		if promConfig := f.GetPrometheusConfig(); promConfig != nil {
			config.ServerAddr = promConfig.GetListen()
		}
		if alertConfig := f.GetAlertManagerConfig(); alertConfig != nil {
			config.AlertPath = alertConfig.GetPath()
		}
		return NewMeshtasticHook(config, f)
	}

	return NewMeshtasticHook(config, nil)
}

func (h *MeshtasticHook) ID() string {
	return "meshtastic"
}

func (h *MeshtasticHook) Provides(b byte) bool {
	return b == mqtt.OnPublish || b == mqtt.OnConnect || b == mqtt.OnDisconnect
}

func (h *MeshtasticHook) Init(config any) error {
	if h.config.ServerAddr != "" {
		h.startUnifiedServer()
	}
	h.startStateSaver()
	return nil
}

func (h *MeshtasticHook) OnConnect(cl *mqtt.Client, pk packets.Packet) error {
	h.logger.Debug().Str("client_id", cl.ID).Msg("client connected")
	return nil
}

func (h *MeshtasticHook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	h.logger.Debug().Str("client_id", cl.ID).Bool("expire", expire).Msg("client disconnected")
}

func (h *MeshtasticHook) OnPublish(_ *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	if !strings.HasPrefix(pk.TopicName, h.config.TopicPrefix) {
		return pk, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), domain.DefaultTimeout)
	defer cancel()

	if err := h.processor.ProcessMessage(ctx, pk.TopicName, pk.Payload); err != nil {
		appErr := apperrors.NewProcessingError("message processing failed", err)
		h.logger.Error().Err(appErr).Str("topic", pk.TopicName).Msg("message processing failed")
	}

	return pk, nil
}

func (h *MeshtasticHook) startUnifiedServer() {
	serverConfig := infrastructure.UnifiedServerConfig{
		Addr:         h.config.ServerAddr,
		EnableHealth: h.config.EnableHealth,
		AlertPath:    h.config.AlertPath,
	}

	h.server = infrastructure.NewUnifiedServer(serverConfig, h.collector, h.alerter)
	if err := h.server.Start(context.Background()); err != nil {
		h.logger.Error().Err(err).Msg("failed to start unified server")
	}
}

func (h *MeshtasticHook) startStateSaver() {
	if h.factory == nil {
		return
	}

	prometheusConfig := h.factory.GetPrometheusConfig()
	if prometheusConfig == nil || prometheusConfig.GetStateFile() == "" {
		return
	}

	go func() {
		ticker := time.NewTicker(domain.DefaultStateSaveInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := h.collector.SaveState(prometheusConfig.GetStateFile()); err != nil {
					h.logger.Error().Err(err).Msg("failed to save metrics state")
				} else {
					h.logger.Debug().Msg("metrics state saved")
				}
			case <-h.stopSave:
				return
			}
		}
	}()
}

func (h *MeshtasticHook) Shutdown(ctx context.Context) error {
	close(h.stopSave)

	if h.factory != nil {
		prometheusConfig := h.factory.GetPrometheusConfig()
		if prometheusConfig != nil && prometheusConfig.GetStateFile() != "" {
			if err := h.collector.SaveState(prometheusConfig.GetStateFile()); err != nil {
				h.logger.Error().Err(err).Msg("failed to save final state")
			}
		}
	}

	if h.server != nil {
		return h.server.Shutdown(ctx)
	}
	return nil
}
