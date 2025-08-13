package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"meshtastic-exporter/pkg/adapters"
	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/errors"
	"meshtastic-exporter/pkg/logger"
)

type AlertRoute struct {
	Mode         string   `yaml:"mode"`
	TargetNodes  []string `yaml:"target_nodes"`
	ShowOnSender bool     `yaml:"show_on_sender"`
}

type UnifiedConfig struct {
	Logging struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`

	MQTT struct {
		Host           string   `yaml:"host"`
		Port           int      `yaml:"port"`
		AllowAnonymous bool     `yaml:"allow_anonymous"`
		Username       string   `yaml:"username"`
		Password       string   `yaml:"password"`
		ClientID       string   `yaml:"client_id"`
		Topics         []string `yaml:"topics"`
		Users          []struct {
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"users"`
		TLSConfig struct {
			Enabled            bool   `yaml:"enabled"`
			Port               int    `yaml:"port"`
			CertFile           string `yaml:"cert_file"`
			KeyFile            string `yaml:"key_file"`
			CAFile             string `yaml:"ca_file"`
			InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
			MinVersion         string `yaml:"min_version"`
		} `yaml:"tls_config"`
		Capabilities struct {
			MaximumInflight              int    `yaml:"maximum_inflight"`
			MaximumClientWritesPending   int    `yaml:"maximum_client_writes_pending"`
			ReceiveMaximum               int    `yaml:"receive_maximum"`
			MaximumQoS                   int    `yaml:"maximum_qos"`
			RetainAvailable              bool   `yaml:"retain_available"`
			MaximumMessageExpiryInterval string `yaml:"maximum_message_expiry_interval"`
			MaximumClients               int    `yaml:"maximum_clients"`
		} `yaml:"capabilities"`
	} `yaml:"mqtt"`

	Hook struct {
		Listen     string `yaml:"listen"`
		Prometheus struct {
			Path       string `yaml:"path"`
			MetricsTTL string `yaml:"metrics_ttl"`
			KeepAlive  string `yaml:"keep_alive"`
			Topic      struct {
				Pattern        string `yaml:"pattern"`
				LogAllMessages bool   `yaml:"log_all_messages"`
			} `yaml:"topic"`
			StateFile string `yaml:"state_file"`
		} `yaml:"prometheus"`
		AlertManager struct {
			Path       string `yaml:"path"`
			MQTTTopic  string `yaml:"mqtt_topic"`
			FromNodeID string `yaml:"from_node_id"`
			Routing    struct {
				Default  *AlertRoute `yaml:"default"`
				Critical *AlertRoute `yaml:"critical"`
				Warning  *AlertRoute `yaml:"warning"`
				Info     *AlertRoute `yaml:"info"`
			} `yaml:"routing"`
		} `yaml:"alertmanager"`
	} `yaml:"hook"`
}

func LoadUnifiedConfig(filename string) (domain.Config, error) {
	config := &UnifiedConfig{}
	setDefaults(config)

	if data, err := os.ReadFile(filename); err == nil {
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, errors.NewConfigError("failed to parse yaml", err)
		}
	}

	return convertToAdapter(config)
}

func setDefaults(config *UnifiedConfig) {
	config.Logging.Level = "info"
	config.MQTT.Host = "localhost"
	config.MQTT.Port = 1883
	config.MQTT.ClientID = domain.DefaultMQTTClientID
	config.MQTT.Topics = domain.GetDefaultMQTTTopics()
	config.MQTT.TLSConfig.Port = 8883
	config.MQTT.TLSConfig.InsecureSkipVerify = domain.DefaultTLSInsecureSkipVerify
	config.MQTT.TLSConfig.MinVersion = domain.DefaultTLSVersionString
	config.MQTT.Capabilities.MaximumInflight = 1024
	config.MQTT.Capabilities.MaximumClientWritesPending = 1000
	config.MQTT.Capabilities.ReceiveMaximum = 512
	config.MQTT.Capabilities.MaximumQoS = 2
	config.MQTT.Capabilities.RetainAvailable = true
	config.MQTT.Capabilities.MaximumMessageExpiryInterval = "24h"
	config.MQTT.Capabilities.MaximumClients = 1000
	config.Hook.Listen = "localhost:8100"
	config.Hook.Prometheus.Path = "/metrics"
	config.Hook.Prometheus.MetricsTTL = "30m"
	config.Hook.Prometheus.Topic.Pattern = domain.DefaultTopicPrefix
	config.Hook.Prometheus.Topic.LogAllMessages = false
	config.Hook.AlertManager.Path = domain.DefaultAlertsPath
}

func chooseTLSVersion(config *UnifiedConfig) uint16 {
	tlsMinVersion := uint16(domain.DefaultTLSMinVersion)
	switch config.MQTT.TLSConfig.MinVersion {
	case "1.0":
		tlsMinVersion = 0x0301
	case "1.1":
		tlsMinVersion = 0x0302
	case "1.2":
		tlsMinVersion = 0x0303
	case "1.3":
		tlsMinVersion = 0x0304
	}
	return tlsMinVersion
}

func convertToAdapter(config *UnifiedConfig) (domain.Config, error) {
	logger.SetLogLevel(config.Logging.Level)

	mqttConfig := buildMQTTConfig(config)
	prometheusConfig := buildPrometheusConfig(config)
	alertManagerConfig := buildAlertManagerConfig(config)

	return adapters.NewConfigAdapter(mqttConfig, prometheusConfig, alertManagerConfig), nil
}

func buildMQTTConfig(config *UnifiedConfig) adapters.MQTTConfigAdapter {
	keepAlive := parseKeepAlive(config.Hook.Prometheus.KeepAlive)
	messageExpiry := parseMessageExpiry(config.MQTT.Capabilities.MaximumMessageExpiryInterval)
	users := buildUsersList(config)

	return adapters.MQTTConfigAdapter{
		Host:            config.MQTT.Host,
		Port:            config.MQTT.Port,
		AllowAnonymous:  config.MQTT.AllowAnonymous,
		Users:           users,
		Timeout:         domain.DefaultTimeout,
		KeepAlive:       keepAlive,
		MaxInflight:     config.MQTT.Capabilities.MaximumInflight,
		MaxQueued:       config.MQTT.Capabilities.MaximumClientWritesPending,
		ReceiveMaximum:  config.MQTT.Capabilities.ReceiveMaximum,
		MaxQoS:          config.MQTT.Capabilities.MaximumQoS,
		RetainAvailable: config.MQTT.Capabilities.RetainAvailable,
		MaxClients:      config.MQTT.Capabilities.MaximumClients,
		MessageExpiry:   messageExpiry,
		ClientID:        config.MQTT.ClientID,
		Topics:          config.MQTT.Topics,
		TLSConfig: adapters.TLSConfigAdapter{
			Enabled:            config.MQTT.TLSConfig.Enabled,
			Port:               config.MQTT.TLSConfig.Port,
			CertFile:           config.MQTT.TLSConfig.CertFile,
			KeyFile:            config.MQTT.TLSConfig.KeyFile,
			CAFile:             config.MQTT.TLSConfig.CAFile,
			InsecureSkipVerify: config.MQTT.TLSConfig.InsecureSkipVerify,
			MinVersion:         chooseTLSVersion(config),
		},
	}
}

func buildPrometheusConfig(config *UnifiedConfig) adapters.PrometheusConfigAdapter {
	metricsTTL, err := time.ParseDuration(config.Hook.Prometheus.MetricsTTL)
	if err != nil {
		metricsTTL = domain.DefaultMetricsTTL
	}

	return adapters.PrometheusConfigAdapter{
		Listen:         config.Hook.Listen,
		Path:           config.Hook.Prometheus.Path,
		MetricsTTL:     metricsTTL,
		TopicPattern:   config.Hook.Prometheus.Topic.Pattern,
		LogAllMessages: config.Hook.Prometheus.Topic.LogAllMessages,
		StateFile:      config.Hook.Prometheus.StateFile,
	}
}

func buildAlertManagerConfig(config *UnifiedConfig) adapters.AlertManagerConfigAdapter {
	routingConfig := adapters.AlertRoutingConfig{
		Default:  convertAlertRoute(config.Hook.AlertManager.Routing.Default),
		Critical: convertAlertRoute(config.Hook.AlertManager.Routing.Critical),
		Warning:  convertAlertRoute(config.Hook.AlertManager.Routing.Warning),
		Info:     convertAlertRoute(config.Hook.AlertManager.Routing.Info),
	}

	var fromNodeID uint32
	if config.Hook.AlertManager.FromNodeID != "" {
		if nodeID, err := parseNodeID(config.Hook.AlertManager.FromNodeID); err == nil {
			fromNodeID = nodeID
		}
	}

	return adapters.AlertManagerConfigAdapter{
		Listen:     config.Hook.Listen,
		Path:       config.Hook.AlertManager.Path,
		MQTTTopic:  config.Hook.AlertManager.MQTTTopic,
		FromNodeID: fromNodeID,
		Routing:    routingConfig,
	}
}

func parseKeepAlive(keepAliveStr string) time.Duration {
	if keepAliveStr == "" {
		return domain.DefaultKeepAlive
	}
	if parsed, err := time.ParseDuration(keepAliveStr); err == nil {
		return parsed
	}
	return domain.DefaultKeepAlive
}

func parseMessageExpiry(expiryStr string) time.Duration {
	if expiryStr == "0" {
		return 0
	}
	if expiryStr != "" {
		if parsed, err := time.ParseDuration(expiryStr); err == nil {
			return parsed
		}
	}
	return time.Hour
}

func buildUsersList(config *UnifiedConfig) []adapters.UserAuthAdapter {
	var users []adapters.UserAuthAdapter
	for _, u := range config.MQTT.Users {
		users = append(users, adapters.UserAuthAdapter{
			Username: u.Username,
			Password: u.Password,
		})
	}

	if config.MQTT.Username != "" {
		users = append(users, adapters.UserAuthAdapter{
			Username: config.MQTT.Username,
			Password: config.MQTT.Password,
		})
	}
	return users
}

func convertAlertRoute(route *AlertRoute) *adapters.AlertRouteConfig {
	if route == nil {
		return nil
	}

	var targetNodes []uint32
	for _, nodeStr := range route.TargetNodes {
		if nodeID, err := parseNodeID(nodeStr); err == nil {
			targetNodes = append(targetNodes, nodeID)
		}
	}

	return &adapters.AlertRouteConfig{
		Mode:         route.Mode,
		TargetNodes:  targetNodes,
		ShowOnSender: route.ShowOnSender,
	}
}

func parseNodeID(nodeStr string) (uint32, error) {
	// Поддержка hex формата (0x prefix или !)
	if strings.HasPrefix(nodeStr, "0x") || strings.HasPrefix(nodeStr, "0X") {
		val, err := strconv.ParseUint(nodeStr[2:], 16, 32)
		return uint32(val), err
	}
	if strings.HasPrefix(nodeStr, "!") {
		val, err := strconv.ParseUint(nodeStr[1:], 16, 32)
		return uint32(val), err
	}
	// Десятичный формат по умолчанию
	val, err := strconv.ParseUint(nodeStr, 10, 32)
	return uint32(val), err
}
