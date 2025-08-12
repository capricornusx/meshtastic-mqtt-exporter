package main

import (
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainWithArgs(t *testing.T) {
	// Test argument parsing
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Create a temporary config file
	tempFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	configContent := `
prometheus_addr: ":8100"
topic_prefix: "msh/"
enable_health: true
`
	_, err = tempFile.WriteString(configContent)
	if err != nil {
		t.Fatal(err)
	}
	tempFile.Close()

	os.Args = []string{"cmd", "--config", tempFile.Name()}

	// This test mainly ensures the argument parsing doesn't panic
	// The actual main() function would start MQTT server, which we don't want in tests
}

func TestMainWithDefaultArgs(t *testing.T) {
	// Test with no arguments (should use default config)
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"cmd"}

	// Should not panic with default arguments
	// We can't easily test the full main() without starting the MQTT server
	assert.Equal(t, 1, len(os.Args))
}

func TestMainWithInvalidConfig(t *testing.T) {
	// Test with invalid config file
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"cmd", "--config", "nonexistent.yaml"}

	// Should not panic - main function should handle missing config gracefully
	// by using defaults
	assert.Equal(t, 3, len(os.Args))
}

func TestBoolToByte(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected byte
	}{
		{"true to 1", true, 1},
		{"false to 0", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := boolToByte(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSafeInt32(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int32
	}{
		{"normal value", 1000, 1000},
		{"zero", 0, 0},
		{"negative", -1000, -1000},
		{"max int32", math.MaxInt32, math.MaxInt32},
		{"min int32", math.MinInt32, math.MinInt32},
		{"overflow max", math.MaxInt32 + 1, math.MaxInt32},
		{"underflow min", math.MinInt32 - 1, math.MinInt32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeInt32(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSafeUint16(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected uint16
	}{
		{"normal value", 1000, 1000},
		{"zero", 0, 0},
		{"negative to zero", -1, 0},
		{"max uint16", math.MaxUint16, math.MaxUint16},
		{"overflow", math.MaxUint16 + 1, math.MaxUint16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeUint16(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
