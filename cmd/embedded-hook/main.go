package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/rs/zerolog"

	"meshtastic-exporter/pkg/config"
	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"
	"meshtastic-exporter/pkg/logger"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339
	appLogger := logger.ComponentLogger("embedded-hook")

	configFile := flag.String("config", "config.yaml", "Configuration file path")
	flag.Parse()

	cfg, err := config.LoadUnifiedConfig(*configFile)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("failed to load config")
	}

	server := setupMQTTServer(cfg, appLogger)
	startServer(server, appLogger)
	waitForShutdown(server, appLogger)
}

func setupMQTTServer(cfg domain.Config, logger zerolog.Logger) *mqtt.Server {
	f := factory.NewFactory(cfg)
	server := mqtt.New(&mqtt.Options{InlineClient: false})

	addMeshtasticHook(server, cfg, f, logger)
	addAuthHook(server, cfg, logger)
	addTCPListener(server, cfg, logger)

	return server
}

func addMeshtasticHook(server *mqtt.Server, cfg domain.Config, f *factory.Factory, logger zerolog.Logger) {
	prometheusConfig := cfg.GetPrometheusConfig()
	hookConfig := hooks.MeshtasticHookConfig{
		ServerAddr:   prometheusConfig.GetListen(),
		EnableHealth: true,
		TopicPrefix:  "msh/",
		MetricsTTL:   prometheusConfig.GetMetricsTTL(),
	}

	alertConfig := cfg.GetAlertManagerConfig()
	hookConfig.AlertPath = alertConfig.GetPath()

	hook := hooks.NewMeshtasticHook(hookConfig, f)
	if err := server.AddHook(hook, nil); err != nil {
		logger.Fatal().Err(err).Msg("failed to add meshtastic hook")
	}
	logger.Info().Msg("meshtastic hook enabled")
}

func addAuthHook(server *mqtt.Server, cfg domain.Config, logger zerolog.Logger) {
	mqttConfig := cfg.GetMQTTConfig()
	users := mqttConfig.GetUsers()
	var authRules auth.AuthRules

	if len(users) == 0 {
		authRules = append(authRules, auth.AuthRule{Allow: true})
	} else {
		for _, user := range users {
			authRules = append(authRules, auth.AuthRule{
				Username: auth.RString(user.GetUsername()),
				Password: auth.RString(user.GetPassword()),
				Allow:    true,
			})
		}
		authRules = append(authRules, auth.AuthRule{Allow: true})
	}

	if err := server.AddHook(new(auth.AllowHook), &auth.Options{
		Ledger: &auth.Ledger{Auth: authRules},
	}); err != nil {
		logger.Fatal().Err(err).Msg("failed to add auth hook")
	}
}

func addTCPListener(server *mqtt.Server, cfg domain.Config, logger zerolog.Logger) {
	mqttConfig := cfg.GetMQTTConfig()
	tcp := listeners.NewTCP(listeners.Config{
		ID:      "tcp",
		Address: fmt.Sprintf("%s:%d", mqttConfig.GetHost(), mqttConfig.GetPort()),
	})
	if err := server.AddListener(tcp); err != nil {
		logger.Fatal().Err(err).Msg("failed to add listener")
	}

	logger.Info().Msg("prometheus metrics enabled")
}

func startServer(server *mqtt.Server, logger zerolog.Logger) {
	go func() {
		if err := server.Serve(); err != nil {
			logger.Error().Err(err).Msg("mqtt server error")
		}
	}()
	logger.Info().Msg("mqtt server with meshtastic hooks started")
}

func waitForShutdown(server *mqtt.Server, logger zerolog.Logger) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logger.Info().Msg("shutting down")
	server.Close()
}
