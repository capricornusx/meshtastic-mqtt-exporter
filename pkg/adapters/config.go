package adapters

import (
	"fmt"
	"time"

	"meshtastic-exporter/pkg/domain"
)

type ConfigAdapter struct {
	mqtt         MQTTConfigAdapter
	prometheus   PrometheusConfigAdapter
	alertManager AlertManagerConfigAdapter
}

type MQTTConfigAdapter struct {
	Host           string
	Port           int
	TLS            bool
	AllowAnonymous bool
	Users          []UserAuthAdapter
	Timeout        time.Duration
}

type UserAuthAdapter struct {
	Username string
	Password string
}

type PrometheusConfigAdapter struct {
	Listen         string
	Path           string
	MetricsTTL     time.Duration
	TopicPattern   string
	LogAllMessages bool
	StateFile      string
}

type AlertManagerConfigAdapter struct {
	Listen string
	Path   string
}

func NewConfigAdapter(mqtt MQTTConfigAdapter, prometheus PrometheusConfigAdapter, alertManager AlertManagerConfigAdapter) *ConfigAdapter {
	return &ConfigAdapter{
		mqtt:         mqtt,
		prometheus:   prometheus,
		alertManager: alertManager,
	}
}

func (c *ConfigAdapter) GetMQTTConfig() domain.MQTTConfig {
	return &c.mqtt
}

func (c *ConfigAdapter) GetPrometheusConfig() domain.PrometheusConfig {
	return &c.prometheus
}

func (c *ConfigAdapter) GetAlertManagerConfig() domain.AlertManagerConfig {
	return &c.alertManager
}

func (c *ConfigAdapter) Validate() error {
	if c.mqtt.Host == "" {
		return fmt.Errorf("MQTT host cannot be empty")
	}
	if c.mqtt.Port <= 0 || c.mqtt.Port > 65535 {
		return fmt.Errorf("invalid MQTT port: %d", c.mqtt.Port)
	}
	if c.prometheus.Listen == "" {
		return fmt.Errorf("prometheus listen address cannot be empty")
	}
	if c.alertManager.Listen == "" {
		return fmt.Errorf("alertmanager listen address cannot be empty")
	}
	return nil
}

func (m *MQTTConfigAdapter) GetHost() string           { return m.Host }
func (m *MQTTConfigAdapter) GetPort() int              { return m.Port }
func (m *MQTTConfigAdapter) GetTLS() bool              { return m.TLS }
func (m *MQTTConfigAdapter) GetTimeout() time.Duration { return m.Timeout }
func (m *MQTTConfigAdapter) GetUsers() []domain.UserAuth {
	users := make([]domain.UserAuth, len(m.Users))
	for i, u := range m.Users {
		users[i] = &u
	}
	return users
}

func (u *UserAuthAdapter) GetUsername() string { return u.Username }
func (u *UserAuthAdapter) GetPassword() string { return u.Password }

func (p *PrometheusConfigAdapter) GetListen() string            { return p.Listen }
func (p *PrometheusConfigAdapter) GetPath() string              { return p.Path }
func (p *PrometheusConfigAdapter) GetMetricsTTL() time.Duration { return p.MetricsTTL }
func (p *PrometheusConfigAdapter) GetTopicPattern() string      { return p.TopicPattern }
func (p *PrometheusConfigAdapter) GetLogAllMessages() bool      { return p.LogAllMessages }
func (p *PrometheusConfigAdapter) GetStateFile() string         { return p.StateFile }

func (a *AlertManagerConfigAdapter) GetListen() string { return a.Listen }
func (a *AlertManagerConfigAdapter) GetPath() string   { return a.Path }
