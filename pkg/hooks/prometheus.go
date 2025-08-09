package hooks

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"meshtastic-exporter/pkg/exporter"
)

type PrometheusHook struct {
	mqtt.HookBase
	config   exporter.Config
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
}

func NewPrometheusHook(config exporter.Config) *PrometheusHook {
	h := &PrometheusHook{
		config:   config,
		registry: prometheus.NewRegistry(),
	}
	h.setupMetrics()
	h.mqttUp.Set(1) // MQTT server is up when hook initializes
	h.startServer()
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
	if !h.config.Prometheus.Enabled {
		return
	}
	addr := h.config.Prometheus.Host + ":" + strconv.Itoa(h.config.Prometheus.Port)
	http.Handle("/metrics", promhttp.HandlerFor(h.registry, promhttp.HandlerOpts{}))
	http.HandleFunc("/health", h.healthHandler)
	log.Printf("Starting Prometheus server on %s", addr)
	go func() {
		server := &http.Server{
			Addr:         addr,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Prometheus server error: %v", err)
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
