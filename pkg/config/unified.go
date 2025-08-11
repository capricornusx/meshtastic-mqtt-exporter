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
		TLS            bool   `yaml:"tls"`
		AllowAnonymous bool   `yaml:"allow_anonymous"`
		Username       string `yaml:"username"`
		Password       string `yaml:"password"`
		Users          []struct {
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"users"`
	} `yaml:"mqtt"`

	Hook struct {
		Listen     string `yaml:"listen"`
		Prometheus struct {
			Path       string `yaml:"path"`
			MetricsTTL string `yaml:"metrics_ttl"`
			Topic      struct {
				Pattern string `yaml:"pattern"`
			} `yaml:"topic"`
			Debug struct {
				LogAllMessages bool `yaml:"log_all_messages"`
			} `yaml:"debug"`
			State struct {
				File string `yaml:"file"`
			} `yaml:"state"`
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
	config.Hook.Listen = "localhost:8100"
	config.Hook.Prometheus.Path = "/metrics"
	config.Hook.Prometheus.MetricsTTL = "30m"
	config.Hook.Prometheus.Topic.Pattern = domain.DefaultTopicPrefix
	config.Hook.Prometheus.Debug.LogAllMessages = false
	config.Hook.AlertManager.Path = domain.DefaultAlertsPath
}

func convertToAdapter(config *UnifiedConfig) (domain.Config, error) {
	// Устанавливаем уровень логирования
	logger.SetLogLevel(config.Logging.Level)

	metricsTTL, err := time.ParseDuration(config.Hook.Prometheus.MetricsTTL)
	if err != nil {
		metricsTTL = domain.DefaultMetricsTTL
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
		Host:           config.MQTT.Host,
		Port:           config.MQTT.Port,
		TLS:            config.MQTT.TLS,
		AllowAnonymous: config.MQTT.AllowAnonymous,
		Users:          users,
		Timeout:        domain.DefaultTimeout,
	}

	prometheusConfig := adapters.PrometheusConfigAdapter{
		Listen:         config.Hook.Listen,
		Path:           config.Hook.Prometheus.Path,
		MetricsTTL:     metricsTTL,
		TopicPattern:   config.Hook.Prometheus.Topic.Pattern,
		LogAllMessages: config.Hook.Prometheus.Debug.LogAllMessages,
		StateFile:      config.Hook.Prometheus.State.File,
	}

	alertManagerConfig := adapters.AlertManagerConfigAdapter{
		Listen: config.Hook.Listen,
		Path:   config.Hook.AlertManager.Path,
	}

	return adapters.NewConfigAdapter(mqttConfig, prometheusConfig, alertManagerConfig), nil
}
