package infrastructure

import (
	"encoding/json"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/rs/zerolog/log"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/version"
)

type PrometheusCollector struct {
	registry *prometheus.Registry

	messageCounter *prometheus.CounterVec
	batteryLevel   *prometheus.GaugeVec
	voltage        *prometheus.GaugeVec
	temperature    *prometheus.GaugeVec
	humidity       *prometheus.GaugeVec
	pressure       *prometheus.GaugeVec
	channelUtil    *prometheus.GaugeVec
	airUtilTx      *prometheus.GaugeVec
	uptime         *prometheus.GaugeVec
	rssi           *prometheus.GaugeVec
	snr            *prometheus.GaugeVec
	nodeLastSeen   *prometheus.GaugeVec
	nodeHardware   *prometheus.GaugeVec
	serviceInfo    *prometheus.GaugeVec
}

func NewPrometheusCollector() *PrometheusCollector {
	return NewPrometheusCollectorWithMode("hook")
}

func NewPrometheusCollectorWithMode(mode string) *PrometheusCollector {
	registry := prometheus.NewRegistry()

	collector := &PrometheusCollector{
		registry: registry,
	}

	collector.setupMetrics()
	collector.setupServiceInfo(mode)
	return collector
}

func (c *PrometheusCollector) setupMetrics() {
	c.messageCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: domain.MetricMessagesTotal, Help: "Total messages by type"},
		[]string{"type", "from_node"})

	c.batteryLevel = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: domain.MetricBatteryLevel, Help: "Battery level"},
		[]string{"node_id"})

	c.voltage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: domain.MetricVoltage, Help: "Battery voltage"},
		[]string{"node_id"})

	c.temperature = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: domain.MetricTemperature, Help: "Temperature"},
		[]string{"node_id"})

	c.humidity = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: domain.MetricHumidity, Help: "Humidity"},
		[]string{"node_id"})

	c.pressure = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: domain.MetricPressure, Help: "Pressure"},
		[]string{"node_id"})

	c.channelUtil = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: domain.MetricChannelUtil, Help: "Channel utilization"},
		[]string{"node_id"})

	c.airUtilTx = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: domain.MetricAirUtilTx, Help: "Air utilization TX"},
		[]string{"node_id"})

	c.uptime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: domain.MetricUptime, Help: "Uptime"},
		[]string{"node_id"})

	c.rssi = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: domain.MetricRSSI, Help: "RSSI signal strength"},
		[]string{"node_id"})

	c.snr = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: domain.MetricSNR, Help: "Signal-to-noise ratio"},
		[]string{"node_id"})

	c.nodeLastSeen = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: domain.MetricNodeLastSeen, Help: "Last seen timestamp"},
		[]string{"node_id"})

	c.nodeHardware = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: domain.MetricNodeInfo, Help: "Node information"},
		[]string{"node_id", "longname", "shortname", "hardware", "role"})

	c.registry.MustRegister(
		c.messageCounter, c.batteryLevel, c.voltage, c.temperature,
		c.humidity, c.pressure, c.channelUtil, c.airUtilTx,
		c.uptime, c.rssi, c.snr, c.nodeLastSeen, c.nodeHardware,
	)
}

func (c *PrometheusCollector) setupServiceInfo(mode string) {
	c.serviceInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: domain.MetricExporterInfo, Help: "Service information"},
		[]string{"version", "mode", "git_commit", "build_date"})
	c.registry.MustRegister(c.serviceInfo)
	c.updateServiceInfo(mode)
}

func (c *PrometheusCollector) updateServiceInfo(mode string) {
	version, gitCommit, buildDate := getVersionInfo()
	c.serviceInfo.WithLabelValues(version, mode, gitCommit, buildDate).Set(1)
}

func getVersionInfo() (string, string, string) {
	return version.GetBuildInfo()
}

func (c *PrometheusCollector) CollectTelemetry(data domain.TelemetryData) error {
	c.nodeLastSeen.WithLabelValues(data.NodeID).SetToCurrentTime()
	c.messageCounter.WithLabelValues(domain.MessageTypeTelemetry, data.NodeID).Inc()

	c.setTelemetryMetrics(data)
	return nil
}

func (c *PrometheusCollector) setTelemetryMetrics(data domain.TelemetryData) {
	c.setBasicMetrics(data)
	c.setEnvironmentalMetrics(data)
	c.setNetworkMetrics(data)
}

func (c *PrometheusCollector) setBasicMetrics(data domain.TelemetryData) {
	if data.BatteryLevel != nil {
		c.batteryLevel.WithLabelValues(data.NodeID).Set(*data.BatteryLevel)
	}
	if data.Voltage != nil {
		c.voltage.WithLabelValues(data.NodeID).Set(*data.Voltage)
	}
	if data.UptimeSeconds != nil {
		c.uptime.WithLabelValues(data.NodeID).Set(*data.UptimeSeconds)
	}
}

func (c *PrometheusCollector) setEnvironmentalMetrics(data domain.TelemetryData) {
	if data.Temperature != nil {
		c.temperature.WithLabelValues(data.NodeID).Set(*data.Temperature)
	}
	if data.RelativeHumidity != nil {
		c.humidity.WithLabelValues(data.NodeID).Set(*data.RelativeHumidity)
	}
	if data.BarometricPressure != nil {
		c.pressure.WithLabelValues(data.NodeID).Set(*data.BarometricPressure)
	}
}

func (c *PrometheusCollector) setNetworkMetrics(data domain.TelemetryData) {
	if data.ChannelUtilization != nil {
		c.channelUtil.WithLabelValues(data.NodeID).Set(*data.ChannelUtilization)
	}
	if data.AirUtilTx != nil {
		c.airUtilTx.WithLabelValues(data.NodeID).Set(*data.AirUtilTx)
	}
	if data.RSSI != nil {
		c.rssi.WithLabelValues(data.NodeID).Set(*data.RSSI)
	}
	if data.SNR != nil {
		c.snr.WithLabelValues(data.NodeID).Set(*data.SNR)
	}
}

func (c *PrometheusCollector) CollectNodeInfo(info domain.NodeInfo) error {
	c.nodeLastSeen.WithLabelValues(info.NodeID).SetToCurrentTime()
	c.messageCounter.WithLabelValues(domain.MessageTypeNodeInfo, info.NodeID).Inc()
	c.nodeHardware.WithLabelValues(info.NodeID, info.LongName, info.ShortName, info.Hardware, info.Role).Set(1)
	return nil
}

func (c *PrometheusCollector) GetRegistry() *prometheus.Registry {
	return c.registry
}

func (c *PrometheusCollector) SaveState(filename string) error {
	if filename == "" {
		return nil
	}

	metricFamilies, err := c.registry.Gather()
	if err != nil {
		return err
	}

	nodeMetrics := c.extractNodeMetrics(metricFamilies)
	state := c.buildStateSnapshot(nodeMetrics)

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, domain.StateFilePermissions)
}

func (c *PrometheusCollector) extractNodeMetrics(metricFamilies []*dto.MetricFamily) map[string]domain.MetricState {
	nodeMetrics := make(map[string]domain.MetricState)

	for _, mf := range metricFamilies {
		for _, metric := range mf.GetMetric() {
			nodeID, labels := c.extractLabels(metric)
			if nodeID == "" {
				continue
			}

			if _, exists := nodeMetrics[nodeID]; !exists {
				nodeMetrics[nodeID] = domain.MetricState{
					NodeID:    nodeID,
					Timestamp: time.Now().Unix(),
					Metrics:   make(map[string]float64),
					Labels:    labels,
				}
			}

			value := c.extractMetricValue(metric)
			nodeState := nodeMetrics[nodeID]
			nodeState.Metrics[mf.GetName()] = value
			nodeMetrics[nodeID] = nodeState
		}
	}

	return nodeMetrics
}

func (c *PrometheusCollector) extractLabels(metric *dto.Metric) (string, map[string]string) {
	var nodeID string
	labels := make(map[string]string)

	for _, label := range metric.GetLabel() {
		labels[label.GetName()] = label.GetValue()
		if label.GetName() == "node_id" {
			nodeID = label.GetValue()
		}
	}

	return nodeID, labels
}

func (c *PrometheusCollector) extractMetricValue(metric *dto.Metric) float64 {
	if metric.GetGauge() != nil {
		return metric.GetGauge().GetValue()
	}
	if metric.GetCounter() != nil {
		return metric.GetCounter().GetValue()
	}
	return 0
}

func (c *PrometheusCollector) buildStateSnapshot(nodeMetrics map[string]domain.MetricState) domain.StateSnapshot {
	state := domain.StateSnapshot{
		Version:   "1.0",
		Timestamp: time.Now().Unix(),
		Nodes:     make([]domain.MetricState, 0, len(nodeMetrics)),
	}

	for _, nodeState := range nodeMetrics {
		state.Nodes = append(state.Nodes, nodeState)
	}

	return state
}

func (c *PrometheusCollector) LoadState(filename string) error {
	if filename == "" {
		return nil
	}

	state, err := c.readStateFile(filename)
	if err != nil {
		return err
	}
	if state == nil {
		return nil
	}

	log.Info().Int("nodes", len(state.Nodes)).Str("version", state.Version).Msg("restoring metrics state")
	c.restoreMetrics(state.Nodes)
	return nil
}

func (c *PrometheusCollector) readStateFile(filename string) (*domain.StateSnapshot, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info().Str("file", filename).Msg("state file not found, starting fresh")
			return nil, nil
		}
		return nil, err
	}

	var state domain.StateSnapshot
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

func (c *PrometheusCollector) restoreMetrics(nodes []domain.MetricState) {
	for _, nodeState := range nodes {
		for metricName, value := range nodeState.Metrics {
			c.restoreMetric(metricName, value, nodeState)
		}
	}
}

func (c *PrometheusCollector) restoreMetric(metricName string, value float64, nodeState domain.MetricState) {
	metricMap := map[string]*prometheus.GaugeVec{
		domain.MetricBatteryLevel: c.batteryLevel,
		domain.MetricVoltage:      c.voltage,
		domain.MetricTemperature:  c.temperature,
		domain.MetricHumidity:     c.humidity,
		domain.MetricPressure:     c.pressure,
		domain.MetricChannelUtil:  c.channelUtil,
		domain.MetricAirUtilTx:    c.airUtilTx,
		domain.MetricUptime:       c.uptime,
		domain.MetricRSSI:         c.rssi,
		domain.MetricSNR:          c.snr,
		domain.MetricNodeLastSeen: c.nodeLastSeen,
	}

	if gauge, exists := metricMap[metricName]; exists {
		gauge.WithLabelValues(nodeState.NodeID).Set(value)
	} else if metricName == domain.MetricNodeInfo {
		longname := nodeState.Labels["longname"]
		shortname := nodeState.Labels["shortname"]
		hardware := nodeState.Labels["hardware"]
		role := nodeState.Labels["role"]
		c.nodeHardware.WithLabelValues(nodeState.NodeID, longname, shortname, hardware, role).Set(value)
	}
}
