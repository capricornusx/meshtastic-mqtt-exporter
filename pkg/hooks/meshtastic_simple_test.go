package hooks

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/factory"
)

func TestNewMeshtasticHookSimple_ActualFunction(t *testing.T) {
	f := factory.NewDefaultFactory()
	hook := NewMeshtasticHookSimple(f) // ← Тестируем реальную функцию

	assert.Equal(t, "localhost:8100", hook.config.ServerAddr)
	assert.Equal(t, domain.DefaultTopicPrefix, hook.config.TopicPrefix) // ← Проверяем TopicPrefix
	assert.True(t, hook.config.EnableHealth)
}

func TestFactory_MetricsCollectorSingleton(t *testing.T) {
	f := factory.NewDefaultFactory()

	collector1 := f.CreateMetricsCollector()
	collector2 := f.CreateMetricsCollector()

	// Должны быть одним и тем же объектом
	assert.Same(t, collector1, collector2)
}
