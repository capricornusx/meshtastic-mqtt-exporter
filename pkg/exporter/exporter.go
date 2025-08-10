package exporter

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

type MQTTUser struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Config struct {
	MQTT struct {
		Host           string     `yaml:"host"`
		Port           int        `yaml:"port"`
		Username       string     `yaml:"username"`
		Password       string     `yaml:"password"`
		TLS            bool       `yaml:"tls"`
		AllowAnonymous bool       `yaml:"allow_anonymous"`
		Users          []MQTTUser `yaml:"users,omitempty"`
		Broker         struct {
			MaxInflight     int  `yaml:"max_inflight"`
			MaxQueued       int  `yaml:"max_queued"`
			RetainAvailable bool `yaml:"retain_available"`
			MaxPacketSize   int  `yaml:"max_packet_size"`
			KeepAlive       int  `yaml:"keep_alive"`
		} `yaml:"broker,omitempty"`
	} `yaml:"mqtt"`
	Prometheus struct {
		Enabled    bool   `yaml:"enabled"`
		Port       int    `yaml:"port"`
		Host       string `yaml:"host"`
		MetricsTTL string `yaml:"metrics_ttl"`
	} `yaml:"prometheus"`
	State struct {
		Enabled bool   `yaml:"enabled"`
		File    string `yaml:"file"`
	} `yaml:"state"`
}

type NodeState struct {
	BatteryLevel float64 `json:"battery_level,omitempty"`
	Voltage      float64 `json:"voltage,omitempty"`
	ChannelUtil  float64 `json:"channel_util,omitempty"`
	AirUtilTx    float64 `json:"air_util_tx,omitempty"`
	Uptime       float64 `json:"uptime,omitempty"`
	Temperature  float64 `json:"temperature,omitempty"`
	Humidity     float64 `json:"humidity,omitempty"`
	Pressure     float64 `json:"pressure,omitempty"`
	Hardware     float64 `json:"hardware,omitempty"`
	Role         float64 `json:"role,omitempty"`
	Longname     string  `json:"longname,omitempty"`
	Shortname    string  `json:"shortname,omitempty"`
}

type State struct {
	Nodes     map[string]NodeState `json:"nodes"`
	Timestamp time.Time            `json:"timestamp"`
}

type MeshtasticExporter struct {
	config    Config
	client    mqtt.Client
	nodeNames map[uint32]string

	stateFile string

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

func LoadConfig(filename string) (Config, error) {
	const localhost = "localhost"

	var config Config
	config.MQTT.Host = localhost
	config.MQTT.Port = 1883
	config.MQTT.Broker.MaxInflight = 50
	config.MQTT.Broker.MaxQueued = 1000
	config.MQTT.Broker.RetainAvailable = true
	config.MQTT.Broker.MaxPacketSize = 131072
	config.MQTT.Broker.KeepAlive = 60
	config.Prometheus.Enabled = true
	config.Prometheus.Port = 8000
	config.Prometheus.Host = localhost
	config.Prometheus.MetricsTTL = "10m"
	config.State.Enabled = false
	config.State.File = "meshtastic_state.json"

	data, err := os.ReadFile(filename)
	if err != nil {
		logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
		logger.Warn().Err(err).Str("component", "config").Msg("config file not found, using defaults")
		return config, nil
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse config: %w", err)
	}
	return config, nil
}

func New(config Config) *MeshtasticExporter {
	return &MeshtasticExporter{
		config:    config,
		nodeNames: make(map[uint32]string),
		stateFile: config.State.File,
	}
}

func (e *MeshtasticExporter) Init() {
	e.setupMetrics()
	if e.config.State.Enabled {
		e.loadState()
	}
}

func (e *MeshtasticExporter) setupMetrics() {
	e.messageCounter = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "meshtastic_messages_total", Help: "Total messages by type"}, []string{"type", "from_node"})
	e.rssi = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_rssi_dbm", Help: "RSSI"}, []string{"from_node", "to_node"})
	e.snr = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_snr_db", Help: "SNR"}, []string{"from_node", "to_node"})
	e.batteryLevel = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_battery_level_percent", Help: "Battery level"}, []string{"node_id"})
	e.voltage = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_voltage_volts", Help: "Battery voltage"}, []string{"node_id"})
	e.channelUtil = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_channel_utilization_percent", Help: "Channel utilization"}, []string{"node_id"})
	e.airUtilTx = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_air_util_tx_percent", Help: "Air utilization TX"}, []string{"node_id"})
	e.uptime = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_uptime_seconds", Help: "Uptime"}, []string{"node_id"})
	e.temperature = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_temperature_celsius", Help: "Temperature"}, []string{"node_id"})
	e.humidity = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_humidity_percent", Help: "Humidity"}, []string{"node_id"})
	e.pressure = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_pressure_hpa", Help: "Pressure"}, []string{"node_id"})
	e.nodeLastSeen = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_node_last_seen_timestamp", Help: "Last seen timestamp"}, []string{"node_id"})
	e.mqttUp = prometheus.NewGauge(prometheus.GaugeOpts{Name: "meshtastic_mqtt_up", Help: "MQTT connection status"})
	e.nodeHardware = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "meshtastic_node_info", Help: "Node information"}, []string{"node_id", "longname", "shortname", "hardware", "role"})
}

func (e *MeshtasticExporter) setupMQTT() error {
	opts := mqtt.NewClientOptions()
	var broker string
	if e.config.MQTT.Host == "::" {
		if e.config.MQTT.TLS {
			broker = fmt.Sprintf("ssl://[::1]:%d", e.config.MQTT.Port)
		} else {
			broker = fmt.Sprintf("tcp://[::1]:%d", e.config.MQTT.Port)
		}
	} else {
		if e.config.MQTT.TLS {
			broker = fmt.Sprintf("ssl://%s:%d", e.config.MQTT.Host, e.config.MQTT.Port)
		} else {
			broker = fmt.Sprintf("tcp://%s:%d", e.config.MQTT.Host, e.config.MQTT.Port)
		}
	}
	opts.SetTLSConfig(&tls.Config{
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
	})
	opts.AddBroker(broker)
	opts.SetClientID("meshtastic-exporter")
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	if e.config.MQTT.Username != "" {
		opts.SetUsername(e.config.MQTT.Username)
		opts.SetPassword(e.config.MQTT.Password)
	}
	opts.SetOnConnectHandler(e.onConnect)
	opts.SetConnectionLostHandler(e.onConnectionLost)
	e.client = mqtt.NewClient(opts)
	if token := e.client.Connect(); token.Wait() && token.Error() != nil {
		e.mqttUp.Set(0)
		return fmt.Errorf("failed to connect to MQTT: %w", token.Error())
	}
	e.mqttUp.Set(1)
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
	logger.Info().Str("component", "mqtt").Msg("connected to broker")
	return nil
}

func (e *MeshtasticExporter) onConnect(client mqtt.Client) {
	topics := []string{"msh/+/+/json/+/+", "msh/2/json/+/+"}
	for _, topic := range topics {
		if token := client.Subscribe(topic, 0, e.messageHandler); token.Wait() && token.Error() != nil {
			logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
			logger.Error().Err(token.Error()).Str("component", "mqtt").Str("topic", topic).Msg("failed to subscribe")
		} else {
			logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
			logger.Info().Str("component", "mqtt").Str("topic", topic).Msg("subscribed")
		}
	}
}

func (e *MeshtasticExporter) onConnectionLost(_ mqtt.Client, err error) {
	e.mqttUp.Set(0)
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
	logger.Error().Err(err).Str("component", "mqtt").Msg("connection lost")
}

func (e *MeshtasticExporter) messageHandler(_ mqtt.Client, msg mqtt.Message) {
	var data map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
		logger.Error().Err(err).Str("component", "mqtt").Msg("failed to parse json")
		return
	}
	e.processMessage(data)
}

func (e *MeshtasticExporter) processMessage(data map[string]interface{}) {
	fromNode := getUint32(data, "from")
	toNode := getUint32(data, "to")
	msgType := getString(data, "type")
	if fromNode == 0 {
		return
	}
	nodeID := strconv.FormatUint(uint64(fromNode), 10)
	e.nodeLastSeen.WithLabelValues(nodeID).SetToCurrentTime()
	e.messageCounter.WithLabelValues(msgType, nodeID).Inc()
	if rssi, ok := data["rssi"].(float64); ok {
		e.rssi.WithLabelValues(nodeID, strconv.FormatUint(uint64(toNode), 10)).Set(rssi)
	}
	if snr, ok := data["snr"].(float64); ok {
		e.snr.WithLabelValues(nodeID, strconv.FormatUint(uint64(toNode), 10)).Set(roundFloat(snr, 1))
	}
	if payload, ok := data["payload"].(map[string]interface{}); ok {
		e.processPayload(nodeID, msgType, payload)
	}
}

func (e *MeshtasticExporter) processPayload(nodeID, msgType string, payload map[string]interface{}) {
	switch msgType {
	case "telemetry":
		e.processTelemetry(nodeID, payload)
	case "nodeinfo":
		e.processNodeInfo(nodeID, payload)
	}
}

func (e *MeshtasticExporter) processTelemetry(nodeID string, payload map[string]interface{}) {
	if val, ok := payload["battery_level"].(float64); ok {
		e.batteryLevel.WithLabelValues(nodeID).Set(val)
	}
	if val, ok := payload["voltage"].(float64); ok {
		e.voltage.WithLabelValues(nodeID).Set(roundFloat(val, 2))
	}
	if val, ok := payload["channel_utilization"].(float64); ok {
		e.channelUtil.WithLabelValues(nodeID).Set(roundFloat(val, 2))
	}
	if val, ok := payload["air_util_tx"].(float64); ok {
		e.airUtilTx.WithLabelValues(nodeID).Set(roundFloat(val, 2))
	}
	if val, ok := payload["uptime_seconds"].(float64); ok {
		e.uptime.WithLabelValues(nodeID).Set(val)
	}
	if val, ok := payload["temperature"].(float64); ok {
		e.temperature.WithLabelValues(nodeID).Set(roundFloat(val, 1))
	}
	if val, ok := payload["relative_humidity"].(float64); ok {
		e.humidity.WithLabelValues(nodeID).Set(roundFloat(val, 1))
	}
	if val, ok := payload["barometric_pressure"].(float64); ok {
		e.pressure.WithLabelValues(nodeID).Set(roundFloat(val, 1))
	}
}

func (e *MeshtasticExporter) processNodeInfo(nodeID string, payload map[string]interface{}) {
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
	e.nodeHardware.WithLabelValues(nodeID, longname, shortname, hardware, role).Set(1)
}

func (e *MeshtasticExporter) startPrometheusServer() {
	if !e.config.Prometheus.Enabled {
		return
	}
	registry := prometheus.NewRegistry()
	registry.MustRegister(e.messageCounter, e.rssi, e.snr, e.batteryLevel, e.voltage, e.channelUtil, e.airUtilTx, e.uptime, e.temperature, e.humidity, e.pressure, e.nodeLastSeen, e.mqttUp, e.nodeHardware)
	var addr string
	if e.config.Prometheus.Host == "::" {
		addr = fmt.Sprintf("[::]:%d", e.config.Prometheus.Port)
	} else {
		addr = fmt.Sprintf("%s:%d", e.config.Prometheus.Host, e.config.Prometheus.Port)
	}
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.HandleFunc("/health", e.healthHandler)
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
			logger.Fatal().Err(err).Str("component", "prometheus").Msg("failed to start server")
		}
	}()
}

func (e *MeshtasticExporter) saveState() {
	if !e.config.State.Enabled || e.stateFile == "" {
		return
	}
	state := State{Nodes: e.extractMetricValues(), Timestamp: time.Now()}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
		logger.Error().Err(err).Str("component", "state").Msg("failed to marshal")
		return
	}
	if err := os.WriteFile(e.stateFile, data, 0600); err != nil {
		logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
		logger.Error().Err(err).Str("component", "state").Msg("failed to save")
	}
}

func (e *MeshtasticExporter) loadState() {
	if !e.config.State.Enabled || e.stateFile == "" {
		return
	}
	data, err := os.ReadFile(e.stateFile)
	if err != nil {
		return
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}
	nodeCount := 0
	for nodeID, node := range state.Nodes {
		if node.BatteryLevel > 0 {
			e.batteryLevel.WithLabelValues(nodeID).Set(node.BatteryLevel)
		}
		if node.Voltage > 0 {
			e.voltage.WithLabelValues(nodeID).Set(node.Voltage)
		}
		if node.Temperature != 0 {
			e.temperature.WithLabelValues(nodeID).Set(node.Temperature)
		}
		if node.Uptime > 0 {
			e.uptime.WithLabelValues(nodeID).Set(node.Uptime)
		}
		nodeCount++
	}
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
	logger.Info().Str("component", "state").Int("nodes", nodeCount).Str("file", e.stateFile).Msg("loaded metrics")
}

func (e *MeshtasticExporter) extractMetricValues() map[string]NodeState {
	nodes := make(map[string]NodeState)
	extractFromMetric := func(vec *prometheus.GaugeVec, setValue func(*NodeState, float64)) {
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
	extractFromMetric(e.batteryLevel, func(n *NodeState, v float64) { n.BatteryLevel = v })
	extractFromMetric(e.voltage, func(n *NodeState, v float64) { n.Voltage = v })
	extractFromMetric(e.temperature, func(n *NodeState, v float64) { n.Temperature = v })
	extractFromMetric(e.uptime, func(n *NodeState, v float64) { n.Uptime = v })
	return nodes
}

func (e *MeshtasticExporter) Run() error {
	if err := e.setupMQTT(); err != nil {
		return err
	}
	e.startPrometheusServer()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
	logger.Info().Str("component", "standalone").Msg("shutting down")
	if e.config.State.Enabled {
		e.saveState()
	}
	e.client.Disconnect(250)
	return nil
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

func (e *MeshtasticExporter) healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := "ok"
	if e.client != nil && !e.client.IsConnected() {
		status = "degraded"
	}
	json.NewEncoder(w).Encode(map[string]string{"status": status, "service": "meshtastic-exporter"})
}

func roundFloat(val float64, precision int) float64 {
	ratio := 1.0
	for i := 0; i < precision; i++ {
		ratio *= 10
	}
	return float64(int(val*ratio+0.5)) / ratio
}
