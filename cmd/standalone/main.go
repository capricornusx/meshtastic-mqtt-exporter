package main

import (
	"flag"
	"time"

	"meshtastic-exporter/pkg/exporter"
	"meshtastic-exporter/pkg/logger"

	"github.com/rs/zerolog"
)

func main() {
	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339
	appLogger := logger.ComponentLogger("standalone")

	configFile := flag.String("config", "config.yaml", "Configuration file path")
	flag.Parse()

	config, err := exporter.LoadConfig(*configFile)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to load config")
	}

	exp := exporter.New(config)
	exp.Init()

	appLogger.Info().Msg("Starting Meshtastic Exporter")
	if err := exp.Run(); err != nil {
		appLogger.Fatal().Err(err).Msg("Exporter failed")
	}
}
