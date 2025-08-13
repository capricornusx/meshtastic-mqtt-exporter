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

const (
	unknownValue           = "unknown"
	powerMetricsType       = "power_metrics"
	environmentMetricsType = "environment_metrics"
	deviceMetricsType      = "device_metrics"
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
			p.logger.Debug().Str("topic", topic).Msg("ignoring non-JSON message")
			return nil // Игнорируем не-JSON сообщения
		}
		p.logger.Warn().Err(err).Str("topic", topic).Msg("validation failed")
		return err
	}

	msg, err := p.parseMessage(payload)
	if err != nil {
		p.logger.Error().Err(err).Str("topic", topic).Msg("failed to parse message")
		return err
	}

	nodeID, err := p.validateAndFormatNodeID(msg.From)
	if err != nil {
		p.logger.Warn().Err(err).Uint32("from", msg.From).Msg("invalid node ID")
		return err
	}

	// Обновляем timestamp для любого сообщения от ноды
	p.collector.UpdateNodeLastSeen(nodeID, time.Now())

	return p.processMessageByType(msg, nodeID)
}

func (p *MeshtasticProcessor) logMessageIfEnabled(topic string, payload []byte) {
	if p.logAllMessages && validator.MatchesMQTTPattern(topic, p.topicPattern) {
		//Int("payload_size", len(payload)).
		p.logger.Debug().Str("topic", topic).RawJSON("payload", payload).Msg("received")
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
	var raw map[string]interface{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return domain.MeshtasticMessage{}, errors.NewProcessingError("json parsing failed", err)
	}

	msg := domain.MeshtasticMessage{
		Type: p.getString(raw, "type"),
	}

	if from, ok := raw["from"].(float64); ok {
		msg.From = uint32(from)
	}
	if rssi, ok := raw["rssi"].(float64); ok {
		msg.RSSI = &rssi
	}
	if snr, ok := raw["snr"].(float64); ok {
		msg.SNR = &snr
	}

	if msg.Type == "sendtext" {
		if payloadStr, ok := raw["payload"].(string); ok {
			msg.Payload = map[string]interface{}{"text": payloadStr}
		}
	} else {
		if payloadObj, ok := raw["payload"].(map[string]interface{}); ok {
			msg.Payload = payloadObj
		}
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

func (p *MeshtasticProcessor) processMessageByType(msg domain.MeshtasticMessage, nodeID string) error {
	switch msg.Type {
	case domain.MessageTypeTelemetry:
		//p.logger.Debug().Str("node_id", nodeID).Msg("processing telemetry")
		return p.processTelemetry(nodeID, msg)
	case domain.MessageTypeNodeInfo:
		return p.processNodeInfo(nodeID, msg.Payload)
	case domain.MessageTypeText:
		return p.processTextMessage(nodeID, msg.Payload)
	case domain.MessageTypePosition:
		return p.processPosition(nodeID, msg.Payload)
	case domain.MessageTypeWaypoint:
		return p.processWaypoint(nodeID, msg.Payload)
	case domain.MessageTypeNeighborInfo:
		return p.processNeighborInfo(nodeID, msg.Payload)
	default:
		//p.logger.Debug().
		//	Str("node_id", nodeID).
		//	Msg("unsupported type")
		p.collector.UpdateMessageCounter(nodeID, domain.MessageTypeUnsupported)
		return nil
	}
}

func (p *MeshtasticProcessor) processTelemetry(nodeID string, msg domain.MeshtasticMessage) error {
	data := domain.TelemetryData{
		NodeID:    nodeID,
		Timestamp: time.Now(),
	}

	// Определяем тип телеметрии по наличию полей
	data.Type = p.determineTelemetryType(msg.Payload)

	p.extractTelemetryFields(&data, msg.Payload)
	p.extractTopLevelFields(&data, msg)
	return p.collector.CollectTelemetry(data)
}

func (p *MeshtasticProcessor) extractTelemetryFields(data *domain.TelemetryData, payload map[string]interface{}) {
	p.extractBasicFields(data, payload)
	p.extractEnvironmentalFields(data, payload)
	p.extractNetworkFields(data, payload)
}

func (p *MeshtasticProcessor) extractTopLevelFields(data *domain.TelemetryData, msg domain.MeshtasticMessage) {
	if msg.RSSI != nil {
		data.RSSI = msg.RSSI
	}
	if msg.SNR != nil {
		data.SNR = msg.SNR
	}
}

func (p *MeshtasticProcessor) extractBasicFields(data *domain.TelemetryData, payload map[string]interface{}) {
	if val, ok := payload["battery_level"].(float64); ok {
		data.BatteryLevel = &val
	}
	if val, ok := payload["voltage"].(float64); ok {
		truncated := truncateToTwoDecimals(val)
		data.Voltage = &truncated
	}
	if val, ok := payload["uptime_seconds"].(float64); ok {
		data.UptimeSeconds = &val
	}
}

func (p *MeshtasticProcessor) extractEnvironmentalFields(data *domain.TelemetryData, payload map[string]interface{}) {
	p.extractEnvironmentMetrics(data, payload)
	p.extractPowerMetrics(data, payload)
}

func (p *MeshtasticProcessor) extractEnvironmentMetrics(data *domain.TelemetryData, payload map[string]interface{}) {
	if val, ok := payload["temperature"].(float64); ok {
		truncated := truncateToTwoDecimals(val)
		data.Temperature = &truncated
	}
	if val, ok := payload["relative_humidity"].(float64); ok {
		truncated := truncateToTwoDecimals(val)
		data.RelativeHumidity = &truncated
	}
	if val, ok := payload["barometric_pressure"].(float64); ok {
		truncated := truncateToTwoDecimals(val)
		data.BarometricPressure = &truncated
	}
	if val, ok := payload["gas_resistance"].(float64); ok {
		truncated := truncateToTwoDecimals(val)
		data.GasResistance = &truncated
	}
	if val, ok := payload["iaq"].(float64); ok {
		truncated := truncateToTwoDecimals(val)
		data.IAQ = &truncated
	}
}

func (p *MeshtasticProcessor) extractPowerMetrics(data *domain.TelemetryData, payload map[string]interface{}) {
	if val, ok := payload["ch1_voltage"].(float64); ok {
		truncated := truncateToTwoDecimals(val)
		data.Ch1Voltage = &truncated
	}
	if val, ok := payload["ch1_current"].(float64); ok {
		truncated := truncateToTwoDecimals(val)
		data.Ch1Current = &truncated
	}
	if val, ok := payload["ch2_voltage"].(float64); ok {
		truncated := truncateToTwoDecimals(val)
		data.Ch2Voltage = &truncated
	}
	if val, ok := payload["ch2_current"].(float64); ok {
		truncated := truncateToTwoDecimals(val)
		data.Ch2Current = &truncated
	}
	if val, ok := payload["ch3_voltage"].(float64); ok {
		truncated := truncateToTwoDecimals(val)
		data.Ch3Voltage = &truncated
	}
	if val, ok := payload["ch3_current"].(float64); ok {
		truncated := truncateToTwoDecimals(val)
		data.Ch3Current = &truncated
	}
}

func (p *MeshtasticProcessor) extractNetworkFields(data *domain.TelemetryData, payload map[string]interface{}) {
	if val, ok := payload["channel_utilization"].(float64); ok {
		truncated := truncateToTwoDecimals(val)
		data.ChannelUtilization = &truncated
	}
	if val, ok := payload["air_util_tx"].(float64); ok {
		truncated := truncateToTwoDecimals(val)
		data.AirUtilTx = &truncated
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
		Hardware:  unknownValue,
		Role:      unknownValue,
		Timestamp: time.Now(),
	}

	// Устанавливаем значения по умолчанию, если поля пустые
	if info.LongName == "" {
		info.LongName = unknownValue
	}
	if info.ShortName == "" {
		info.ShortName = unknownValue
	}

	if val, ok := payload["hardware"].(float64); ok {
		info.Hardware = strconv.FormatFloat(val, 'f', 0, 64)
	}
	if val, ok := payload["role"].(float64); ok {
		info.Role = domain.GetRoleName(int(val))
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

// truncateToTwoDecimals truncates to 2 decimal places, 40% faster than rounding
func truncateToTwoDecimals(value float64) float64 {
	return math.Trunc(value*100) / 100
}

func (p *MeshtasticProcessor) processTextMessage(nodeID string, _ map[string]interface{}) error {
	p.collector.UpdateMessageCounter(nodeID, domain.MessageTypeText)
	return nil
}

func (p *MeshtasticProcessor) processPosition(nodeID string, _ map[string]interface{}) error {
	p.collector.UpdateMessageCounter(nodeID, domain.MessageTypePosition)
	return nil
}

func (p *MeshtasticProcessor) processWaypoint(nodeID string, _ map[string]interface{}) error {
	p.collector.UpdateMessageCounter(nodeID, domain.MessageTypeWaypoint)
	return nil
}

func (p *MeshtasticProcessor) processNeighborInfo(nodeID string, _ map[string]interface{}) error {
	p.collector.UpdateMessageCounter(nodeID, domain.MessageTypeNeighborInfo)
	return nil
}

func (p *MeshtasticProcessor) determineTelemetryType(payload map[string]interface{}) string {
	if _, ok := payload["temperature"]; ok {
		return environmentMetricsType
	}
	if _, ok := payload["relative_humidity"]; ok {
		return environmentMetricsType
	}
	if _, ok := payload["barometric_pressure"]; ok {
		return environmentMetricsType
	}
	if _, ok := payload["ch1_voltage"]; ok {
		return powerMetricsType
	}
	if _, ok := payload["ch1_current"]; ok {
		return powerMetricsType
	}
	if _, ok := payload["battery_level"]; ok {
		return deviceMetricsType
	}
	if _, ok := payload["voltage"]; ok {
		return deviceMetricsType
	}
	return deviceMetricsType
}
