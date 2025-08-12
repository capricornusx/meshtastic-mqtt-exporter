package factory

import (
	mqtt "github.com/mochi-mqtt/server/v2"

	"meshtastic-exporter/pkg/application"
	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/infrastructure"
)

type Factory struct {
	config    domain.Config
	collector domain.MetricsCollector
}

func NewFactory(config domain.Config) *Factory {
	return &Factory{config: config}
}

// NewDefaultFactory creates a factory with empty config for tests.
func NewDefaultFactory() *Factory {
	return &Factory{config: nil}
}

func (f *Factory) CreateMetricsCollector() domain.MetricsCollector {
	return f.CreateMetricsCollectorWithMode("hook")
}

func (f *Factory) CreateMetricsCollectorWithMode(mode string) domain.MetricsCollector {
	if f.collector == nil {
		if f.config != nil {
			prometheusConfig := f.config.GetPrometheusConfig()
			ttl := prometheusConfig.GetMetricsTTL()
			f.collector = infrastructure.NewPrometheusCollectorWithConfig(mode, ttl)

			if stateFile := prometheusConfig.GetStateFile(); stateFile != "" {
				if err := f.collector.LoadState(stateFile); err != nil {
					_ = err
				}
			}
		} else {
			f.collector = infrastructure.NewPrometheusCollectorWithMode(mode)
		}
	}
	return f.collector
}

func (f *Factory) CreateAlertSender() domain.AlertSender {
	return infrastructure.NewLoRaAlertSender(nil, infrastructure.LoRaConfig{})
}

func (f *Factory) CreateAlertSenderWithMQTT(mqttServer interface{}) domain.AlertSender {
	if server, ok := mqttServer.(*mqtt.Server); ok {
		return infrastructure.NewLoRaAlertSender(server, infrastructure.LoRaConfig{})
	}
	return f.CreateAlertSender()
}

func (f *Factory) CreateMessageProcessor() domain.MessageProcessor {
	collector := f.CreateMetricsCollector()
	alerter := f.CreateAlertSender()
	logAllMessages := false
	topicPattern := ""
	if f.config != nil {
		prometheusConfig := f.config.GetPrometheusConfig()
		logAllMessages = prometheusConfig.GetLogAllMessages()
		topicPattern = prometheusConfig.GetTopicPattern()
	}
	return application.NewMeshtasticProcessor(collector, alerter, logAllMessages, topicPattern)
}

func (f *Factory) CreateMQTTClient(processor domain.MessageProcessor) *infrastructure.MQTTClient {
	return infrastructure.NewMQTTClient(f.config.GetMQTTConfig(), processor)
}

func (f *Factory) CreateHTTPServer(collector domain.MetricsCollector, alerter domain.AlertSender) *infrastructure.HTTPServer {
	prometheusConfig := f.config.GetPrometheusConfig()
	addr := prometheusConfig.GetListen()
	return infrastructure.NewHTTPServer(addr, collector, alerter)
}

func (f *Factory) GetPrometheusConfig() domain.PrometheusConfig {
	if f.config == nil {
		return nil
	}
	return f.config.GetPrometheusConfig()
}

func (f *Factory) GetAlertManagerConfig() domain.AlertManagerConfig {
	if f.config == nil {
		return nil
	}
	return f.config.GetAlertManagerConfig()
}
