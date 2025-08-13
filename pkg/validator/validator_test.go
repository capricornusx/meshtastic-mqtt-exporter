package validator

import "testing"

func TestValidateMeshtasticMessage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		payload []byte
		wantErr bool
	}{
		{"valid JSON", []byte(`{"from":123,"type":"telemetry"}`), false},
		{"invalid JSON", []byte(`{invalid`), true},
		{"empty payload", []byte(``), true},
		{"non-JSON", []byte(`plain text`), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateMeshtasticMessage(tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMeshtasticMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTopicName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		topic   string
		wantErr bool
	}{
		{"valid topic", "msh/123/json/LongFast/telemetry", false},
		{"empty topic", "", true},
		{"too long topic", string(make([]byte, 300)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateTopicName(tt.topic)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTopicName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNodeID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		nodeID  string
		wantErr bool
	}{
		{"valid node ID", "123456789", false},
		{"empty node ID", "", true},
		{"too long node ID", string(make([]byte, 50)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateNodeID(tt.nodeID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNodeID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal string", "test", "test"},
		{"string with newlines", "test\nline", "testline"},
		{"string with tabs", "test\tvalue", "testvalue"},
		{"string with carriage return", "test\rvalue", "testvalue"},
		{"mixed whitespace", "test\n\t\rvalue", "testvalue"},
		{"ascii only", "Hello World!", "Hello World!"},
		{"with non-ascii", "test\x01\x02", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMatchesMQTTPattern(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		topic    string
		pattern  string
		expected bool
	}{
		{
			name:     "empty pattern matches all",
			topic:    "msh/test/topic",
			pattern:  "",
			expected: true,
		},
		{
			name:     "exact match",
			topic:    "msh/test/topic",
			pattern:  "msh/test/topic",
			expected: true,
		},
		{
			name:     "single level wildcard +",
			topic:    "msh/123/json",
			pattern:  "msh/+/json",
			expected: true,
		},
		{
			name:     "single level wildcard + no match",
			topic:    "msh/123/456/json",
			pattern:  "msh/+/json",
			expected: false,
		},
		{
			name:     "multi level wildcard #",
			topic:    "msh/123/json/telemetry/data",
			pattern:  "msh/#",
			expected: true,
		},
		{
			name:     "multi level wildcard # at end",
			topic:    "msh/123/json/telemetry",
			pattern:  "msh/123/json/#",
			expected: true,
		},
		{
			name:     "complex pattern with + and #",
			topic:    "msh/123/2/json/telemetry/data",
			pattern:  "msh/+/2/json/#",
			expected: true,
		},
		{
			name:     "complex pattern no match",
			topic:    "msh/123/3/json/telemetry",
			pattern:  "msh/+/2/json/#",
			expected: false,
		},
		{
			name:     "pattern from config example",
			topic:    "msh/node123/2/json/telemetry",
			pattern:  "msh/+/2/json/#",
			expected: true,
		},
		{
			name:     "pattern from config example no match",
			topic:    "msh/node123/1/json/telemetry",
			pattern:  "msh/+/2/json/#",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := MatchesMQTTPattern(tt.topic, tt.pattern)
			if result != tt.expected {
				t.Errorf("MatchesMQTTPattern(%q, %q) = %v, expected %v", tt.topic, tt.pattern, result, tt.expected)
			}
		})
	}
}
