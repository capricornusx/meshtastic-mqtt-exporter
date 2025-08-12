package standalone

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/adapters"
	"meshtastic-exporter/pkg/domain"
)

func TestApp_Shutdown_SavesState(t *testing.T) {
	// Создаем временный файл для состояния
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "test_state.json")

	// Создаем конфигурацию с файлом состояния
	config := createTestConfig(stateFile)

	// Создаем приложение
	app := NewApp(config)

	// Добавляем тестовые данные в коллектор
	data := domain.TelemetryData{
		NodeID:       "123456789",
		BatteryLevel: floatPtr(85.5),
		Temperature:  floatPtr(23.4),
		Timestamp:    time.Now(),
	}
	err := app.collector.CollectTelemetry(data)
	require.NoError(t, err)

	// Вызываем shutdown
	err = app.Shutdown()
	require.NoError(t, err)

	// Проверяем, что файл состояния создался
	_, err = os.Stat(stateFile)
	assert.NoError(t, err, "state file should be created after shutdown")

	// Проверяем содержимое файла
	data_bytes, err := os.ReadFile(stateFile)
	require.NoError(t, err)
	assert.Contains(t, string(data_bytes), "123456789", "state file should contain node data")
}

func createTestConfig(stateFile string) domain.Config {
	mqttConfig := adapters.MQTTConfigAdapter{
		Host:     "localhost",
		Port:     1883,
		ClientID: "test-client",
		Topics:   []string{"test/+"},
	}

	prometheusConfig := adapters.PrometheusConfigAdapter{
		Listen:    "localhost:8100",
		Path:      "/metrics",
		StateFile: stateFile,
	}

	alertManagerConfig := adapters.AlertManagerConfigAdapter{
		Listen: "localhost:8080",
		Path:   "/alerts",
	}

	return adapters.NewConfigAdapter(mqttConfig, prometheusConfig, alertManagerConfig)
}

func floatPtr(f float64) *float64 {
	return &f
}
