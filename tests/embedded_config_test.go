package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/adapters"
	"meshtastic-exporter/pkg/config"
	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"
)

func TestEmbeddedModeUsesConfigTopics(t *testing.T) {
	// Создаем конфигурацию напрямую
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

	// Создаем хук с конфигурацией из factory
	hookConfig := hooks.MeshtasticHookConfig{
		ServerAddr:  prometheusConfig.GetListen(),
		TopicPrefix: prometheusConfig.GetTopicPattern(), // Используем из конфигурации
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
	// Тестируем загрузку конфигурации с кастомными топиками
	// Используем существующий config.yaml, но проверяем, что он правильно парсится

	cfg, err := config.LoadUnifiedConfig("../config.yaml")
	require.NoError(t, err)

	mqttCfg := cfg.GetMQTTConfig()
	topics := mqttCfg.GetTopics()

	// Проверяем, что загружены дефолтные топики (если не указаны в конфиге)
	assert.NotEmpty(t, topics)
	assert.True(t, len(topics) > 0, "Should have at least one topic")

	promCfg := cfg.GetPrometheusConfig()
	topicPattern := promCfg.GetTopicPattern()

	// Проверяем, что паттерн из конфига загружен правильно
	assert.Equal(t, "msh/+/2/json/#", topicPattern)
}
