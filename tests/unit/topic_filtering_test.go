package tests

import (
	"encoding/json"
	"fmt"
	"testing"

	mqtt "github.com/mochi-mqtt/server/v2/packets"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"
)

func TestTopicFiltering_E2E(t *testing.T) {
	// Этот тест проверяет, что хук правильно фильтрует сообщения по TopicPrefix
	// Это была основная проблема - без правильного префикса сообщения игнорировались

	f := factory.NewDefaultFactory()
	hook := hooks.NewMeshtasticHookSimple(f)

	// Получаем registry один раз
	registry := f.CreateMetricsCollector().GetRegistry()

	testCases := []struct {
		name          string
		topic         string
		nodeID        int
		shouldProcess bool
	}{
		{
			name:          "meshtastic topic should be processed",
			topic:         "msh/2/2/json/LongFast/!broadcast",
			nodeID:        111111111,
			shouldProcess: true,
		},
		{
			name:          "meshtastic topic with different path should be processed",
			topic:         "msh/1/c/ShortFast/!direct",
			nodeID:        222222222,
			shouldProcess: true,
		},
		{
			name:          "non-meshtastic topic should be ignored",
			topic:         "other/topic/data",
			nodeID:        333333333,
			shouldProcess: false,
		},
		{
			name:          "topic without msh prefix should be ignored",
			topic:         "test/msh/data",
			nodeID:        444444444,
			shouldProcess: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Подготавливаем сообщение с уникальным nodeID
			testMessage := map[string]interface{}{
				"from": tc.nodeID,
				"type": "telemetry",
				"payload": map[string]interface{}{
					"battery_level": 85.5,
				},
			}
			payload, err := json.Marshal(testMessage)
			require.NoError(t, err)

			// Проверяем, есть ли метрика для этого узла
			initialNodeExists := hasNodeMetric(registry, tc.nodeID)

			// Создаем пакет
			packet := mqtt.Packet{
				TopicName: tc.topic,
				Payload:   payload,
			}

			// Обрабатываем сообщение через хук
			resultPacket, err := hook.OnPublish(nil, packet)
			require.NoError(t, err)
			assert.Equal(t, packet.TopicName, resultPacket.TopicName)

			// Проверяем, появилась ли метрика
			finalNodeExists := hasNodeMetric(registry, tc.nodeID)

			if tc.shouldProcess {
				assert.True(t, finalNodeExists,
					"Node metric should exist for processed messages")
			} else {
				assert.Equal(t, initialNodeExists, finalNodeExists,
					"Node metric existence should not change for ignored messages")
			}
		})
	}
}

func hasNodeMetric(registry *prometheus.Registry, nodeID int) bool {
	metricFamilies, err := registry.Gather()
	if err != nil {
		return false
	}

	nodeIDStr := fmt.Sprintf("%d", nodeID)
	for _, mf := range metricFamilies {
		if mf.GetName() == domain.MetricBatteryLevel {
			for _, metric := range mf.GetMetric() {
				for _, label := range metric.GetLabel() {
					if label.GetName() == "node_id" && label.GetValue() == nodeIDStr {
						return true
					}
				}
			}
		}
	}
	return false
}
