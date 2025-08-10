package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// ComponentLogger creates a logger with a predefined component field.
func ComponentLogger(component string) zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}).With().
		Timestamp().
		Str("component", component).
		Logger()
}

// SubLogger creates a logger with additional context fields.
func SubLogger(base zerolog.Logger, fields map[string]string) zerolog.Logger {
	ctx := base.With()
	for k, v := range fields {
		ctx = ctx.Str(k, v)
	}
	return ctx.Logger()
}
