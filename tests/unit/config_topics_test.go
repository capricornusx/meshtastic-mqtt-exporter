package tests

import (
	"encoding/json"
	"testing"

	mqtt "github.com/mochi-mqtt/server/v2/packets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/adapters"
	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"
)

func TestConfigurableTopicPrefix(t *testing.T) {
	testCases := []struct {
		name          string
		topicPrefix   string
		testTopic     string
		shouldProcess bool
	}{
		{
			name:          "default msh prefix should work",
			topicPrefix:   "msh/",
			testTopic:     "msh/2/2/json/LongFast/!broadcast",
			shouldProcess: true,
		},
		{
			name:          "custom prefix should work",
			topicPrefix:   "custom/",
			testTopic:     "custom/2/2/json/LongFast/!broadcast",
			shouldProcess: true,
		},
		{
			name:          "wrong prefix should be ignored",
			topicPrefix:   "custom/",
			testTopic:     "msh/2/2/json/LongFast/!broadcast",
			shouldProcess: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем конфигурацию с кастомным префиксом
			mqttConfig := adapters.MQTTConfigAdapter{
				Host:   "localhost",
				Port:   1883,
				Topics: []string{tc.topicPrefix + "+/+/json/+/+"},
			}

			prometheusConfig := adapters.PrometheusConfigAdapter{
				Listen:       "localhost:8100",
				Path:         "/metrics",
				TopicPattern: tc.topicPrefix,
			}

			alertConfig := adapters.AlertManagerConfigAdapter{
				Listen: "localhost:8100",
				Path:   "/alerts",
			}

			config := adapters.NewConfigAdapter(mqttConfig, prometheusConfig, alertConfig)
			f := factory.NewFactory(config)

			// Создаем хук с конфигурацией
			hookConfig := hooks.MeshtasticHookConfig{
				ServerAddr:  "localhost:8100",
				TopicPrefix: tc.topicPrefix,
			}
			hook := hooks.NewMeshtasticHook(hookConfig, f)

			// Подготавливаем тестовое сообщение
			testMessage := map[string]interface{}{
				"from": 123456789,
				"type": "telemetry",
				"payload": map[string]interface{}{
					"battery_level": 85.5,
				},
			}
			payload, err := json.Marshal(testMessage)
			require.NoError(t, err)

			// Создаем пакет
			packet := mqtt.Packet{
				TopicName: tc.testTopic,
				Payload:   payload,
			}

			// Получаем registry для проверки метрик
			registry := f.CreateMetricsCollector().GetRegistry()
			initialNodeExists := hasNodeMetric(registry, 123456789)

			// Обрабатываем сообщение
			resultPacket, err := hook.OnPublish(nil, packet)
			require.NoError(t, err)
			assert.Equal(t, packet.TopicName, resultPacket.TopicName)

			// Проверяем результат
			finalNodeExists := hasNodeMetric(registry, 123456789)

			if tc.shouldProcess {
				assert.True(t, finalNodeExists,
					"Сообщение с префиксом '%s' должно обрабатываться", tc.topicPrefix)
			} else {
				assert.Equal(t, initialNodeExists, finalNodeExists,
					"Сообщение с неправильным префиксом не должно обрабатываться")
			}
		})
	}
}
