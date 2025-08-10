package main

import (
	"flag"
	"os"
	"time"

	"meshtastic-exporter/pkg/exporter"

	"github.com/rs/zerolog"
)

func main() {
	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	configFile := flag.String("config", "config.yaml", "Configuration file path")
	flag.Parse()

	config, err := exporter.LoadConfig(*configFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load config")
	}

	exp := exporter.New(config)
	exp.Init()

	logger.Info().Str("component", "standalone").Msg("Starting Meshtastic Exporter")
	if err := exp.Run(); err != nil {
		logger.Fatal().Err(err).Str("component", "standalone").Msg("Exporter failed")
	}
}
