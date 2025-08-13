package infrastructure

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/logger"
	"meshtastic-exporter/pkg/version"
)

const (
	unknownValue              = "unknown"
	metricsCollectorComponent = "metrics-collector"
	minCleanupInterval        = 500 * time.Millisecond
	maxCleanupInterval        = 30 * time.Second
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

	metricTimestamps map[string]map[string]time.Time // nodeID -> metricName -> timestamp
	metricsTTL       time.Duration
	cleanupCancel    context.CancelFunc
	mu               sync.RWMutex
}

func NewPrometheusCollector() *PrometheusCollector {
	return NewPrometheusCollectorWithTTL("hook", domain.DefaultMetricsTTL)
}

func NewPrometheusCollectorWithTTL(mode string, ttl time.Duration) *PrometheusCollector {
	return NewPrometheusCollectorWithConfig(mode, ttl)
}

func NewPrometheusCollectorWithMode(mode string) *PrometheusCollector {
	return NewPrometheusCollectorWithConfig(mode, domain.DefaultMetricsTTL)
}

func NewPrometheusCollectorWithConfig(mode string, ttl time.Duration) *PrometheusCollector {
	registry := prometheus.NewRegistry()

	if ttl <= 0 {
		ttl = domain.DefaultMetricsTTL
	}

	collector := &PrometheusCollector{
		registry:         registry,
		metricTimestamps: make(map[string]map[string]time.Time),
		metricsTTL:       ttl,
	}

	collector.setupMetrics()
	collector.setupServiceInfo(mode)
	if ttl > 0 {
		go collector.startMetricsTTLCleanup()
	}
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
	c.UpdateNodeLastSeen(data.NodeID, time.Now())
	c.UpdateMessageCounter(data.NodeID, domain.MessageTypeTelemetry)
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
		c.updateMetricTimestamp(data.NodeID, domain.MetricTemperature)
	}
	if data.RelativeHumidity != nil {
		c.humidity.WithLabelValues(data.NodeID).Set(*data.RelativeHumidity)
		c.updateMetricTimestamp(data.NodeID, domain.MetricHumidity)
	}
	if data.BarometricPressure != nil {
		c.pressure.WithLabelValues(data.NodeID).Set(*data.BarometricPressure)
		c.updateMetricTimestamp(data.NodeID, domain.MetricPressure)
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
	c.UpdateNodeLastSeen(info.NodeID, time.Now())
	c.UpdateMessageCounter(info.NodeID, domain.MessageTypeNodeInfo)
	c.nodeHardware.WithLabelValues(info.NodeID, info.LongName, info.ShortName, info.Hardware, info.Role).Set(1)
	return nil
}

func (c *PrometheusCollector) CollectTextMessage(msg domain.TextMessage) error {
	c.UpdateNodeLastSeen(msg.NodeID, time.Now())
	c.UpdateMessageCounter(msg.NodeID, domain.MessageTypeText)
	return nil
}

func (c *PrometheusCollector) CollectPosition(pos domain.Position) error {
	c.UpdateNodeLastSeen(pos.NodeID, time.Now())
	c.UpdateMessageCounter(pos.NodeID, domain.MessageTypePosition)
	return nil
}

func (c *PrometheusCollector) CollectWaypoint(wp domain.Waypoint) error {
	c.UpdateNodeLastSeen(wp.NodeID, time.Now())
	c.UpdateMessageCounter(wp.NodeID, domain.MessageTypeWaypoint)
	return nil
}

func (c *PrometheusCollector) CollectNeighborInfo(ni domain.NeighborInfo) error {
	c.UpdateNodeLastSeen(ni.NodeID, time.Now())
	c.UpdateMessageCounter(ni.NodeID, domain.MessageTypeNeighborInfo)
	return nil
}

func (c *PrometheusCollector) UpdateNodeLastSeen(nodeID string, timestamp time.Time) {
	c.nodeLastSeen.WithLabelValues(nodeID).Set(float64(timestamp.Unix()))
}

func (c *PrometheusCollector) UpdateMessageCounter(nodeID string, messageType string) {
	c.messageCounter.WithLabelValues(messageType, nodeID).Inc()
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
	if len(nodeMetrics) == 0 {
		return nil
	}

	state := c.buildStateSnapshot(nodeMetrics)

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	log := logger.ComponentLogger(metricsCollectorComponent)
	log.Info().Int("nodes", len(state.Nodes)).Str("file", filename).Msg("saving metrics state")
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

			c.ensureNodeState(nodeMetrics, nodeID)
			c.updateNodeMetric(nodeMetrics, nodeID, mf.GetName(), metric, labels)
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

func (c *PrometheusCollector) ensureNodeState(nodeMetrics map[string]domain.MetricState, nodeID string) {
	if _, exists := nodeMetrics[nodeID]; !exists {
		nodeMetrics[nodeID] = domain.MetricState{
			NodeID:    nodeID,
			Timestamp: time.Now().Unix(),
			Metrics:   make(map[string]float64),
			Labels:    make(map[string]string),
		}
	}
}

func (c *PrometheusCollector) updateNodeMetric(nodeMetrics map[string]domain.MetricState, nodeID, metricName string, metric *dto.Metric, labels map[string]string) {
	nodeState := nodeMetrics[nodeID]
	nodeState.Metrics[metricName] = c.extractMetricValue(metric)
	c.updateNodeLabels(&nodeState, metricName, nodeID, labels)
	nodeMetrics[nodeID] = nodeState
}

func (c *PrometheusCollector) updateNodeLabels(nodeState *domain.MetricState, metricName, nodeID string, labels map[string]string) {
	if metricName == domain.MetricNodeInfo {
		for k, v := range labels {
			nodeState.Labels[k] = v
		}
	} else {
		nodeState.Labels["node_id"] = nodeID
	}
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

	log := logger.ComponentLogger(metricsCollectorComponent)
	log.Info().Int("nodes", len(state.Nodes)).Str("version", state.Version).Str("file", filename).Msg("restoring metrics state")
	c.restoreMetrics(state.Nodes)
	return nil
}

func (c *PrometheusCollector) readStateFile(filename string) (*domain.StateSnapshot, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			log := logger.ComponentLogger(metricsCollectorComponent)
			log.Debug().Str("file", filename).Msg("state file not found, starting fresh")
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
		if longname == "" {
			longname = unknownValue
		}
		shortname := nodeState.Labels["shortname"]
		if shortname == "" {
			shortname = unknownValue
		}
		hardware := nodeState.Labels["hardware"]
		if hardware == "" {
			hardware = unknownValue
		}
		role := nodeState.Labels["role"]
		if role == "" {
			role = unknownValue
		}
		c.nodeHardware.WithLabelValues(nodeState.NodeID, longname, shortname, hardware, role).Set(value)
	}
}
func (c *PrometheusCollector) updateMetricTimestamp(nodeID, metricName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.metricTimestamps[nodeID] == nil {
		c.metricTimestamps[nodeID] = make(map[string]time.Time)
	}
	c.metricTimestamps[nodeID][metricName] = time.Now()
}

func (c *PrometheusCollector) startMetricsTTLCleanup() {
	ctx, cancel := context.WithCancel(context.Background())
	c.mu.Lock()
	c.cleanupCancel = cancel
	c.mu.Unlock()

	cleanupInterval := c.metricsTTL / 2
	if cleanupInterval < minCleanupInterval {
		cleanupInterval = minCleanupInterval
	}
	if cleanupInterval > maxCleanupInterval {
		cleanupInterval = maxCleanupInterval
	}

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.cleanupExpiredMetrics()
		}
	}
}

func (c *PrometheusCollector) cleanupExpiredMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	for nodeID, metrics := range c.metricTimestamps {
		for metricName, timestamp := range metrics {
			if now.Sub(timestamp) > c.metricsTTL {
				c.deleteMetric(nodeID, metricName)
				delete(metrics, metricName)
			}
		}
		if len(metrics) == 0 {
			delete(c.metricTimestamps, nodeID)
		}
	}
}

func (c *PrometheusCollector) deleteMetric(nodeID, metricName string) {
	switch metricName {
	case domain.MetricTemperature:
		c.temperature.DeleteLabelValues(nodeID)
	case domain.MetricHumidity:
		c.humidity.DeleteLabelValues(nodeID)
	case domain.MetricPressure:
		c.pressure.DeleteLabelValues(nodeID)
	}
}

func (c *PrometheusCollector) Shutdown() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cleanupCancel != nil {
		c.cleanupCancel()
		c.cleanupCancel = nil
	}
}
