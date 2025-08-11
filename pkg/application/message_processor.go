package application

import (
	"context"
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/errors"
	"meshtastic-exporter/pkg/logger"
	"meshtastic-exporter/pkg/validator"
)

type MeshtasticProcessor struct {
	collector      domain.MetricsCollector
	alerter        domain.AlertSender
	logger         zerolog.Logger
	logAllMessages bool
	topicPattern   string
}

func NewMeshtasticProcessor(collector domain.MetricsCollector, alerter domain.AlertSender, logAllMessages bool, topicPattern string) *MeshtasticProcessor {
	return &MeshtasticProcessor{
		collector:      collector,
		alerter:        alerter,
		logger:         logger.ComponentLogger("message-processor"),
		logAllMessages: logAllMessages,
		topicPattern:   topicPattern,
	}
}

func (p *MeshtasticProcessor) ProcessMessage(ctx context.Context, topic string, payload []byte) error {
	p.logMessageIfEnabled(topic, payload)

	if err := p.validateInput(topic, payload); err != nil {
		if strings.Contains(err.Error(), "not JSON") {
			return nil // Игнорируем не-JSON сообщения
		}
		return err
	}

	msg, err := p.parseMessage(payload)
	if err != nil {
		return err
	}

	nodeID, err := p.validateAndFormatNodeID(msg.From)
	if err != nil {
		return err
	}

	return p.processMessageByType(msg.Type, nodeID, msg.Payload)
}

func (p *MeshtasticProcessor) logMessageIfEnabled(topic string, payload []byte) {
	if p.logAllMessages && validator.MatchesMQTTPattern(topic, p.topicPattern) {
		p.logger.Debug().Str("topic", topic).RawJSON("payload", payload).Msg("mqtt message received")
	}
}

func (p *MeshtasticProcessor) validateInput(topic string, payload []byte) error {
	if err := validator.ValidateTopicName(topic); err != nil {
		p.logger.Warn().Err(err).Str("topic", topic).Msg("invalid topic")
		return errors.NewValidationError("invalid topic", err)
	}

	if err := validator.ValidateMeshtasticMessage(payload); err != nil {
		if strings.Contains(err.Error(), "not JSON format") {
			return errors.NewValidationError("not JSON", err)
		}
		p.logger.Warn().Err(err).Msg("invalid message")
		return errors.NewValidationError("invalid message", err)
	}

	return nil
}

func (p *MeshtasticProcessor) parseMessage(payload []byte) (domain.MeshtasticMessage, error) {
	var msg domain.MeshtasticMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		return msg, errors.NewProcessingError("json parsing failed", err)
	}
	return msg, nil
}

func (p *MeshtasticProcessor) validateAndFormatNodeID(from uint32) (string, error) {
	if from == 0 {
		return "", errors.NewValidationError("empty sender", nil)
	}

	nodeID := strconv.FormatUint(uint64(from), 10)
	if err := validator.ValidateNodeID(nodeID); err != nil {
		p.logger.Warn().Err(err).Str("node_id", nodeID).Msg("invalid node id")
		return "", errors.NewValidationError("invalid node id", err)
	}

	return nodeID, nil
}

func (p *MeshtasticProcessor) processMessageByType(msgType, nodeID string, payload map[string]interface{}) error {
	switch msgType {
	case domain.MessageTypeTelemetry:
		return p.processTelemetry(nodeID, payload)
	case domain.MessageTypeNodeInfo:
		return p.processNodeInfo(nodeID, payload)
	}
	return nil
}

func (p *MeshtasticProcessor) processTelemetry(nodeID string, payload map[string]interface{}) error {
	data := domain.TelemetryData{
		NodeID:    nodeID,
		Timestamp: time.Now(),
	}

	p.extractTelemetryFields(&data, payload)
	return p.collector.CollectTelemetry(data)
}

func (p *MeshtasticProcessor) extractTelemetryFields(data *domain.TelemetryData, payload map[string]interface{}) {
	p.extractBasicFields(data, payload)
	p.extractEnvironmentalFields(data, payload)
	p.extractNetworkFields(data, payload)
}

func (p *MeshtasticProcessor) extractBasicFields(data *domain.TelemetryData, payload map[string]interface{}) {
	if val, ok := payload["battery_level"].(float64); ok {
		data.BatteryLevel = &val
	}
	if val, ok := payload["voltage"].(float64); ok {
		rounded := roundToTwoDecimals(val)
		data.Voltage = &rounded
	}
	if val, ok := payload["uptime_seconds"].(float64); ok {
		data.UptimeSeconds = &val
	}
}

func (p *MeshtasticProcessor) extractEnvironmentalFields(data *domain.TelemetryData, payload map[string]interface{}) {
	if val, ok := payload["temperature"].(float64); ok {
		rounded := roundToTwoDecimals(val)
		data.Temperature = &rounded
	}
	if val, ok := payload["relative_humidity"].(float64); ok {
		rounded := roundToTwoDecimals(val)
		data.RelativeHumidity = &rounded
	}
	if val, ok := payload["barometric_pressure"].(float64); ok {
		rounded := roundToTwoDecimals(val)
		data.BarometricPressure = &rounded
	}
}

func (p *MeshtasticProcessor) extractNetworkFields(data *domain.TelemetryData, payload map[string]interface{}) {
	if val, ok := payload["channel_utilization"].(float64); ok {
		rounded := roundToTwoDecimals(val)
		data.ChannelUtilization = &rounded
	}
	if val, ok := payload["air_util_tx"].(float64); ok {
		rounded := roundToTwoDecimals(val)
		data.AirUtilTx = &rounded
	}
	if val, ok := payload["rssi"].(float64); ok {
		data.RSSI = &val
	}
	if val, ok := payload["snr"].(float64); ok {
		data.SNR = &val
	}
}

func (p *MeshtasticProcessor) processNodeInfo(nodeID string, payload map[string]interface{}) error {
	info := domain.NodeInfo{
		NodeID:    nodeID,
		LongName:  validator.SanitizeString(p.getString(payload, "longname")),
		ShortName: validator.SanitizeString(p.getString(payload, "shortname")),
		Hardware:  "unknown",
		Role:      "unknown",
		Timestamp: time.Now(),
	}

	if val, ok := payload["hardware"].(float64); ok {
		info.Hardware = strconv.FormatFloat(val, 'f', 0, 64)
	}
	if val, ok := payload["role"].(float64); ok {
		info.Role = strconv.FormatFloat(val, 'f', 0, 64)
	}

	return p.collector.CollectNodeInfo(info)
}

func (p *MeshtasticProcessor) getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func roundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}
