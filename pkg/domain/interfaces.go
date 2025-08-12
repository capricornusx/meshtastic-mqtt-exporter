package domain

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollector interface {
	CollectTelemetry(data TelemetryData) error
	CollectNodeInfo(info NodeInfo) error
	GetRegistry() *prometheus.Registry
	SaveState(filename string) error
	LoadState(filename string) error
}

type AlertSender interface {
	SendAlert(ctx context.Context, alert Alert) error
}

type MessageProcessor interface {
	ProcessMessage(ctx context.Context, topic string, payload []byte) error
}

type Config interface {
	GetMQTTConfig() MQTTConfig
	GetPrometheusConfig() PrometheusConfig
	GetAlertManagerConfig() AlertManagerConfig
	Validate() error
}

type MQTTConfig interface {
	GetHost() string
	GetPort() int
	GetUsers() []UserAuth
	GetTimeout() time.Duration
	GetKeepAlive() time.Duration
	GetTLSConfig() TLSConfig
	GetMaxInflight() int
	GetMaxQueued() int
	GetReceiveMaximum() int
	GetMaxQoS() int
	GetRetainAvailable() bool
	GetMessageExpiry() int64
	GetMaxClients() int
	GetClientID() string
	GetTopics() []string
}

type TLSConfig interface {
	GetEnabled() bool
	GetPort() int
	GetCertFile() string
	GetKeyFile() string
	GetCAFile() string
	GetInsecureSkipVerify() bool
	GetMinVersion() uint16
}

type PrometheusConfig interface {
	GetListen() string
	GetPath() string
	GetMetricsTTL() time.Duration
	GetTopicPattern() string
	GetLogAllMessages() bool
	GetStateFile() string
}

type AlertManagerConfig interface {
	GetListen() string
	GetPath() string
}

type UserAuth interface {
	GetUsername() string
	GetPassword() string
}
