package hooks

import (
	"context"
	"fmt"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/rs/zerolog"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/infrastructure"
	"meshtastic-exporter/pkg/logger"
)

type AlertmanagerHook struct {
	mqtt.HookBase
	alerter domain.AlertSender
	server  *infrastructure.UnifiedServer
	config  AlertManagerConfig
	logger  zerolog.Logger
}

type AlertManagerConfig struct {
	HTTPHost string
	HTTPPort int
	HTTPPath string
}

func NewAlertmanagerHook(alerter domain.AlertSender, config AlertManagerConfig) *AlertmanagerHook {
	if config.HTTPPath == "" {
		config.HTTPPath = "/alerts/webhook"
	}
	if config.HTTPHost == "" {
		config.HTTPHost = "localhost"
	}
	if config.HTTPPort == 0 {
		config.HTTPPort = 8080
	}

	return &AlertmanagerHook{
		alerter: alerter,
		config:  config,
		logger:  logger.ComponentLogger("alertmanager"),
	}
}

func (h *AlertmanagerHook) ID() string {
	return "alertmanager-lora"
}

func (h *AlertmanagerHook) Provides(b byte) bool {
	return false // HTTP-only hook
}

func (h *AlertmanagerHook) Init(config any) error {
	return h.startHTTPServer()
}

func (h *AlertmanagerHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	return pk, nil
}

func (h *AlertmanagerHook) startHTTPServer() error {
	serverConfig := infrastructure.UnifiedServerConfig{
		Addr:         fmt.Sprintf("%s:%d", h.config.HTTPHost, h.config.HTTPPort),
		EnableHealth: false,
		AlertPath:    h.config.HTTPPath,
	}

	h.server = infrastructure.NewUnifiedServer(serverConfig, nil, h.alerter)
	return h.server.Start(context.Background())
}

func (h *AlertmanagerHook) Shutdown(ctx context.Context) error {
	if h.server != nil {
		return h.server.Shutdown(ctx)
	}
	return nil
}
