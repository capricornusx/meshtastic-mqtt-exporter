package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"math"
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
	mqttConfig := cfg.GetMQTTConfig()
	server := mqtt.New(&mqtt.Options{
		InlineClient: false,
		Capabilities: &mqtt.Capabilities{
			MaximumInflight:              safeUint16(mqttConfig.GetMaxInflight()),
			MaximumClientWritesPending:   safeInt32(mqttConfig.GetMaxQueued()),
			ReceiveMaximum:               safeUint16(mqttConfig.GetReceiveMaximum()),
			MaximumQos:                   byte(mqttConfig.GetMaxQoS()),
			RetainAvailable:              boolToByte(mqttConfig.GetRetainAvailable()),
			MaximumMessageExpiryInterval: mqttConfig.GetMessageExpiry(),
			MaximumClients:               int64(mqttConfig.GetMaxClients()),
		},
	})

	addMeshtasticHook(server, cfg, f, logger)
	addAuthHook(server, cfg, logger)
	addListener(server, cfg, logger)

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

func addListener(server *mqtt.Server, cfg domain.Config, logger zerolog.Logger) {
	mqttConfig := cfg.GetMQTTConfig()
	tcpAddress := fmt.Sprintf("%s:%d", mqttConfig.GetHost(), mqttConfig.GetPort())
	addTCPListener(server, tcpAddress, logger)

	tlsConfig := mqttConfig.GetTLSConfig()
	if tlsConfig.GetEnabled() {
		tlsAddress := fmt.Sprintf("%s:%d", mqttConfig.GetHost(), tlsConfig.GetPort())
		addTLSListener(server, tlsConfig, tlsAddress, logger)
	}
}

func addTCPListener(server *mqtt.Server, address string, logger zerolog.Logger) {
	tcp := listeners.NewTCP(listeners.Config{
		ID:      "tcp",
		Address: address,
	})
	if err := server.AddListener(tcp); err != nil {
		logger.Fatal().Err(err).Msg("failed to add tcp listener")
	}
	logger.Info().Str("address", address).Msg("tcp listener enabled")
}

func addTLSListener(server *mqtt.Server, tlsConfig domain.TLSConfig, address string, logger zerolog.Logger) {
	cert, err := tls.LoadX509KeyPair(tlsConfig.GetCertFile(), tlsConfig.GetKeyFile())
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load TLS certificate")
	}

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:        "tls",
		Address:   address,
		TLSConfig: tlsConf,
	})
	if err := server.AddListener(tcp); err != nil {
		logger.Fatal().Err(err).Msg("failed to add tls listener")
	}
	logger.Info().Str("address", address).Msg("tls listener enabled")
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

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func safeUint16(val int) uint16 {
	if val < 0 {
		return 0
	}
	if val > math.MaxUint16 {
		return math.MaxUint16
	}
	return uint16(val)
}

func safeInt32(val int) int32 {
	if val < math.MinInt32 {
		return math.MinInt32
	}
	if val > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(val)
}
