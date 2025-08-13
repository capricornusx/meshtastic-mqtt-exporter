package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/logger"
)

type StandaloneLoRaAlertSender struct {
	mqttClient        mqtt.Client
	defaultChannel    string
	defaultMode       string
	targetNodes       []uint32
	fromNodeID        uint32
	mqttDownlinkTopic string
	logger            zerolog.Logger
}

func NewStandaloneLoRaAlertSender(mqttClient mqtt.Client, config LoRaConfig) *StandaloneLoRaAlertSender {
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

	return &StandaloneLoRaAlertSender{
		mqttClient:        mqttClient,
		defaultChannel:    config.DefaultChannel,
		defaultMode:       config.DefaultMode,
		targetNodes:       config.TargetNodes,
		fromNodeID:        config.FromNodeID,
		mqttDownlinkTopic: config.MQTTDownlinkTopic,
		logger:            logger.ComponentLogger("standalone-lora-alerter"),
	}
}

func (s *StandaloneLoRaAlertSender) SendAlert(ctx context.Context, alert domain.Alert) error {
	if s.mqttClient == nil || !s.mqttClient.IsConnected() {
		return fmt.Errorf("mqtt client not connected")
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
		return s.sendBroadcast(ctx, alert.Message)
	}

	return s.sendDirect(ctx, alert.Message, targets)
}

func (s *StandaloneLoRaAlertSender) sendBroadcast(_ context.Context, message string) error {
	payload := map[string]interface{}{
		"type":    defaultMessageType,
		"payload": message,
		"from":    s.fromNodeID,
		"to":      uint32(domain.LoRaBroadcastNodeID),
	}

	return s.publishMessage(s.mqttDownlinkTopic, payload)
}

func (s *StandaloneLoRaAlertSender) sendDirect(_ context.Context, message string, targets []uint32) error {
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

func (s *StandaloneLoRaAlertSender) publishMessage(topic string, payload map[string]interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	token := s.mqttClient.Publish(topic, 0, false, data)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("publish to topic %s: %w", topic, token.Error())
	}

	s.logger.Info().Str("topic", topic).Msg("🔥 alert sent to MQTT->LoRa")
	return nil
}
