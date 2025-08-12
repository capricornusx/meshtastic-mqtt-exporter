// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 mochi-mqtt
// SPDX-FileContributor: capricornusx

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"

	"meshtastic-exporter/pkg/config"
	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"
)

func main() {
	// Create mochi-mqtt server
	server := mqtt.New(&mqtt.Options{
		InlineClient: false,
	})

	// Load minimal config for factory
	cfg, err := config.LoadUnifiedConfig("../../config.yaml")
	if err != nil {
		// Use defaults if config not found
		log.Printf("Config not found, using defaults: %v", err)
	}
	f := factory.NewFactory(cfg)

	// Add Meshtastic Prometheus hook - simple version
	meshtasticHook := hooks.NewMeshtasticHookSimple(f)

	if err := server.AddHook(meshtasticHook, nil); err != nil {
		log.Fatalf("Failed to add Meshtastic hook: %v", err)
	}

	// Optional: Add authentication
	if err := server.AddHook(new(auth.AllowHook), &auth.Options{
		Ledger: &auth.Ledger{
			Auth: auth.AuthRules{{
				Allow: true, // Allow all connections
			}},
		},
	}); err != nil {
		log.Fatalf("Failed to add auth hook: %v", err)
	}

	// Add TCP listener
	tcp := listeners.NewTCP(listeners.Config{
		ID:      "tcp",
		Address: ":1886", // Avoid conflict with existing MQTT
	})
	if err := server.AddListener(tcp); err != nil {
		log.Fatalf("Failed to add listener: %v", err)
	}

	// Start server
	go func() {
		if err := server.Serve(); err != nil {
			log.Printf("MQTT server error: %v", err)
		}
	}()

	log.Println("MQTT server with Meshtastic hook started")
	log.Println("Prometheus metrics: http://localhost:8100/metrics")
	log.Println("Health check: http://localhost:8100/health")

	// Wait for interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down...")
	server.Close()
}
