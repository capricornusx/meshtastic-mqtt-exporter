package validator

import (
	"encoding/json"
	"fmt"
	"strings"

	"meshtastic-exporter/pkg/domain"
)

func ValidateMeshtasticMessage(payload []byte) error {
	if len(payload) == 0 {
		return fmt.Errorf("empty payload")
	}

	if len(payload) > 1024*1024 { // 1MB limit
		return fmt.Errorf("payload too large: %d bytes", len(payload))
	}

	// Check if payload looks like JSON
	if !isLikelyJSON(payload) {
		return fmt.Errorf("not JSON format")
	}

	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Check required fields
	if _, ok := data["from"]; !ok {
		return fmt.Errorf("missing 'from' field")
	}

	return nil
}

func isLikelyJSON(payload []byte) bool {
	if len(payload) == 0 {
		return false
	}

	// Trim whitespace
	start := 0
	for start < len(payload) && (payload[start] == ' ' || payload[start] == '\t' || payload[start] == '\n' || payload[start] == '\r') {
		start++
	}

	if start >= len(payload) {
		return false
	}

	// JSON должен начинаться с { или [
	return payload[start] == '{' || payload[start] == '['
}

func ValidateTopicName(topic string) error {
	if len(topic) == 0 {
		return fmt.Errorf("empty topic")
	}

	if len(topic) > domain.MaxTopicLength {
		return fmt.Errorf("topic too long: %d chars", len(topic))
	}

	// Check for invalid characters
	if strings.Contains(topic, "\x00") {
		return fmt.Errorf("topic contains null byte")
	}

	return nil
}

func ValidateNodeID(nodeID string) error {
	if len(nodeID) == 0 {
		return fmt.Errorf("empty node ID")
	}

	if len(nodeID) > domain.MaxNodeIDLength {
		return fmt.Errorf("node ID too long: %d chars", len(nodeID))
	}

	// numeric ID
	for _, r := range nodeID {
		if r < '0' || r > '9' {
			return fmt.Errorf("invalid node ID format: %s", nodeID)
		}
	}

	return nil
}

func SanitizeString(s string) string {
	var result strings.Builder
	for _, r := range s {
		if r >= 32 && r < 127 { // only ASCII
			result.WriteRune(r)
		}
	}

	sanitized := result.String()
	if len(sanitized) > domain.MaxTopicLength {
		sanitized = sanitized[:domain.MaxTopicLength]
	}

	return sanitized
}

// MatchesMQTTPattern проверяет соответствие топика MQTT паттерну с поддержкой wildcards + и #.
func MatchesMQTTPattern(topic, pattern string) bool {
	if pattern == "" {
		return true // пустой паттерн соответствует всем топикам
	}

	// Простая реализация MQTT pattern matching
	return matchMQTTPattern(topic, pattern)
}

// matchMQTTPattern рекурсивно сопоставляет топик с паттерном.
func matchMQTTPattern(topic, pattern string) bool {
	if pattern == "#" || (pattern == "" && topic == "") {
		return true
	}

	if pattern == "" || topic == "" {
		return false
	}

	patternSlash := strings.Index(pattern, "/")
	topicSlash := strings.Index(topic, "/")

	if patternSlash == -1 {
		return matchLastSegment(topic, pattern, topicSlash)
	}

	return matchSegments(topic, pattern, patternSlash, topicSlash)
}

func matchLastSegment(topic, pattern string, topicSlash int) bool {
	if pattern == "+" {
		return topicSlash == -1
	}
	if pattern == "#" {
		return true
	}
	return topic == pattern
}

func matchSegments(topic, pattern string, patternSlash, topicSlash int) bool {
	if topicSlash == -1 {
		return false
	}

	patternSegment := pattern[:patternSlash]
	patternRest := pattern[patternSlash+1:]
	topicSegment := topic[:topicSlash]
	topicRest := topic[topicSlash+1:]

	if patternSegment == "+" {
		return matchMQTTPattern(topicRest, patternRest)
	}

	if patternSegment == "#" {
		return true
	}

	if patternSegment != topicSegment {
		return false
	}

	return matchMQTTPattern(topicRest, patternRest)
}
