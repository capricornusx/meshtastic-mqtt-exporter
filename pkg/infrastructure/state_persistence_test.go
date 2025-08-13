package infrastructure

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/domain"
)

func TestPrometheusCollector_StatePersistence(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	// Добавляем тестовые данные
	data := domain.TelemetryData{
		NodeID:       "123456789",
		BatteryLevel: floatPtr(85.5),
		Temperature:  floatPtr(23.4),
		Voltage:      floatPtr(4.1),
		Timestamp:    time.Now(),
	}
	err := collector.CollectTelemetry(data)
	require.NoError(t, err)

	info := domain.NodeInfo{
		NodeID:    "123456789",
		LongName:  "Test Node",
		ShortName: "TN01",
		Hardware:  "1",
		Role:      "2",
		Timestamp: time.Now(),
	}
	err = collector.CollectNodeInfo(info)
	require.NoError(t, err)

	// Сохраняем состояние
	tempFile := filepath.Join(os.TempDir(), "test_state.json")
	defer os.Remove(tempFile)

	err = collector.SaveState(tempFile)
	require.NoError(t, err)

	// Проверяем, что файл создался
	_, err = os.Stat(tempFile)
	require.NoError(t, err)

	// Создаем новый collector и загружаем состояние
	newCollector := NewPrometheusCollector()
	err = newCollector.LoadState(tempFile)
	require.NoError(t, err)
}

func TestPrometheusCollector_StateCorruptedFile(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	// Создаем поврежденный файл
	tempFile := filepath.Join(os.TempDir(), "corrupted_state.json")
	defer os.Remove(tempFile)

	err := os.WriteFile(tempFile, []byte("invalid json content"), 0600)
	require.NoError(t, err)

	// Загрузка должна завершиться с ошибкой
	err = collector.LoadState(tempFile)
	require.Error(t, err)
}

func TestPrometheusCollector_StateNonExistentFile(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	// Попытка загрузить несуществующий файл - должна завершиться успешно (файл не найден = начинаем с чистого состояния)
	err := collector.LoadState("/non/existent/file.json")
	require.NoError(t, err)
}

func TestPrometheusCollector_StateEmptyFile(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	// Создаем пустой файл
	tempFile := filepath.Join(os.TempDir(), "empty_state.json")
	defer os.Remove(tempFile)

	err := os.WriteFile(tempFile, []byte(""), 0600)
	require.NoError(t, err)

	// Загрузка пустого файла должна завершиться с ошибкой
	err = collector.LoadState(tempFile)
	require.Error(t, err)
}

func TestPrometheusCollector_StatePermissionDenied(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	// Добавляем данные чтобы было что сохранять
	data := domain.TelemetryData{
		NodeID:       "test_node",
		BatteryLevel: floatPtr(50.0),
		Timestamp:    time.Now(),
	}
	err := collector.CollectTelemetry(data)
	require.NoError(t, err)

	// Попытка сохранить в несуществующую директорию
	err = collector.SaveState("/nonexistent/directory/state.json")
	require.Error(t, err)
}

func TestPrometheusCollector_StateLargeDataset(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	// Добавляем много данных
	const numNodes = 100
	for i := 0; i < numNodes; i++ {
		data := domain.TelemetryData{
			NodeID:             fmt.Sprintf("node_%d", i),
			BatteryLevel:       floatPtr(float64(i % 100)),
			Temperature:        floatPtr(float64(20 + i%50)),
			RelativeHumidity:   floatPtr(float64(30 + i%70)),
			BarometricPressure: floatPtr(float64(1000 + i%100)),
			Voltage:            floatPtr(float64(3.0 + float64(i%20)/10)),
			Timestamp:          time.Now(),
		}
		err := collector.CollectTelemetry(data)
		require.NoError(t, err)
	}

	// Сохраняем и загружаем большой dataset
	tempFile := filepath.Join(os.TempDir(), "large_state.json")
	defer os.Remove(tempFile)

	err := collector.SaveState(tempFile)
	require.NoError(t, err)

	// Проверяем размер файла
	stat, err := os.Stat(tempFile)
	require.NoError(t, err)
	assert.Greater(t, stat.Size(), int64(1000)) // Файл должен быть больше 1KB

	// Загружаем в новый collector
	newCollector := NewPrometheusCollector()
	err = newCollector.LoadState(tempFile)
	require.NoError(t, err)
}
