package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/adapters"
	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"
)

// тест требует рефакторинга
func TestEmbeddedModeUsesConfigTopics(t *testing.T) {
	mqttConfig := adapters.MQTTConfigAdapter{
		Host:   "localhost",
		Port:   1883,
		Topics: []string{"custom/+/+/json/+/+", "test/+/json/+/+"},
	}

	prometheusConfig := adapters.PrometheusConfigAdapter{
		Listen:       "localhost:8100",
		Path:         "/metrics",
		TopicPattern: "custom/",
	}

	alertConfig := adapters.AlertManagerConfigAdapter{
		Listen: "localhost:8100",
		Path:   "/alerts",
	}

	cfg := adapters.NewConfigAdapter(mqttConfig, prometheusConfig, alertConfig)
	f := factory.NewFactory(cfg)

	hookConfig := hooks.MeshtasticHookConfig{
		ServerAddr:  prometheusConfig.GetListen(),
		TopicPrefix: prometheusConfig.GetTopicPattern(),
	}
	hook := hooks.NewMeshtasticHook(hookConfig, f)

	require.NotNil(t, hook)
	assert.Equal(t, "meshtastic", hook.ID())

	// Проверяем, что MQTT конфигурация содержит правильные топики
	mqttCfg := cfg.GetMQTTConfig()
	topics := mqttCfg.GetTopics()
	assert.Contains(t, topics, "custom/+/+/json/+/+")
	assert.Contains(t, topics, "test/+/json/+/+")

	// Проверяем, что Prometheus конфигурация содержит правильный паттерн
	promCfg := cfg.GetPrometheusConfig()
	assert.Equal(t, "custom/", promCfg.GetTopicPattern())
}

func TestConfigLoadingWithCustomTopics(t *testing.T) {
	// Тестируем загрузку конфигурации с кастомными топиками используя мок
	mqttConfig := adapters.MQTTConfigAdapter{
		Host:   "localhost",
		Port:   1883,
		Topics: []string{"msh/+/2/json/#", "test/+/json/+/+"},
	}

	prometheusConfig := adapters.PrometheusConfigAdapter{
		Listen:       "localhost:8100",
		Path:         "/metrics",
		TopicPattern: "msh/+/2/json/#",
	}

	alertConfig := adapters.AlertManagerConfigAdapter{
		Listen: "localhost:8100",
		Path:   "/alerts",
	}

	cfg := adapters.NewConfigAdapter(mqttConfig, prometheusConfig, alertConfig)

	mqttCfg := cfg.GetMQTTConfig()
	topics := mqttCfg.GetTopics()

	// Проверяем, что топики загружены правильно
	assert.NotEmpty(t, topics)
	assert.Contains(t, topics, "msh/+/2/json/#")
	assert.Contains(t, topics, "test/+/json/+/+")

	promCfg := cfg.GetPrometheusConfig()
	topicPattern := promCfg.GetTopicPattern()

	// Проверяем, что паттерн загружен правильно
	assert.Equal(t, "msh/+/2/json/#", topicPattern)
}
