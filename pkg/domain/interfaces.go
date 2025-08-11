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
	GetTLS() bool
	GetUsers() []UserAuth
	GetTimeout() time.Duration
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
