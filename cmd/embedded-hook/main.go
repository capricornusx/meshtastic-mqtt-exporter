package main

import (
	"flag"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"meshtastic-exporter/pkg/exporter"
	"meshtastic-exporter/pkg/hooks"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/rs/zerolog"
)

func main() {
	// Configure zerolog to match mochi-mqtt format
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Use console writer to match mochi-mqtt format
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()

	configFile := flag.String("config", "config.yaml", "Configuration file path")
	flag.Parse()

	config, err := exporter.LoadConfig(*configFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load config")
	}

	server := mqtt.New(&mqtt.Options{
		InlineClient: false,
	})

	// Add Prometheus hook
	prometheusHook := hooks.NewPrometheusHook(config)
	err = server.AddHook(prometheusHook, nil)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to add Prometheus hook")
	}

	// Configure authentication only if not allowing anonymous
	if !config.MQTT.AllowAnonymous {
		var authRules auth.AuthRules
		if len(config.MQTT.Users) > 0 {
			for _, user := range config.MQTT.Users {
				authRules = append(authRules, auth.AuthRule{
					Username: auth.RString(user.Username),
					Password: auth.RString(user.Password),
					Allow:    true,
				})
			}
		} else if config.MQTT.Username != "" {
			authRules = append(authRules, auth.AuthRule{
				Username: auth.RString(config.MQTT.Username),
				Password: auth.RString(config.MQTT.Password),
				Allow:    true,
			})
		}
		if len(authRules) > 0 {
			err := server.AddHook(new(auth.AllowHook), &auth.Options{
				Ledger: &auth.Ledger{Auth: authRules},
			})
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to add auth")
			}
		}
	} else {
		// Allow all connections when anonymous is enabled
		err := server.AddHook(new(auth.AllowHook), &auth.Options{
			Ledger: &auth.Ledger{
				Auth: auth.AuthRules{{
					Allow: true,
				}},
			},
		})
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to add anonymous auth")
		}
	}

	// Add TCP listener
	var addr string
	if config.MQTT.Host == "::" {
		addr = "[::]:" + strconv.Itoa(config.MQTT.Port)
	} else {
		addr = config.MQTT.Host + ":" + strconv.Itoa(config.MQTT.Port)
	}
	tcp := listeners.NewTCP(listeners.Config{ID: "tcp", Address: addr})
	err = server.AddListener(tcp)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to add listener")
	}

	go func() {
		err := server.Serve()
		if err != nil {
			logger.Error().Err(err).Msg("MQTT server error")
		}
	}()

	logger.Info().Str("component", "embedded").Str("address", addr).Msg("mqtt broker started")
	logger.Info().Str("component", "embedded").Str("metrics_url", config.Prometheus.Host+":"+strconv.Itoa(config.Prometheus.Port)+"/metrics").Msg("prometheus metrics available")
	time.Sleep(time.Second)

	// Wait for interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logger.Info().Str("component", "embedded").Msg("shutting down")
	prometheusHook.SaveStateOnShutdown()
	server.Close()
}
