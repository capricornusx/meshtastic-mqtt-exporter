package main

import (
	"flag"
	"log"

	"meshtastic-exporter/pkg/exporter"
)

func main() {
	configFile := flag.String("config", "config.yaml", "Configuration file path")
	flag.Parse()

	config, err := exporter.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	exp := exporter.New(config)
	exp.Init()

	log.Println("Starting Meshtastic Exporter (Go version)")
	if err := exp.Run(); err != nil {
		log.Fatalf("Exporter failed: %v", err)
	}
}
