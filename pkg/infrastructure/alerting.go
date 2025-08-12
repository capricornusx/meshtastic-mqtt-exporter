package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/rs/zerolog"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/logger"
)

const (
	defaultChannel = "LongFast"
	defaultMode    = "broadcast"
)

type LoRaAlertSender struct {
	mqttServer     *mqtt.Server
	defaultChannel string
	defaultMode    string
	targetNodes    []string
	logger         zerolog.Logger
}

type LoRaConfig struct {
	DefaultChannel string
	DefaultMode    string
	TargetNodes    []string
}

func NewLoRaAlertSender(server *mqtt.Server, config LoRaConfig) *LoRaAlertSender {
	if config.DefaultChannel == "" {
		config.DefaultChannel = defaultChannel
	}
	if config.DefaultMode == "" {
		config.DefaultMode = defaultMode
	}

	return &LoRaAlertSender{
		mqttServer:     server,
		defaultChannel: config.DefaultChannel,
		defaultMode:    config.DefaultMode,
		targetNodes:    config.TargetNodes,
		logger:         logger.ComponentLogger("lora-alerter"),
	}
}

func (s *LoRaAlertSender) SendAlert(ctx context.Context, alert domain.Alert) error {
	channel := alert.Channel
	if channel == "" {
		channel = s.defaultChannel
	}

	mode := alert.Mode
	if mode == "" {
		mode = s.defaultMode
	}

	targets := alert.TargetNodes
	if len(targets) == 0 {
		targets = s.targetNodes
	}

	if mode == defaultMode {
		return s.sendBroadcast(ctx, alert.Message, channel)
	}

	return s.sendDirect(ctx, alert.Message, channel, targets)
}

func (s *LoRaAlertSender) sendBroadcast(_ context.Context, message, channel string) error {
	topic := fmt.Sprintf("msh/2/c/%s/!broadcast", channel)
	payload := map[string]interface{}{
		"type":    "text", //TODO возможно это поле лишнее.
		"payload": fmt.Sprintf("[ALERT] %s", message),
		"from":    domain.LoRaBroadcastNodeID, //TODO возможно это поле лишнее.
		"to":      domain.LoRaBroadcastNodeID,
	}

	return s.publishMessage(topic, payload)
}

func (s *LoRaAlertSender) sendDirect(_ context.Context, message, channel string, targets []string) error {
	for _, nodeID := range targets {
		topic := fmt.Sprintf("msh/2/c/%s/!%s", channel, nodeID)
		payload := map[string]interface{}{
			"type":    "text",
			"payload": fmt.Sprintf("[ALERT] %s", message),
		}

		if err := s.publishMessage(topic, payload); err != nil {
			s.logger.Error().Err(err).Str("node", nodeID).Msg("failed to send direct alert")
			return err
		}
	}
	return nil
}

func (s *LoRaAlertSender) publishMessage(topic string, payload map[string]interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	if s.mqttServer != nil {
		if err := s.mqttServer.Publish(topic, data, false, 0); err != nil {
			return fmt.Errorf("publish to topic %s: %w", topic, err)
		}
		s.logger.Info().Str("topic", topic).Msg("alert sent to LoRa")
	}

	return nil
}
