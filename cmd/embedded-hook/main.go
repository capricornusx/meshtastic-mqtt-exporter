package main

import (
	"flag"
	"log"
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
	// Suppress connection warnings
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	configFile := flag.String("config", "config.yaml", "Configuration file path")
	flag.Parse()

	config, err := exporter.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	server := mqtt.New(&mqtt.Options{
		InlineClient: false,
	})

	// Add Prometheus hook
	prometheusHook := hooks.NewPrometheusHook(config)
	err = server.AddHook(prometheusHook, nil)
	if err != nil {
		log.Fatalf("Failed to add Prometheus hook: %v", err)
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
				log.Fatalf("Failed to add auth: %v", err)
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
			log.Fatalf("Failed to add anonymous auth: %v", err)
		}
	}

	// Add TCP listener
	addr := config.MQTT.Host + ":" + strconv.Itoa(config.MQTT.Port)
	tcp := listeners.NewTCP(listeners.Config{ID: "tcp", Address: addr})
	err = server.AddListener(tcp)
	if err != nil {
		log.Fatalf("Failed to add listener: %v", err)
	}

	go func() {
		err := server.Serve()
		if err != nil {
			log.Printf("MQTT server error: %v", err)
		}
	}()

	log.Printf("MQTT broker with Prometheus hook started on %s", addr)
	time.Sleep(time.Second)

	// Wait for interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down...")
	if prometheusHook.Config.State.Enabled {
		prometheusHook.SaveState()
	}
	server.Close()
}
