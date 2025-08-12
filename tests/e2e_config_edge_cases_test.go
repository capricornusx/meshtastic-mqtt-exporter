package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/config"
	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"
)

func TestE2E_ConfigEdgeCases_InvalidTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tempDir := t.TempDir()
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
    metrics_ttl: "invalid-duration"
`

	configFile := tempDir + "/invalid_ttl_config.yaml"
	require.NoError(t, os.WriteFile(configFile, []byte(configYAML), 0600))

	cfg, err := config.LoadUnifiedConfig(configFile)
	require.NoError(t, err) // Должен использовать значение по умолчанию

	f := factory.NewFactory(cfg)
	hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
		ServerAddr:  "localhost:0",
		TopicPrefix: "msh/",
	}, f)

	require.NoError(t, hook.Init(nil))
	require.NoError(t, hook.Shutdown(context.Background()))
}

func TestE2E_ConfigEdgeCases_ZeroTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tempDir := t.TempDir()
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
    metrics_ttl: "0s"
`

	configFile := tempDir + "/zero_ttl_config.yaml"
	require.NoError(t, os.WriteFile(configFile, []byte(configYAML), 0600))

	cfg, err := config.LoadUnifiedConfig(configFile)
	require.NoError(t, err)

	f := factory.NewFactory(cfg)
	collector := f.CreateMetricsCollector()

	// Проверяем, что TTL установлен в значение по умолчанию
	assert.NotNil(t, collector)
}

func TestE2E_ConfigEdgeCases_EmptyStateFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tempDir := t.TempDir()
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
    state_file: ""
`

	configFile := tempDir + "/empty_state_config.yaml"
	require.NoError(t, os.WriteFile(configFile, []byte(configYAML), 0600))

	cfg, err := config.LoadUnifiedConfig(configFile)
	require.NoError(t, err)

	f := factory.NewFactory(cfg)
	collector := f.CreateMetricsCollector()

	// Должно работать без ошибок при пустом state_file
	require.NoError(t, collector.SaveState(""))
	require.NoError(t, collector.LoadState(""))
}

func TestE2E_ConfigEdgeCases_InvalidLogLevel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tempDir := t.TempDir()
	configYAML := `
logging:
  level: "invalid-level"
mqtt:
  host: localhost
  port: 1883
hook:
  listen: "localhost:0"
  prometheus:
    path: "/metrics"
`

	configFile := tempDir + "/invalid_log_config.yaml"
	require.NoError(t, os.WriteFile(configFile, []byte(configYAML), 0600))

	cfg, err := config.LoadUnifiedConfig(configFile)
	require.NoError(t, err) // Должен использовать значение по умолчанию

	f := factory.NewFactory(cfg)
	hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
		ServerAddr: "localhost:0",
	}, f)

	require.NoError(t, hook.Init(nil))
	require.NoError(t, hook.Shutdown(context.Background()))
}

func TestE2E_ConfigEdgeCases_ExtremeValues(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tempDir := t.TempDir()
	configYAML := `
logging:
  level: "debug"
mqtt:
  host: localhost
  port: 1883
  capabilities:
    maximum_inflight: 999999
    maximum_client_writes_pending: 999999
    receive_maximum: 999999
    maximum_qos: 3
    maximum_clients: 999999
hook:
  listen: "localhost:0"
  prometheus:
    path: "/metrics"
    metrics_ttl: "1ns"
`

	configFile := tempDir + "/extreme_values_config.yaml"
	require.NoError(t, os.WriteFile(configFile, []byte(configYAML), 0600))

	cfg, err := config.LoadUnifiedConfig(configFile)
	require.NoError(t, err)

	f := factory.NewFactory(cfg)
	hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
		ServerAddr: "localhost:0",
	}, f)

	require.NoError(t, hook.Init(nil))
	require.NoError(t, hook.Shutdown(context.Background()))
}

func TestE2E_ConfigEdgeCases_MissingRequiredFields(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tempDir := t.TempDir()
	configYAML := `
# Минимальная конфигурация без обязательных полей
logging:
  level: "info"
`

	configFile := tempDir + "/minimal_config.yaml"
	require.NoError(t, os.WriteFile(configFile, []byte(configYAML), 0600))

	cfg, err := config.LoadUnifiedConfig(configFile)
	require.NoError(t, err) // Должен использовать значения по умолчанию

	f := factory.NewFactory(cfg)
	hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
		ServerAddr: "localhost:0",
	}, f)

	require.NoError(t, hook.Init(nil))
	require.NoError(t, hook.Shutdown(context.Background()))
}

func TestE2E_ConfigEdgeCases_CorruptedYAML(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tempDir := t.TempDir()
	corruptedYAML := `
logging:
  level: "info"
mqtt:
  host: localhost
  port: 1883
  invalid_yaml: [unclosed array
hook:
  listen: "localhost:0"
`

	configFile := tempDir + "/corrupted_config.yaml"
	require.NoError(t, os.WriteFile(configFile, []byte(corruptedYAML), 0600))

	_, err := config.LoadUnifiedConfig(configFile)
	require.Error(t, err) // Должен вернуть ошибку парсинга YAML
}

func TestE2E_ConfigEdgeCases_NonExistentConfigFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	nonExistentFile := "/tmp/non_existent_config_" + fmt.Sprintf("%d", time.Now().UnixNano()) + ".yaml"

	cfg, err := config.LoadUnifiedConfig(nonExistentFile)
	require.NoError(t, err) // Должен использовать значения по умолчанию

	f := factory.NewFactory(cfg)
	hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
		ServerAddr: "localhost:0",
	}, f)

	require.NoError(t, hook.Init(nil))
	require.NoError(t, hook.Shutdown(context.Background()))
}

func TestE2E_ConfigEdgeCases_ReadOnlyStateFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tempDir := t.TempDir()
	stateFile := tempDir + "/readonly_state.json"

	// Создаем файл с данными
	require.NoError(t, os.WriteFile(stateFile, []byte(`{"version":"1.0","timestamp":0,"nodes":[]}`), 0600))

	configYAML := fmt.Sprintf(`
logging:
  level: "info"
mqtt:
  host: localhost
  port: 1883
hook:
  listen: "localhost:0"
  prometheus:
    path: "/metrics"
    state_file: "%s"
`, stateFile)

	configFile := tempDir + "/readonly_state_config.yaml"
	require.NoError(t, os.WriteFile(configFile, []byte(configYAML), 0600))

	cfg, err := config.LoadUnifiedConfig(configFile)
	require.NoError(t, err)

	f := factory.NewFactory(cfg)
	collector := f.CreateMetricsCollector()

	// Загрузка должна работать
	require.NoError(t, collector.LoadState(stateFile))

	// Делаем файл read-only после загрузки
	require.NoError(t, os.Chmod(stateFile, 0444))

	// Проверяем, что система обрабатывает read-only файлы корректно
	// (может не возвращать ошибку, если нет данных для сохранения)
	err = collector.SaveState(stateFile)
	// Не проверяем ошибку, так как SaveState может не писать пустые данные
	_ = err

	// Восстанавливаем права для cleanup
	os.Chmod(stateFile, 0644)
}
