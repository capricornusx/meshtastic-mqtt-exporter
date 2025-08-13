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
	defaultChannel           = "LongFast"
	defaultMode              = "broadcast"
	defaultMessageType       = "sendtext"
	defaultMQTTDownlinkTopic = "msh/MY_433/2/json/mqtt/"
)

// GetDefaultMQTTDownlinkTopic возвращает дефолтный MQTT топик для тестов
func GetDefaultMQTTDownlinkTopic() string {
	return defaultMQTTDownlinkTopic
}

type LoRaAlertSender struct {
	mqttServer        *mqtt.Server
	defaultChannel    string
	defaultMode       string
	targetNodes       []uint32
	fromNodeID        uint32
	mqttDownlinkTopic string
	logger            zerolog.Logger
}

type LoRaConfig struct {
	DefaultChannel    string
	DefaultMode       string
	TargetNodes       []uint32
	FromNodeID        uint32
	MQTTDownlinkTopic string
}

func NewLoRaAlertSender(server *mqtt.Server, config LoRaConfig) *LoRaAlertSender {
	if config.DefaultChannel == "" {
		config.DefaultChannel = defaultChannel
	}
	if config.DefaultMode == "" {
		config.DefaultMode = defaultMode
	}
	if config.FromNodeID == 0 {
		config.FromNodeID = uint32(domain.LoRaBroadcastNodeID)
	}
	if config.MQTTDownlinkTopic == "" {
		config.MQTTDownlinkTopic = defaultMQTTDownlinkTopic
	}

	return &LoRaAlertSender{
		mqttServer:        server,
		defaultChannel:    config.DefaultChannel,
		defaultMode:       config.DefaultMode,
		targetNodes:       config.TargetNodes,
		fromNodeID:        config.FromNodeID,
		mqttDownlinkTopic: config.MQTTDownlinkTopic,
		logger:            logger.ComponentLogger("lora-alerter"),
	}
}

func (s *LoRaAlertSender) SendAlert(ctx context.Context, alert domain.Alert) error {
	mode := alert.Mode
	if mode == "" {
		mode = s.defaultMode
	}

	targets := alert.TargetNodes
	if len(targets) == 0 {
		targets = s.targetNodes
	}

	if mode == defaultMode {
		return s.sendBroadcast(ctx, alert.Message)
	}

	return s.sendDirect(ctx, alert.Message, targets)
}

func (s *LoRaAlertSender) sendBroadcast(_ context.Context, message string) error {
	payload := map[string]interface{}{
		"type":    defaultMessageType,
		"payload": message,
		"from":    s.fromNodeID,
		"to":      uint32(domain.LoRaBroadcastNodeID),
	}

	return s.publishMessage(s.mqttDownlinkTopic, payload)
}

func (s *LoRaAlertSender) sendDirect(_ context.Context, message string, targets []uint32) error {
	for _, nodeID := range targets {
		payload := map[string]interface{}{
			"type":    defaultMessageType,
			"payload": message,
			"from":    s.fromNodeID,
			"to":      nodeID,
		}

		if err := s.publishMessage(s.mqttDownlinkTopic, payload); err != nil {
			s.logger.Error().Err(err).Uint32("node", nodeID).Msg("failed to send direct alert")
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
		s.logger.Info().Str("topic", topic).Msg("🔥 alert sent to MQTT->LoRa")
	}

	return nil
}
