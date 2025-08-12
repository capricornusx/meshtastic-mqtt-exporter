package factory

import (
	"testing"

	"meshtastic-exporter/pkg/adapters"
)

func TestNewFactory(t *testing.T) {
	mqttConfig := adapters.MQTTConfigAdapter{Host: "localhost", Port: 1883}
	prometheusConfig := adapters.PrometheusConfigAdapter{Listen: "localhost:8100", Path: "/metrics", TopicPattern: "msh/#", LogAllMessages: false}
	alertConfig := adapters.AlertManagerConfigAdapter{Listen: "localhost:8100", Path: "/alerts"}
	config := adapters.NewConfigAdapter(mqttConfig, prometheusConfig, alertConfig)

	factory := NewFactory(config)
	if factory == nil {
		t.Fatal("Expected factory to be created")
	}
	if factory.config != config {
		t.Error("Expected config to be set")
	}
}

func TestNewDefaultFactory(t *testing.T) {
	factory := NewDefaultFactory()
	if factory == nil {
		t.Fatal("Expected factory to be created")
	}
	if factory.config != nil {
		t.Error("Expected config to be nil")
	}
}

func TestCreateMetricsCollector(t *testing.T) {
	factory := NewDefaultFactory()
	collector := factory.CreateMetricsCollector()
	if collector == nil {
		t.Fatal("Expected metrics collector to be created")
	}
}

func TestCreateAlertSender(t *testing.T) {
	factory := NewDefaultFactory()
	alerter := factory.CreateAlertSender()
	if alerter == nil {
		t.Fatal("Expected alert sender to be created")
	}
}

func TestCreateMessageProcessor(t *testing.T) {
	factory := NewDefaultFactory()
	processor := factory.CreateMessageProcessor()
	if processor == nil {
		t.Fatal("Expected message processor to be created")
	}
}

func TestCreateMQTTClient(t *testing.T) {
	mqttConfig := adapters.MQTTConfigAdapter{Host: "localhost", Port: 1883}
	prometheusConfig := adapters.PrometheusConfigAdapter{Listen: "localhost:8100", Path: "/metrics", TopicPattern: "msh/#", LogAllMessages: false}
	alertConfig := adapters.AlertManagerConfigAdapter{Listen: "localhost:8100", Path: "/alerts"}
	config := adapters.NewConfigAdapter(mqttConfig, prometheusConfig, alertConfig)

	factory := NewFactory(config)
	processor := factory.CreateMessageProcessor()
	client := factory.CreateMQTTClient(processor)
	if client == nil {
		t.Fatal("Expected MQTT client to be created")
	}
}

func TestCreateHTTPServer(t *testing.T) {
	mqttConfig := adapters.MQTTConfigAdapter{Host: "localhost", Port: 1883}
	prometheusConfig := adapters.PrometheusConfigAdapter{Listen: "localhost:8100", Path: "/metrics", TopicPattern: "msh/#", LogAllMessages: false}
	alertConfig := adapters.AlertManagerConfigAdapter{Listen: "localhost:8100", Path: "/alerts"}
	config := adapters.NewConfigAdapter(mqttConfig, prometheusConfig, alertConfig)

	factory := NewFactory(config)
	collector := factory.CreateMetricsCollector()
	alerter := factory.CreateAlertSender()
	server := factory.CreateHTTPServer(collector, alerter)
	if server == nil {
		t.Fatal("Expected HTTP server to be created")
	}
}
