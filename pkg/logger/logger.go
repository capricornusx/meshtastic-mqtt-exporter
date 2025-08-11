package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var globalLogLevel = zerolog.InfoLevel

// SetLogLevel устанавливает глобальный уровень логирования.
func SetLogLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		globalLogLevel = zerolog.DebugLevel
	case "info":
		globalLogLevel = zerolog.InfoLevel
	case "warn", "warning":
		globalLogLevel = zerolog.WarnLevel
	case "error":
		globalLogLevel = zerolog.ErrorLevel
	case "fatal":
		globalLogLevel = zerolog.FatalLevel
	default:
		globalLogLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(globalLogLevel)
}

// ComponentLogger creates a logger with a predefined component field.
func ComponentLogger(component string) zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}).With().
		Timestamp().
		Str("component", component).
		Logger().Level(globalLogLevel)
}

// SubLogger creates a logger with additional context fields.
func SubLogger(base zerolog.Logger, fields map[string]string) zerolog.Logger {
	ctx := base.With()
	for k, v := range fields {
		ctx = ctx.Str(k, v)
	}
	return ctx.Logger()
}
