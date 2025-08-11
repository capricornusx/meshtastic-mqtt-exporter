package validator

import "testing"

func TestMatchesMQTTPattern(t *testing.T) {
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
			result := MatchesMQTTPattern(tt.topic, tt.pattern)
			if result != tt.expected {
				t.Errorf("MatchesMQTTPattern(%q, %q) = %v, expected %v", tt.topic, tt.pattern, result, tt.expected)
			}
		})
	}
}
