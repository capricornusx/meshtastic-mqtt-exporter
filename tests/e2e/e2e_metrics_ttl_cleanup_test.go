package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/config"
	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"
)

func TestE2E_MetricsTTLCleanup_BasicCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tempDir := t.TempDir()
	stateFile := tempDir + "/ttl_test_state.json"

	// Конфигурация с коротким TTL для быстрого тестирования
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
    metrics_ttl: "2s"
    state_file: "%s"
`, stateFile)

	configFile := tempDir + "/ttl_config.yaml"
	require.NoError(t, os.WriteFile(configFile, []byte(configYAML), 0600))

	cfg, err := config.LoadUnifiedConfig(configFile)
	require.NoError(t, err)

	mqttPort := findFreePort(t)
	httpPort := findFreePortExcluding(t, mqttPort)

	// Создаем MQTT сервер с коротким TTL
	server := createMQTTServerWithTTL(t, mqttPort, httpPort, cfg)
	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	waitForHTTPServer(t, httpPort)

	client := createMQTTClient(t, mqttPort, "ttl-test-client")

	// Отправляем телеметрию
	sendTelemetryMessage(t, client, 123456789, 85.5)
	time.Sleep(200 * time.Millisecond)

	// Проверяем, что метрика присутствует
	metrics := getMetrics(t, httpPort)
	assert.Contains(t, metrics, `meshtastic_battery_level_percent{node_id="123456789"} 85.5`)

	// Ждем истечения TTL + небольшой буфер
	time.Sleep(3 * time.Second)

	// Проверяем, что метрика удалена (для environmental метрик)
	_ = getMetrics(t, httpPort)
	// Батарея не должна быть удалена, так как она не environmental метрика
	// Но если мы отправим temperature, она должна быть удалена

	// Отправляем environmental метрику
	sendEnvironmentalTelemetry(t, client, 123456789, 25.5)
	time.Sleep(200 * time.Millisecond)

	// Проверяем наличие temperature метрики
	metrics = getMetrics(t, httpPort)
	assert.Contains(t, metrics, `meshtastic_temperature_celsius{node_id="123456789"} 25.5`)

	// Ждем истечения TTL
	time.Sleep(3 * time.Second)

	// Проверяем, что temperature метрика удалена
	metrics = getMetrics(t, httpPort)
	assert.NotContains(t, metrics, `meshtastic_temperature_celsius{node_id="123456789"}`)

	client.Disconnect(250)
	server.Close()
}

func TestE2E_MetricsTTLCleanup_MultipleNodes(t *testing.T) {
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
    metrics_ttl: "3s"
`

	configFile := tempDir + "/multi_node_ttl_config.yaml"
	require.NoError(t, os.WriteFile(configFile, []byte(configYAML), 0600))

	cfg, err := config.LoadUnifiedConfig(configFile)
	require.NoError(t, err)

	mqttPort := findFreePort(t)
	httpPort := findFreePortExcluding(t, mqttPort)

	server := createMQTTServerWithTTL(t, mqttPort, httpPort, cfg)
	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	waitForHTTPServer(t, httpPort)

	client := createMQTTClient(t, mqttPort, "multi-node-ttl-test")

	// Отправляем environmental метрики от разных узлов
	nodes := []uint64{111111111, 222222222, 333333333}
	for i, nodeID := range nodes {
		temp := 20.0 + float64(i*5)
		sendEnvironmentalTelemetry(t, client, nodeID, temp)
		time.Sleep(50 * time.Millisecond) // Небольшая задержка между отправками
	}
	time.Sleep(300 * time.Millisecond)

	// Проверяем наличие всех метрик
	metrics := getMetrics(t, httpPort)
	for i, nodeID := range nodes {
		temp := 20.0 + float64(i*5)
		expectedMetric := fmt.Sprintf(`meshtastic_temperature_celsius{node_id="%d"} %g`, nodeID, temp)
		assert.Contains(t, metrics, expectedMetric)
	}

	// Ждем 2 секунды, затем обновляем метрику только для первого узла
	time.Sleep(2 * time.Second)
	sendEnvironmentalTelemetry(t, client, nodes[0], 30.0)
	time.Sleep(300 * time.Millisecond)

	// Ждем еще 2 секунды для истечения TTL старых метрик
	time.Sleep(2 * time.Second)

	// Проверяем, что метрики для узлов 2 и 3 удалены, а для узла 1 - обновлена
	metrics = getMetrics(t, httpPort)
	assert.Contains(t, metrics, fmt.Sprintf(`meshtastic_temperature_celsius{node_id="%d"} 30`, nodes[0]))
	assert.NotContains(t, metrics, fmt.Sprintf(`meshtastic_temperature_celsius{node_id="%d"}`, nodes[1]))
	assert.NotContains(t, metrics, fmt.Sprintf(`meshtastic_temperature_celsius{node_id="%d"}`, nodes[2]))

	client.Disconnect(250)
	server.Close()
}

func TestE2E_MetricsTTLCleanup_StatePersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tempDir := t.TempDir()
	stateFile := tempDir + "/ttl_persistence_state.json"

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
    metrics_ttl: "5s"
    state_file: "%s"
`, stateFile)

	configFile := tempDir + "/ttl_persistence_config.yaml"
	require.NoError(t, os.WriteFile(configFile, []byte(configYAML), 0600))

	cfg, err := config.LoadUnifiedConfig(configFile)
	require.NoError(t, err)

	mqttPort := findFreePort(t)
	httpPort := findFreePortExcluding(t, mqttPort)

	// Первый запуск - создаем метрики
	server1 := createMQTTServerWithTTL(t, mqttPort, httpPort, cfg)
	go func() {
		if err := server1.Serve(); err != nil {
			t.Logf("MQTT server1 error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	waitForHTTPServer(t, httpPort)

	client1 := createMQTTClient(t, mqttPort, "ttl-persistence-test-1")

	// Отправляем метрики
	sendEnvironmentalTelemetry(t, client1, 444444444, 22.5)
	sendTelemetryMessage(t, client1, 444444444, 90.0)
	time.Sleep(200 * time.Millisecond)

	// Проверяем метрики
	metrics := getMetrics(t, httpPort)
	assert.Contains(t, metrics, `meshtastic_temperature_celsius{node_id="444444444"} 22.5`)
	assert.Contains(t, metrics, `meshtastic_battery_level_percent{node_id="444444444"} 90`)

	// Останавливаем первый сервер
	client1.Disconnect(250)
	server1.Close()
	time.Sleep(200 * time.Millisecond)

	// Проверяем, что состояние сохранено
	_, err = os.Stat(stateFile)
	require.NoError(t, err)

	// Второй запуск - восстанавливаем состояние
	server2 := createMQTTServerWithTTL(t, mqttPort, httpPort, cfg)
	go func() {
		if err := server2.Serve(); err != nil {
			t.Logf("MQTT server2 error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	waitForHTTPServer(t, httpPort)

	// Проверяем восстановленные метрики
	metrics = getMetrics(t, httpPort)
	assert.Contains(t, metrics, `meshtastic_temperature_celsius{node_id="444444444"} 22.5`)
	assert.Contains(t, metrics, `meshtastic_battery_level_percent{node_id="444444444"} 90`)

	server2.Close()
}

func TestE2E_MetricsTTLCleanup_ZeroTTL(t *testing.T) {
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

	mqttPort := findFreePort(t)
	httpPort := findFreePortExcluding(t, mqttPort)

	server := createMQTTServerWithTTL(t, mqttPort, httpPort, cfg)
	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	waitForHTTPServer(t, httpPort)

	client := createMQTTClient(t, mqttPort, "zero-ttl-test")

	// Отправляем environmental метрику
	sendEnvironmentalTelemetry(t, client, 555555555, 18.0)
	time.Sleep(200 * time.Millisecond)

	// Проверяем наличие метрики
	metrics := getMetrics(t, httpPort)
	assert.Contains(t, metrics, `meshtastic_temperature_celsius{node_id="555555555"} 18`)

	// Ждем некоторое время - метрика не должна быть удалена при TTL=0
	time.Sleep(2 * time.Second)

	metrics = getMetrics(t, httpPort)
	assert.Contains(t, metrics, `meshtastic_temperature_celsius{node_id="555555555"} 18`)

	client.Disconnect(250)
	server.Close()
}

func TestE2E_MetricsTTLCleanup_PartialCleanup(t *testing.T) {
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
    metrics_ttl: "3s"
`

	configFile := tempDir + "/partial_cleanup_config.yaml"
	require.NoError(t, os.WriteFile(configFile, []byte(configYAML), 0600))

	cfg, err := config.LoadUnifiedConfig(configFile)
	require.NoError(t, err)

	mqttPort := findFreePort(t)
	httpPort := findFreePortExcluding(t, mqttPort)

	server := createMQTTServerWithTTL(t, mqttPort, httpPort, cfg)
	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	waitForHTTPServer(t, httpPort)

	client := createMQTTClient(t, mqttPort, "partial-cleanup-test")

	// Отправляем разные типы метрик
	sendComplexTelemetry(t, client, 666666666)
	time.Sleep(200 * time.Millisecond)

	// Проверяем наличие всех метрик
	metrics := getMetrics(t, httpPort)
	assert.Contains(t, metrics, `meshtastic_temperature_celsius{node_id="666666666"}`)
	assert.Contains(t, metrics, `meshtastic_humidity_percent{node_id="666666666"}`)
	assert.Contains(t, metrics, `meshtastic_pressure_hpa{node_id="666666666"}`)
	assert.Contains(t, metrics, `meshtastic_battery_level_percent{node_id="666666666"}`)

	// Ждем истечения TTL (TTL=3s + cleanup interval=1.5s + buffer)
	time.Sleep(5 * time.Second)

	// Проверяем, что только environmental метрики удалены
	metrics = getMetrics(t, httpPort)
	assert.NotContains(t, metrics, `meshtastic_temperature_celsius{node_id="666666666"}`)
	assert.NotContains(t, metrics, `meshtastic_humidity_percent{node_id="666666666"}`)
	assert.NotContains(t, metrics, `meshtastic_pressure_hpa{node_id="666666666"}`)
	// Батарея должна остаться, так как она не environmental метрика
	assert.Contains(t, metrics, `meshtastic_battery_level_percent{node_id="666666666"}`)

	client.Disconnect(250)
	server.Close()
}

// Вспомогательные функции

func createMQTTServerWithTTL(t *testing.T, mqttPort, httpPort int, cfg domain.Config) *mqtt.Server {
	server := mqtt.New(&mqtt.Options{InlineClient: false})

	err := server.AddHook(new(auth.AllowHook), &auth.Options{
		Ledger: &auth.Ledger{
			Auth: auth.AuthRules{{Allow: true}},
		},
	})
	require.NoError(t, err)

	f := factory.NewFactory(cfg)
	hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
		ServerAddr:   fmt.Sprintf(":%d", httpPort),
		EnableHealth: true,
		TopicPrefix:  "msh/",
	}, f)

	err = server.AddHook(hook, nil)
	require.NoError(t, err)

	tcp := listeners.NewTCP(listeners.Config{
		ID:      "tcp",
		Address: fmt.Sprintf(":%d", mqttPort),
	})
	err = server.AddListener(tcp)
	require.NoError(t, err)

	return server
}

func sendEnvironmentalTelemetry(t *testing.T, client paho.Client, nodeID uint64, temperature float64) {
	telemetryMsg := map[string]interface{}{
		"from": nodeID,
		"type": "telemetry",
		"payload": map[string]interface{}{
			"temperature": temperature,
		},
	}

	telemetryJSON, err := json.Marshal(telemetryMsg)
	require.NoError(t, err)

	token := client.Publish("msh/2/2/json/LongFast/!broadcast", 0, false, telemetryJSON)
	require.True(t, token.WaitTimeout(5*time.Second))
	require.NoError(t, token.Error())
}

func sendComplexTelemetry(t *testing.T, client paho.Client, nodeID uint64) {
	telemetryMsg := map[string]interface{}{
		"from": nodeID,
		"type": "telemetry",
		"payload": map[string]interface{}{
			"battery_level":       75.0,
			"temperature":         23.5,
			"relative_humidity":   60.0,
			"barometric_pressure": 1013.25,
			"voltage":             3.8,
		},
	}

	telemetryJSON, err := json.Marshal(telemetryMsg)
	require.NoError(t, err)

	token := client.Publish("msh/2/2/json/LongFast/!broadcast", 0, false, telemetryJSON)
	require.True(t, token.WaitTimeout(5*time.Second))
	require.NoError(t, token.Error())
}
