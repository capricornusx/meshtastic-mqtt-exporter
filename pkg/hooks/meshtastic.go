package hooks

import (
	"context"
	"fmt"
	"os"
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
	stopped  bool
}

func NewMeshtasticHook(cfg MeshtasticHookConfig, f *factory.Factory) *MeshtasticHook {
	if cfg.TopicPrefix == "" {
		cfg.TopicPrefix = domain.DefaultTopicPrefix
	}
	// Убеждаемся что префикс заканчивается на /
	if !strings.HasSuffix(cfg.TopicPrefix, "/") {
		cfg.TopicPrefix += "/"
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

	collector := f.CreateMetricsCollectorWithMode("embedded")
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

func NewMeshtasticHookWithMQTT(cfg MeshtasticHookConfig, f *factory.Factory, mqttServer *mqtt.Server) *MeshtasticHook {
	if cfg.TopicPrefix == "" {
		cfg.TopicPrefix = domain.DefaultTopicPrefix
	}
	if !strings.HasSuffix(cfg.TopicPrefix, "/") {
		cfg.TopicPrefix += "/"
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

	collector := f.CreateMetricsCollectorWithMode("embedded")
	alerter := f.CreateAlertSenderWithMQTT(mqttServer)
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

func (h *MeshtasticHook) ID() string {
	return "meshtastic"
}

func (h *MeshtasticHook) Provides(b byte) bool {
	return b == mqtt.OnPublish || b == mqtt.OnConnect || b == mqtt.OnDisconnect || b == mqtt.OnStopped
}

func (h *MeshtasticHook) Init(config any) error {
	if h.config.ServerAddr != "" {
		h.startUnifiedServer()
	}
	if err := h.validateStateFile(); err != nil {
		return err
	}
	h.startStateSaver()
	return nil
}

func (h *MeshtasticHook) OnConnect(cl *mqtt.Client, pk packets.Packet) error {
	h.logger.Debug().
		Str("client_id", cl.ID).
		Str("remote_addr", cl.Net.Remote).
		Uint8("packet_type", pk.FixedHeader.Type).
		Msg("client connected")
	return nil
}

func (h *MeshtasticHook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	logEvent := h.logger.Debug().
		Str("client_id", cl.ID).
		Str("remote_addr", cl.Net.Remote).
		Bool("expire", expire)

	if err != nil {
		logEvent = logEvent.Err(err)
	}

	var reason string
	if expire {
		if err != nil {
			reason = "error (network/timeout)"
		} else {
			reason = "abrupt (no DISCONNECT)"
		}
	} else {
		reason = "graceful (DISCONNECT)"
	}

	logEvent.Str("reason", reason).Msg("client disconnected")
}

func (h *MeshtasticHook) OnPublish(_ *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	//h.logger.Debug().
	//	Str("topic", pk.TopicName).
	//	Int("payload_size", len(pk.Payload)).
	//	Msg("received MQTT message")

	// Проверяем соответствие топика паттерну
	if !h.matchesTopicPattern(pk.TopicName) {
		//h.logger.Debug().
		//	Str("topic", pk.TopicName).
		//	Str("expected_prefix", h.config.TopicPrefix).
		//	Msg("topic does not match the prefix, skipping")
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

func (h *MeshtasticHook) validateStateFile() error {
	if h.factory == nil {
		return nil
	}

	prometheusConfig := h.factory.GetPrometheusConfig()
	if prometheusConfig == nil || prometheusConfig.GetStateFile() == "" {
		return nil
	}

	stateFile := prometheusConfig.GetStateFile()
	if err := h.checkFileWritable(stateFile); err != nil {
		return fmt.Errorf("cannot write to state file %s: %w", stateFile, err)
	}

	h.logger.Debug().Str("file", stateFile).Msg("state file validation successful")
	return nil
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
				}
			case <-h.stopSave:
				return
			}
		}
	}()
}

func (h *MeshtasticHook) checkFileWritable(filename string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, domain.StateFilePermissions)
	if err != nil {
		return err
	}
	file.Close()
	return nil
}

func (h *MeshtasticHook) matchesTopicPattern(topic string) bool {
	parts := strings.Split(topic, "/")
	patternParts := strings.Split(strings.TrimSuffix(h.config.TopicPrefix, "/"), "/")

	for i, patternPart := range patternParts {
		if patternPart == "#" {
			return true
		}
		if i >= len(parts) {
			return false
		}
		if patternPart != "+" && parts[i] != patternPart {
			return false
		}
	}
	return len(parts) >= len(patternParts)
}

func (h *MeshtasticHook) OnStopped() {
	if h.stopped {
		return
	}
	h.stopped = true

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
		ctx, cancel := context.WithTimeout(context.Background(), domain.DefaultTimeout)
		defer cancel()
		if err := h.server.Shutdown(ctx); err != nil {
			h.logger.Error().Err(err).Msg("unified server shutdown error")
		}
	}
}

func (h *MeshtasticHook) Shutdown(ctx context.Context) error {
	// Этот метод оставляем для совместимости с тестами
	h.OnStopped()
	return nil
}
