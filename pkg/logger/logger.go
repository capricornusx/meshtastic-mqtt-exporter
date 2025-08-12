package logger

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	mu     sync.RWMutex
	level  zerolog.Level
	writer zerolog.ConsoleWriter
}

var defaultLogger = &Logger{
	level: zerolog.InfoLevel,
	writer: zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	},
}

func SetLogLevel(level string) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()

	switch strings.ToLower(level) {
	case "debug":
		defaultLogger.level = zerolog.DebugLevel
	case "info":
		defaultLogger.level = zerolog.InfoLevel
	case "warn", "warning":
		defaultLogger.level = zerolog.WarnLevel
	case "error":
		defaultLogger.level = zerolog.ErrorLevel
	case "fatal":
		defaultLogger.level = zerolog.FatalLevel
	default:
		defaultLogger.level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(defaultLogger.level)
}

func ComponentLogger(component string) zerolog.Logger {
	defaultLogger.mu.RLock()
	defer defaultLogger.mu.RUnlock()

	return zerolog.New(defaultLogger.writer).With().
		Timestamp().
		Str("component", component).
		Logger().Level(defaultLogger.level)
}

func SubLogger(base zerolog.Logger, fields map[string]string) zerolog.Logger {
	ctx := base.With()
	for k, v := range fields {
		ctx = ctx.Str(k, v)
	}
	return ctx.Logger()
}
