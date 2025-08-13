package logger

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestComponentLogger(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer

	logger := zerolog.New(&buf).With().
		Timestamp().
		Str("component", "test").
		Logger()

	logger.Info().Msg("test message")

	output := buf.String()
	assert.Contains(t, output, "test message")
}

func TestSubLogger(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer

	baseLogger := zerolog.New(&buf).With().
		Timestamp().
		Str("component", "base").
		Logger()

	fields := map[string]string{
		"operation": "test_op",
		"user":      "test_user",
	}

	subLogger := SubLogger(baseLogger, fields)
	subLogger.Info().Msg("sub logger test")

	output := buf.String()
	assert.Contains(t, output, "sub logger test")
}

func TestSubLoggerEmptyFields(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer

	baseLogger := zerolog.New(&buf).With().
		Timestamp().
		Str("component", "base").
		Logger()

	fields := map[string]string{}

	subLogger := SubLogger(baseLogger, fields)
	subLogger.Info().Msg("empty fields test")

	output := buf.String()
	assert.Contains(t, output, "empty fields test")
}

func TestSetLogLevel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		level string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"warning", "warning"},
		{"error", "error"},
		{"fatal", "fatal"},
		{"invalid", "invalid"},
		{"uppercase", "DEBUG"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			SetLogLevel(tt.level)
			logger := ComponentLogger("test")
			assert.NotNil(t, logger)
		})
	}
}
