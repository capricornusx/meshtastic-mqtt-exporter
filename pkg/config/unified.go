package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"meshtastic-exporter/pkg/adapters"
	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/errors"
	"meshtastic-exporter/pkg/logger"
)

type UnifiedConfig struct {
	Logging struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`

	MQTT struct {
		Host           string `yaml:"host"`
		Port           int    `yaml:"port"`
		AllowAnonymous bool   `yaml:"allow_anonymous"`
		Username       string `yaml:"username"`
		Password       string `yaml:"password"`
		Users          []struct {
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"users"`
		TLSConfig struct {
			Enabled  bool   `yaml:"enabled"`
			Port     int    `yaml:"port"`
			CertFile string `yaml:"cert_file"`
			KeyFile  string `yaml:"key_file"`
			CAFile   string `yaml:"ca_file"`
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
			Path string `yaml:"path"`
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
	config.MQTT.TLSConfig.Port = 8883
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

func convertToAdapter(config *UnifiedConfig) (domain.Config, error) {
	// Устанавливаем уровень логирования
	logger.SetLogLevel(config.Logging.Level)

	metricsTTL, err := time.ParseDuration(config.Hook.Prometheus.MetricsTTL)
	if err != nil {
		metricsTTL = domain.DefaultMetricsTTL
	}

	keepAlive := domain.DefaultKeepAlive
	if config.Hook.Prometheus.KeepAlive != "" {
		if parsed, err := time.ParseDuration(config.Hook.Prometheus.KeepAlive); err == nil {
			keepAlive = parsed
		}
	}

	messageExpiry := time.Hour
	if config.MQTT.Capabilities.MaximumMessageExpiryInterval != "" && config.MQTT.Capabilities.MaximumMessageExpiryInterval != "0" {
		if parsed, err := time.ParseDuration(config.MQTT.Capabilities.MaximumMessageExpiryInterval); err == nil {
			messageExpiry = parsed
		}
	} else if config.MQTT.Capabilities.MaximumMessageExpiryInterval == "0" {
		messageExpiry = 0
	}

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

	mqttConfig := adapters.MQTTConfigAdapter{
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
		TLSConfig: adapters.TLSConfigAdapter{
			Enabled:  config.MQTT.TLSConfig.Enabled,
			Port:     config.MQTT.TLSConfig.Port,
			CertFile: config.MQTT.TLSConfig.CertFile,
			KeyFile:  config.MQTT.TLSConfig.KeyFile,
			CAFile:   config.MQTT.TLSConfig.CAFile,
		},
	}

	prometheusConfig := adapters.PrometheusConfigAdapter{
		Listen:         config.Hook.Listen,
		Path:           config.Hook.Prometheus.Path,
		MetricsTTL:     metricsTTL,
		TopicPattern:   config.Hook.Prometheus.Topic.Pattern,
		LogAllMessages: config.Hook.Prometheus.Topic.LogAllMessages,
		StateFile:      config.Hook.Prometheus.StateFile,
	}

	alertManagerConfig := adapters.AlertManagerConfigAdapter{
		Listen: config.Hook.Listen,
		Path:   config.Hook.AlertManager.Path,
	}

	return adapters.NewConfigAdapter(mqttConfig, prometheusConfig, alertManagerConfig), nil
}
