package main

import (
	"flag"
	"time"

	"meshtastic-exporter/pkg/config"
	"meshtastic-exporter/pkg/logger"
	"meshtastic-exporter/pkg/standalone"

	"github.com/rs/zerolog"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339
	appLogger := logger.ComponentLogger("standalone")

	configFile := flag.String("config", "config.yaml", "Configuration file path")
	flag.Parse()

	cfg, err := config.LoadUnifiedConfig(*configFile)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("failed to load configuration")
	}

	app := standalone.NewApp(cfg)

	appLogger.Info().Msg("starting meshtastic exporter (standalone)")
	if err := app.Run(); err != nil {
		appLogger.Fatal().Err(err).Msg("application error")
	}
}
