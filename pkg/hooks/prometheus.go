package hooks

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/rs/zerolog"

	"meshtastic-exporter/pkg/exporter"
)

type PrometheusHook struct {
	mqtt.HookBase
	Config   exporter.Config
	registry *prometheus.Registry

	messageCounter *prometheus.CounterVec
	rssi           *prometheus.GaugeVec
	snr            *prometheus.GaugeVec
	batteryLevel   *prometheus.GaugeVec
	voltage        *prometheus.GaugeVec
	channelUtil    *prometheus.GaugeVec
	airUtilTx      *prometheus.GaugeVec
	uptime         *prometheus.GaugeVec
	temperature    *prometheus.GaugeVec
	humidity       *prometheus.GaugeVec
	pressure       *prometheus.GaugeVec
	nodeLastSeen   *prometheus.GaugeVec
	mqttUp         prometheus.Gauge
	nodeHardware   *prometheus.GaugeVec

	nodeMetrics map[string]time.Time
	mutex       sync.RWMutex
}

func NewPrometheusHook(config exporter.Config) *PrometheusHook {
	h := &PrometheusHook{
		Config:      config,
		registry:    prometheus.NewRegistry(),
		nodeMetrics: make(map[string]time.Time),
	}
	h.setupMetrics()
	h.mqttUp.Set(1) // MQTT server is up when hook initializes
	if config.State.Enabled {
		h.loadState()
	}
	h.startServer()
	h.startCleanupRoutine()
	return h
}

func (h *PrometheusHook) ID() string {
	return "prometheus-exporter"
}

func (h *PrometheusHook) Provides(b byte) bool {
	return b == mqtt.OnPublish
}

func (h *PrometheusHook) OnPublish(_ *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	if pk.TopicName == "" || len(pk.Payload) == 0 {
		return pk, nil
	}

	if len(pk.TopicName) < 4 || pk.TopicName[:4] != "msh/" {
		return pk, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(pk.Payload, &data); err != nil {
		return pk, nil
	}

	h.processMessage(data)
	return pk, nil
}

func (h *PrometheusHook) setupMetrics() {
	h.messageCounter = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "meshtastic_messages_total", Help: "Total messages by type"}, []string{"type", "from_node"})
	h.rssi = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_rssi_dbm", Help: "RSSI"}, []string{"from_node", "to_node"})
	h.snr = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_snr_db", Help: "SNR"}, []string{"from_node", "to_node"})
	h.batteryLevel = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_battery_level_percent", Help: "Battery level"}, []string{"node_id"})
	h.voltage = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_voltage_volts", Help: "Battery voltage"}, []string{"node_id"})
	h.channelUtil = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_channel_utilization_percent", Help: "Channel utilization"}, []string{"node_id"})
	h.airUtilTx = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_air_util_tx_percent", Help: "Air utilization TX"}, []string{"node_id"})
	h.uptime = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_uptime_seconds", Help: "Uptime"}, []string{"node_id"})
	h.temperature = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_temperature_celsius", Help: "Temperature"}, []string{"node_id"})
	h.humidity = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_humidity_percent", Help: "Humidity"}, []string{"node_id"})
	h.pressure = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_pressure_hpa", Help: "Pressure"}, []string{"node_id"})
	h.nodeLastSeen = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_node_last_seen_timestamp", Help: "Last seen timestamp"}, []string{"node_id"})
	h.mqttUp = prometheus.NewGauge(prometheus.GaugeOpts{Name: "meshtastic_mqtt_up", Help: "MQTT connection status"})
	h.nodeHardware = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_node_info", Help: "Node information"}, []string{"node_id", "longname", "shortname", "hardware", "role"})

	h.registry.MustRegister(h.messageCounter, h.rssi, h.snr, h.batteryLevel, h.voltage, h.channelUtil, h.airUtilTx, h.uptime, h.temperature, h.humidity, h.pressure, h.nodeLastSeen, h.mqttUp, h.nodeHardware)
}

func (h *PrometheusHook) processMessage(data map[string]interface{}) {
	fromNode := getUint32(data, "from")
	toNode := getUint32(data, "to")
	msgType := getString(data, "type")
	if fromNode == 0 {
		return
	}
	nodeID := strconv.FormatUint(uint64(fromNode), 10)

	h.mutex.Lock()
	h.nodeMetrics[nodeID] = time.Now()
	h.mutex.Unlock()

	h.nodeLastSeen.WithLabelValues(nodeID).SetToCurrentTime()
	h.messageCounter.WithLabelValues(msgType, nodeID).Inc()
	if rssi, ok := data["rssi"].(float64); ok {
		h.rssi.WithLabelValues(nodeID, strconv.FormatUint(uint64(toNode), 10)).Set(rssi)
	}
	if snr, ok := data["snr"].(float64); ok {
		h.snr.WithLabelValues(nodeID, strconv.FormatUint(uint64(toNode), 10)).Set(roundFloat(snr, 1))
	}
	if payload, ok := data["payload"].(map[string]interface{}); ok {
		h.processPayload(nodeID, msgType, payload)
	}
}

func (h *PrometheusHook) processPayload(nodeID, msgType string, payload map[string]interface{}) {
	switch msgType {
	case "telemetry":
		h.processTelemetry(nodeID, payload)
	case "nodeinfo":
		h.processNodeInfo(nodeID, payload)
	}
}

func (h *PrometheusHook) processTelemetry(nodeID string, payload map[string]interface{}) {
	if val, ok := payload["battery_level"].(float64); ok {
		h.batteryLevel.WithLabelValues(nodeID).Set(val)
	}
	if val, ok := payload["voltage"].(float64); ok {
		h.voltage.WithLabelValues(nodeID).Set(roundFloat(val, 2))
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
	if val, ok := payload["temperature"].(float64); ok {
		h.temperature.WithLabelValues(nodeID).Set(roundFloat(val, 1))
	}
	if val, ok := payload["relative_humidity"].(float64); ok {
		h.humidity.WithLabelValues(nodeID).Set(roundFloat(val, 1))
	}
	if val, ok := payload["barometric_pressure"].(float64); ok {
		h.pressure.WithLabelValues(nodeID).Set(roundFloat(val, 1))
	}
}

func (h *PrometheusHook) processNodeInfo(nodeID string, payload map[string]interface{}) {
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

func (h *PrometheusHook) startServer() {
	if !h.Config.Prometheus.Enabled {
		return
	}
	var addr string
	if h.Config.Prometheus.Host == "::" {
		addr = "[::]:" + strconv.Itoa(h.Config.Prometheus.Port)
	} else {
		addr = h.Config.Prometheus.Host + ":" + strconv.Itoa(h.Config.Prometheus.Port)
	}
	http.Handle("/metrics", promhttp.HandlerFor(h.registry, promhttp.HandlerOpts{}))
	http.HandleFunc("/health", h.healthHandler)
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
	logger.Info().Str("component", "prometheus").Str("address", addr).Msg("server started")
	go func() {
		server := &http.Server{
			Addr:         addr,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
		if err := server.ListenAndServe(); err != nil {
			logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
			logger.Error().Err(err).Str("component", "prometheus").Msg("server error")
		}
	}()
}

func getUint32(data map[string]interface{}, key string) uint32 {
	if val, ok := data[key].(float64); ok {
		return uint32(val)
	}
	return 0
}

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func (h *PrometheusHook) healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "meshtastic-exporter"})
}

func roundFloat(val float64, precision int) float64 {
	ratio := 1.0
	for i := 0; i < precision; i++ {
		ratio *= 10
	}
	return float64(int(val*ratio+0.5)) / ratio
}

func (h *PrometheusHook) startCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			h.cleanupStaleMetrics()
		}
	}()
}

func (h *PrometheusHook) cleanupStaleMetrics() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	ttl, err := time.ParseDuration(h.Config.Prometheus.MetricsTTL)
	if err != nil {
		ttl = 30 * time.Minute
	}

	staleThreshold := time.Now().Add(-ttl)
	for nodeID, lastSeen := range h.nodeMetrics {
		if lastSeen.Before(staleThreshold) {
			h.deleteNodeMetrics(nodeID)
			delete(h.nodeMetrics, nodeID)
			logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
			logger.Info().Str("component", "cleanup").Str("node_id", nodeID).Dur("ttl", ttl).Msg("removed stale metrics")
		}
	}
}

func (h *PrometheusHook) deleteNodeMetrics(nodeID string) {
	h.batteryLevel.DeleteLabelValues(nodeID)
	h.voltage.DeleteLabelValues(nodeID)
	h.channelUtil.DeleteLabelValues(nodeID)
	h.airUtilTx.DeleteLabelValues(nodeID)
	h.uptime.DeleteLabelValues(nodeID)
	h.temperature.DeleteLabelValues(nodeID)
	h.humidity.DeleteLabelValues(nodeID)
	h.pressure.DeleteLabelValues(nodeID)
	h.nodeLastSeen.DeleteLabelValues(nodeID)
}

func (h *PrometheusHook) loadState() {
	if !h.Config.State.Enabled || h.Config.State.File == "" {
		return
	}
	data, err := os.ReadFile(h.Config.State.File)
	if err != nil {
		return
	}
	var state exporter.State
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}
	nodeCount := 0
	for nodeID, node := range state.Nodes {
		if node.BatteryLevel > 0 {
			h.batteryLevel.WithLabelValues(nodeID).Set(node.BatteryLevel)
		}
		if node.Voltage > 0 {
			h.voltage.WithLabelValues(nodeID).Set(node.Voltage)
		}
		if node.Temperature != 0 {
			h.temperature.WithLabelValues(nodeID).Set(node.Temperature)
		}
		if node.Humidity > 0 {
			h.humidity.WithLabelValues(nodeID).Set(node.Humidity)
		}
		if node.Pressure > 0 {
			h.pressure.WithLabelValues(nodeID).Set(node.Pressure)
		}
		if node.Uptime > 0 {
			h.uptime.WithLabelValues(nodeID).Set(node.Uptime)
		}
		nodeCount++
	}
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
	logger.Info().Str("component", "state").Int("nodes", nodeCount).Str("file", h.Config.State.File).Msg("loaded metrics")
}

func (h *PrometheusHook) SaveState() {
	if !h.Config.State.Enabled || h.Config.State.File == "" {
		return
	}
	state := exporter.State{Nodes: h.extractMetricValues(), Timestamp: time.Now()}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
		logger.Error().Err(err).Str("component", "state").Msg("failed to marshal state")
		return
	}
	if err := os.WriteFile(h.Config.State.File, data, 0600); err != nil {
		logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
		logger.Error().Err(err).Str("component", "state").Msg("failed to save state")
	}
}

func (h *PrometheusHook) SaveStateOnShutdown() {
	if h.Config.State.Enabled {
		h.SaveState()
		logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
		logger.Info().Str("component", "state").Msg("saved on shutdown")
	}
}

func (h *PrometheusHook) extractMetricValues() map[string]exporter.NodeState {
	nodes := make(map[string]exporter.NodeState)
	extractFromMetric := func(vec *prometheus.GaugeVec, setValue func(*exporter.NodeState, float64)) {
		metricChan := make(chan prometheus.Metric, 100)
		go func() {
			vec.Collect(metricChan)
			close(metricChan)
		}()
		for metric := range metricChan {
			dtoMetric := &dto.Metric{}
			if err := metric.Write(dtoMetric); err != nil {
				continue
			}
			nodeID := ""
			for _, label := range dtoMetric.GetLabel() {
				if label.GetName() == "node_id" {
					nodeID = label.GetValue()
					break
				}
			}
			if nodeID != "" && dtoMetric.GetGauge() != nil {
				node := nodes[nodeID]
				setValue(&node, dtoMetric.GetGauge().GetValue())
				nodes[nodeID] = node
			}
		}
	}
	extractFromMetric(h.batteryLevel, func(n *exporter.NodeState, v float64) { n.BatteryLevel = v })
	extractFromMetric(h.voltage, func(n *exporter.NodeState, v float64) { n.Voltage = v })
	extractFromMetric(h.temperature, func(n *exporter.NodeState, v float64) { n.Temperature = v })
	extractFromMetric(h.humidity, func(n *exporter.NodeState, v float64) { n.Humidity = v })
	extractFromMetric(h.pressure, func(n *exporter.NodeState, v float64) { n.Pressure = v })
	extractFromMetric(h.uptime, func(n *exporter.NodeState, v float64) { n.Uptime = v })
	return nodes
}
