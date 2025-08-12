package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/adapters"
	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"
)

func TestMeshtasticHook_Shutdown_SavesState(t *testing.T) {
	// Создаем временный файл для состояния
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "test_embedded_state.json")

	// Создаем конфигурацию с файлом состояния
	config := createTestConfig(stateFile)
	f := factory.NewFactory(config)

	// Создаем хук
	hookConfig := hooks.MeshtasticHookConfig{
		ServerAddr:   "localhost:0", // Используем порт 0 для автоматического выбора
		EnableHealth: true,
		TopicPrefix:  "msh/",
	}
	hook := hooks.NewMeshtasticHook(hookConfig, f)
	require.NotNil(t, hook)

	// Инициализируем хук
	err := hook.Init(nil)
	require.NoError(t, err)

	// Добавляем тестовые данные через коллектор
	collector := f.CreateMetricsCollector()
	data := domain.TelemetryData{
		NodeID:       "987654321",
		BatteryLevel: floatPtr(75.0),
		Temperature:  floatPtr(25.5),
		Timestamp:    time.Now(),
	}
	err = collector.CollectTelemetry(data)
	require.NoError(t, err)

	// Вызываем OnStopped (имитируем остановку сервера)
	hook.OnStopped()

	// Проверяем, что файл состояния создался
	_, err = os.Stat(stateFile)
	assert.NoError(t, err, "state file should be created after hook shutdown")

	// Проверяем содержимое файла
	data_bytes, err := os.ReadFile(stateFile)
	require.NoError(t, err)
	assert.Contains(t, string(data_bytes), "987654321", "state file should contain node data")
}

func createTestConfig(stateFile string) domain.Config {
	mqttConfig := adapters.MQTTConfigAdapter{
		Host:     "localhost",
		Port:     1883,
		ClientID: "test-embedded-client",
		Topics:   []string{"test/+"},
	}

	prometheusConfig := adapters.PrometheusConfigAdapter{
		Listen:    "localhost:0",
		Path:      "/metrics",
		StateFile: stateFile,
	}

	alertManagerConfig := adapters.AlertManagerConfigAdapter{
		Listen: "localhost:0",
		Path:   "/alerts",
	}

	return adapters.NewConfigAdapter(mqttConfig, prometheusConfig, alertManagerConfig)
}

func floatPtr(f float64) *float64 {
	return &f
}
