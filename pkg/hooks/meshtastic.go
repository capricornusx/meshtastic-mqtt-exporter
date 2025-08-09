// Package hooks provide a standalone Meshtastic Prometheus hook for mochi-mqtt
package hooks

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MeshtasticHookConfig configures the Meshtastic hook.
type MeshtasticHookConfig struct {
	PrometheusAddr string // Address to serve Prometheus metrics (e.g., ":8100")
	EnableHealth   bool   // Enable /health endpoint
}

// MeshtasticHook is a mochi-mqtt hook that exports Meshtastic telemetry to Prometheus.
type MeshtasticHook struct {
	mqtt.HookBase
	config MeshtasticHookConfig

	// Prometheus metrics
	messageCounter *prometheus.CounterVec
	batteryLevel   *prometheus.GaugeVec
	voltage        *prometheus.GaugeVec
	temperature    *prometheus.GaugeVec
	humidity       *prometheus.GaugeVec
	pressure       *prometheus.GaugeVec
	channelUtil    *prometheus.GaugeVec
	airUtilTx      *prometheus.GaugeVec
	uptime         *prometheus.GaugeVec
	nodeLastSeen   *prometheus.GaugeVec
	nodeHardware   *prometheus.GaugeVec

	registry *prometheus.Registry
}

// NewMeshtasticHook creates a new Meshtastic hook.
func NewMeshtasticHook(config MeshtasticHookConfig) *MeshtasticHook {
	h := &MeshtasticHook{
		config:   config,
		registry: prometheus.NewRegistry(),
	}
	h.setupMetrics()
	return h
}

// ID returns the hook identifier.
func (h *MeshtasticHook) ID() string {
	return "meshtastic-prometheus"
}

// Provides returns the hook events this hook provides.
func (h *MeshtasticHook) Provides(b byte) bool {
	return b == mqtt.OnPublish
}

// OnPublish processes published messages for Meshtastic telemetry.
func (h *MeshtasticHook) OnPublish(_ *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	// Only process Meshtastic topics
	if !strings.HasPrefix(pk.TopicName, "msh/") {
		return pk, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(pk.Payload, &data); err != nil {
		return pk, nil // Ignore invalid JSON
	}

	h.processMessage(data)
	return pk, nil
}

// Init starts the Prometheus server.
func (h *MeshtasticHook) Init(config any) error {
	if h.config.PrometheusAddr != "" {
		go h.startServer()
	}
	return nil
}

func (h *MeshtasticHook) setupMetrics() {
	h.messageCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "meshtastic_messages_total", Help: "Total messages by type"},
		[]string{"type", "from_node"})

	h.batteryLevel = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "meshtastic_battery_level_percent", Help: "Battery level"},
		[]string{"node_id"})

	h.voltage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "meshtastic_voltage_volts", Help: "Battery voltage"},
		[]string{"node_id"})

	h.temperature = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "meshtastic_temperature_celsius", Help: "Temperature"},
		[]string{"node_id"})

	h.humidity = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "meshtastic_humidity_percent", Help: "Humidity"},
		[]string{"node_id"})

	h.pressure = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "meshtastic_pressure_hpa", Help: "Pressure"},
		[]string{"node_id"})

	h.channelUtil = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "meshtastic_channel_utilization_percent", Help: "Channel utilization"},
		[]string{"node_id"})

	h.airUtilTx = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "meshtastic_air_util_tx_percent", Help: "Air utilization TX"},
		[]string{"node_id"})

	h.uptime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "meshtastic_uptime_seconds", Help: "Uptime"},
		[]string{"node_id"})

	h.nodeLastSeen = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "meshtastic_node_last_seen_timestamp", Help: "Last seen timestamp"},
		[]string{"node_id"})

	h.nodeHardware = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "meshtastic_node_info", Help: "Node information"},
		[]string{"node_id", "longname", "shortname", "hardware", "role"})

	// Register all metrics
	h.registry.MustRegister(
		h.messageCounter, h.batteryLevel, h.voltage, h.temperature,
		h.humidity, h.pressure, h.channelUtil, h.airUtilTx,
		h.uptime, h.nodeLastSeen, h.nodeHardware,
	)
}

func (h *MeshtasticHook) processMessage(data map[string]interface{}) {
	fromNode := getUint32(data, "from")
	if fromNode == 0 {
		return
	}

	nodeID := strconv.FormatUint(uint64(fromNode), 10)
	msgType := getString(data, "type")

	h.nodeLastSeen.WithLabelValues(nodeID).SetToCurrentTime()
	h.messageCounter.WithLabelValues(msgType, nodeID).Inc()

	if payload, ok := data["payload"].(map[string]interface{}); ok {
		h.processPayload(nodeID, msgType, payload)
	}
}

func (h *MeshtasticHook) processPayload(nodeID, msgType string, payload map[string]interface{}) {
	switch msgType {
	case "telemetry":
		h.processTelemetry(nodeID, payload)
	case "nodeinfo":
		h.processNodeInfo(nodeID, payload)
	}
}

func (h *MeshtasticHook) processTelemetry(nodeID string, payload map[string]interface{}) {
	if val, ok := payload["battery_level"].(float64); ok {
		h.batteryLevel.WithLabelValues(nodeID).Set(val)
	}
	if val, ok := payload["voltage"].(float64); ok {
		h.voltage.WithLabelValues(nodeID).Set(roundFloat(val, 2))
	}
	if val, ok := payload["temperature"].(float64); ok {
		h.temperature.WithLabelValues(nodeID).Set(roundFloat(val, 1))
	}
	if val, ok := payload["relative_humidity"].(float64); ok {
		h.humidity.WithLabelValues(nodeID).Set(roundFloat(val, 1))
	}
	if val, ok := payload["barometric_pressure"].(float64); ok {
		h.pressure.WithLabelValues(nodeID).Set(roundFloat(val, 1))
	}
	if val, ok := payload["channel_utilization"].(float64); ok {
		h.channelUtil.WithLabelValues(nodeID).Set(roundFloat(val, 2))
	}
	if val, ok := payload["air_util_tx"].(float64); ok {
		h.airUtilTx.WithLabelValues(nodeID).Set(roundFloat(val, 2))
	}
	if val, ok := payload["uptime_seconds"].(float64); ok {
		h.uptime.WithLabelValues(nodeID).Set(val)
	}
}

func (h *MeshtasticHook) processNodeInfo(nodeID string, payload map[string]interface{}) {
	longname := getString(payload, "longname")
	shortname := getString(payload, "shortname")
	hardware := "unknown"
	role := "unknown"

	if val, ok := payload["hardware"].(float64); ok {
		hardware = strconv.FormatFloat(val, 'f', 0, 64)
	}
	if val, ok := payload["role"].(float64); ok {
		role = strconv.FormatFloat(val, 'f', 0, 64)
	}

	h.nodeHardware.WithLabelValues(nodeID, longname, shortname, hardware, role).Set(1)
}

func (h *MeshtasticHook) startServer() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(h.registry, promhttp.HandlerOpts{}))

	if h.config.EnableHealth {
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "ok",
				"service": "meshtastic-hook",
			})
		})
	}

	log.Printf("Meshtastic hook: Prometheus server starting on %s", h.config.PrometheusAddr)
	if err := http.ListenAndServe(h.config.PrometheusAddr, mux); err != nil {
		log.Printf("Meshtastic hook: Failed to start server: %v", err)
	}
}

// Helper functions are shared with prometheus.go