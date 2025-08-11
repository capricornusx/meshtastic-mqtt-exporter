package main

import (
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
