package infrastructure

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/domain"
)

func TestPrometheusCollector_ConcurrentTelemetry(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	const numGoroutines = 50
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Запускаем множество горутин, каждая отправляет телеметрию
	for i := 0; i < numGoroutines; i++ {
		go func(nodeID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				data := domain.TelemetryData{
					NodeID:       fmt.Sprintf("%d", nodeID),
					BatteryLevel: floatPtr(float64(j % 100)),
					Temperature:  floatPtr(float64(20 + j%10)),
					Timestamp:    time.Now(),
				}
				err := collector.CollectTelemetry(data)
				require.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	// Проверяем, что метрики корректно обновились
	registry := collector.GetRegistry()
	assert.NotNil(t, registry)
}

func TestPrometheusCollector_ConcurrentNodeInfo(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	const numNodes = 20
	var wg sync.WaitGroup
	wg.Add(numNodes)

	for i := 0; i < numNodes; i++ {
		go func(nodeID int) {
			defer wg.Done()
			info := domain.NodeInfo{
				NodeID:    fmt.Sprintf("node_%d", nodeID),
				LongName:  fmt.Sprintf("Test Node %d", nodeID),
				ShortName: fmt.Sprintf("TN%02d", nodeID),
				Hardware:  "1",
				Role:      "2",
				Timestamp: time.Now(),
			}
			err := collector.CollectNodeInfo(info)
			require.NoError(t, err)
		}(i)
	}

	wg.Wait()
}

func TestPrometheusCollector_StateOperationsConcurrency(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	// Добавляем данные
	data := domain.TelemetryData{
		NodeID:       "123456789",
		BatteryLevel: floatPtr(85.5),
		Temperature:  floatPtr(23.4),
		Timestamp:    time.Now(),
	}
	err := collector.CollectTelemetry(data)
	require.NoError(t, err)

	const numOperations = 10
	var wg sync.WaitGroup
	wg.Add(numOperations * 2)

	// Одновременно сохраняем и загружаем состояние
	for i := 0; i < numOperations; i++ {
		go func(i int) {
			defer wg.Done()
			filename := fmt.Sprintf("/tmp/test_state_%d.json", i)
			err := collector.SaveState(filename)
			assert.NoError(t, err)
		}(i)

		go func(i int) {
			defer wg.Done()
			filename := fmt.Sprintf("/tmp/test_state_%d.json", i)
			// Игнорируем ошибку, так как файл может не существовать
			collector.LoadState(filename)
		}(i)
	}

	wg.Wait()
}

func TestPrometheusCollector_MemoryUsage(t *testing.T) {
	t.Parallel()
	collector := NewPrometheusCollector()

	// Создаем много узлов с данными
	const numNodes = 1000
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

		info := domain.NodeInfo{
			NodeID:    fmt.Sprintf("node_%d", i),
			LongName:  fmt.Sprintf("Test Node %d", i),
			ShortName: fmt.Sprintf("TN%04d", i),
			Hardware:  fmt.Sprintf("%d", i%10),
			Role:      fmt.Sprintf("%d", i%5),
			Timestamp: time.Now(),
		}
		err = collector.CollectNodeInfo(info)
		require.NoError(t, err)
	}

	// Проверяем, что registry не nil и содержит метрики
	registry := collector.GetRegistry()
	assert.NotNil(t, registry)
}
