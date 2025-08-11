package tests

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/config"
	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"
)

func TestE2E_StatePersistence(t *testing.T) {
	stateFile := "test_e2e_state.json"
	defer os.Remove(stateFile)

	configYAML := `
logging:
  level: "info"
mqtt:
  host: localhost
  port: 1883
hook:
  listen: "localhost:0"
  prometheus:
    path: "/metrics"
    state:
      file: "` + stateFile + `"
`

	// Создаем временный конфиг файл
	configFile := "test_config.yaml"
	defer os.Remove(configFile)
	require.NoError(t, os.WriteFile(configFile, []byte(configYAML), 0600))

	// Загружаем конфигурацию
	cfg, err := config.LoadUnifiedConfig(configFile)
	require.NoError(t, err)

	// Создаем factory и hook
	f := factory.NewFactory(cfg)
	hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
		ServerAddr:  "localhost:0",
		TopicPrefix: "msh/",
	}, f)

	// Инициализируем hook
	require.NoError(t, hook.Init(nil))

	// Симулируем получение MQTT сообщения
	telemetryJSON := `{
		"from": 123456789,
		"type": "telemetry",
		"payload": {
			"battery_level": 85,
			"temperature": 23.5,
			"voltage": 4.1
		}
	}`

	ctx := context.Background()
	processor := f.CreateMessageProcessor()
	require.NoError(t, processor.ProcessMessage(ctx, "msh/2/json/LongFast/!broadcast", []byte(telemetryJSON)))

	// Принудительно сохраняем состояние
	collector := f.CreateMetricsCollector()
	require.NoError(t, collector.SaveState(stateFile))

	// Проверяем, что файл создался
	_, err = os.Stat(stateFile)
	require.NoError(t, err)

	// Создаем новый factory и загружаем состояние
	newF := factory.NewFactory(cfg)
	newCollector := newF.CreateMetricsCollector()

	// Проверяем восстановленные метрики
	metricFamilies, err := newCollector.GetRegistry().Gather()
	require.NoError(t, err)

	foundBattery := false
	foundTemp := false

	for _, mf := range metricFamilies {
		switch mf.GetName() {
		case domain.MetricBatteryLevel:
			foundBattery = true
			assert.Equal(t, 85.0, mf.GetMetric()[0].GetGauge().GetValue())
		case domain.MetricTemperature:
			foundTemp = true
			assert.Equal(t, 23.5, mf.GetMetric()[0].GetGauge().GetValue())
		}
	}

	assert.True(t, foundBattery, "battery metric not restored")
	assert.True(t, foundTemp, "temperature metric not restored")

	// Завершаем hook
	require.NoError(t, hook.Shutdown(context.Background()))
}

func TestE2E_StateDisabled(t *testing.T) {
	configYAML := `
logging:
  level: "info"
mqtt:
  host: localhost
  port: 1883
hook:
  listen: "localhost:0"
  prometheus:
    path: "/metrics"
`

	configFile := "test_config_no_state.yaml"
	defer os.Remove(configFile)
	require.NoError(t, os.WriteFile(configFile, []byte(configYAML), 0600))

	cfg, err := config.LoadUnifiedConfig(configFile)
	require.NoError(t, err)

	f := factory.NewFactory(cfg)
	collector := f.CreateMetricsCollector()

	// Должно работать без ошибок когда state отключен
	require.NoError(t, collector.SaveState(""))
	require.NoError(t, collector.LoadState(""))
}
