package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"meshtastic-exporter/pkg/factory"
)

const batteryMetricName = "meshtastic_battery_level_percent"

func testContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestFactory_MetricsCollectorSingleton_E2E(t *testing.T) {
	// Этот тест проверяет, что factory возвращает один и тот же collector
	// для всех компонентов, что критично для работы метрик

	f := factory.NewDefaultFactory()

	// Создаем collector напрямую
	collector1 := f.CreateMetricsCollector()

	// Создаем message processor, который внутри создает collector
	processor := f.CreateMessageProcessor()

	// Создаем еще один collector
	collector2 := f.CreateMetricsCollector()

	// Все должны быть одним и тем же объектом
	assert.Same(t, collector1, collector2, "Multiple calls to CreateMetricsCollector should return same instance")

	// Проверяем, что processor использует тот же collector
	// Это можно проверить через registry - они должны быть одинаковыми
	registry1 := collector1.GetRegistry()
	registry2 := collector2.GetRegistry()

	assert.Same(t, registry1, registry2, "All collectors should use same registry")

	// Дополнительная проверка: записываем метрику через processor
	// и проверяем, что она появилась в collector
	ctx := testContext(t)
	telemetryPayload := []byte(`{
		"from": 999888777,
		"type": "telemetry",
		"payload": {
			"battery_level": 95.5
		}
	}`)

	err := processor.ProcessMessage(ctx, "msh/test", telemetryPayload)
	assert.NoError(t, err)

	// Проверяем, что метрика появилась в registry
	metricFamilies, err := registry1.Gather()
	assert.NoError(t, err)

	// Ищем метрику battery_level
	found := false
	for _, mf := range metricFamilies {
		if mf.GetName() == batteryMetricName {
			found = true
			break
		}
	}
	assert.True(t, found, "Battery level metric should be present in registry")
}
