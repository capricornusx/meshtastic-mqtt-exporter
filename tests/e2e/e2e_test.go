package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"
)

func TestE2E_MQTTToPrometheusMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Найти свободные порты с разделением
	mqttPort := findFreePort(t)
	httpPort := findFreePortExcluding(t, mqttPort)

	// Создать MQTT сервер с хуком
	server := mqtt.New(&mqtt.Options{InlineClient: false})

	// Добавить auth хук
	err := server.AddHook(new(auth.AllowHook), &auth.Options{
		Ledger: &auth.Ledger{
			Auth: auth.AuthRules{{Allow: true}},
		},
	})
	require.NoError(t, err)

	// Создать factory и хук
	f := factory.NewDefaultFactory()
	hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
		ServerAddr:   fmt.Sprintf(":%d", httpPort),
		EnableHealth: true,
		TopicPrefix:  "msh/",
	}, f)

	err = server.AddHook(hook, nil)
	require.NoError(t, err)

	// Добавить TCP listener
	tcp := listeners.NewTCP(listeners.Config{
		ID:      "tcp",
		Address: fmt.Sprintf(":%d", mqttPort),
	})
	err = server.AddListener(tcp)
	require.NoError(t, err)

	// Запустить сервер
	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	// Дождаться запуска
	time.Sleep(100 * time.Millisecond)

	// Проверить, что HTTP сервер запустился
	waitForHTTPServer(t, httpPort)

	// Создать MQTT клиента
	opts := paho.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://localhost:%d", mqttPort))
	opts.SetClientID("test-client")
	client := paho.NewClient(opts)

	token := client.Connect()
	require.True(t, token.WaitTimeout(5*time.Second))
	require.NoError(t, token.Error())

	// Отправить тестовые сообщения
	telemetryMsg := map[string]interface{}{
		"from": 123456789,
		"type": "telemetry",
		"payload": map[string]interface{}{
			"battery_level":       85.5,
			"temperature":         23.4,
			"relative_humidity":   65.2,
			"barometric_pressure": 1013.25,
			"voltage":             4.1,
		},
	}

	nodeInfoMsg := map[string]interface{}{
		"from": 123456789,
		"type": "nodeinfo",
		"payload": map[string]interface{}{
			"longname":  "E2E Test Node",
			"shortname": "E2E",
			"hardware":  1,
			"role":      1, // client_mute
		},
	}

	// Отправить телеметрию
	telemetryJSON, _ := json.Marshal(telemetryMsg)
	token = client.Publish("msh/2/2/json/LongFast/!broadcast", 0, false, telemetryJSON)
	require.True(t, token.WaitTimeout(5*time.Second))
	require.NoError(t, token.Error())

	// Отправить node info
	nodeInfoJSON, _ := json.Marshal(nodeInfoMsg)
	token = client.Publish("msh/2/2/json/LongFast/!broadcast", 0, false, nodeInfoJSON)
	require.True(t, token.WaitTimeout(5*time.Second))
	require.NoError(t, token.Error())

	// Дождаться обработки сообщений
	time.Sleep(500 * time.Millisecond)

	// Проверить метрики через HTTP
	metricsURL := fmt.Sprintf("http://localhost:%d/metrics", httpPort)
	// #nosec G107 - URL is constructed from test port, safe for testing
	resp, err := http.Get(metricsURL)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	metrics := string(body)

	// Проверить наличие метрик
	assert.Contains(t, metrics, domain.MetricBatteryLevel)
	assert.Contains(t, metrics, domain.MetricTemperature)
	assert.Contains(t, metrics, domain.MetricHumidity)
	assert.Contains(t, metrics, domain.MetricPressure)
	assert.Contains(t, metrics, domain.MetricVoltage)
	assert.Contains(t, metrics, domain.MetricMessagesTotal)
	assert.Contains(t, metrics, domain.MetricNodeInfo)
	assert.Contains(t, metrics, domain.MetricNodeLastSeen)

	// Проверить конкретные значения
	assert.Contains(t, metrics, fmt.Sprintf(`%s{node_id="123456789"} 85.5`, domain.MetricBatteryLevel))
	assert.Contains(t, metrics, fmt.Sprintf(`%s{node_id="123456789"} 23.4`, domain.MetricTemperature))
	assert.Contains(t, metrics, fmt.Sprintf(`%s{from_node="123456789",type="%s"} 1`, domain.MetricMessagesTotal, domain.MessageTypeTelemetry))
	assert.Contains(t, metrics, fmt.Sprintf(`%s{from_node="123456789",type="%s"} 1`, domain.MetricMessagesTotal, domain.MessageTypeNodeInfo))
	// Проверить, что роль преобразована в читаемое имя (1 = "client_mute")
	assert.Contains(t, metrics, `role="client_mute"`)
	assert.NotContains(t, metrics, `role="1"`)

	// Проверить health endpoint
	healthURL := fmt.Sprintf("http://localhost:%d/health", httpPort)
	// #nosec G107 - URL is constructed from test port, safe for testing
	resp, err = http.Get(healthURL)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	// Очистка
	client.Disconnect(250)
	server.Close()
}

func TestE2E_MultipleNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Найти свободные порты с разделением
	mqttPort := findFreePort(t)
	httpPort := findFreePortExcluding(t, mqttPort)

	// Создать MQTT сервер с хуком
	server := mqtt.New(&mqtt.Options{InlineClient: false})

	err := server.AddHook(new(auth.AllowHook), &auth.Options{
		Ledger: &auth.Ledger{
			Auth: auth.AuthRules{{Allow: true}},
		},
	})
	require.NoError(t, err)

	f := factory.NewDefaultFactory()
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

	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	waitForHTTPServer(t, httpPort)

	// Создать MQTT клиента
	opts := paho.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://localhost:%d", mqttPort))
	opts.SetClientID("test-client-multi")
	client := paho.NewClient(opts)

	token := client.Connect()
	require.True(t, token.WaitTimeout(5*time.Second))
	require.NoError(t, token.Error())

	// Отправить сообщения от разных узлов
	nodes := []uint64{123456789, 987654321, 555666777}

	for i, nodeID := range nodes {
		telemetryMsg := map[string]interface{}{
			"from": nodeID,
			"type": "telemetry",
			"payload": map[string]interface{}{
				"battery_level": 80.0 + float64(i*5),
				"temperature":   20.0 + float64(i*2),
			},
		}

		telemetryJSON, _ := json.Marshal(telemetryMsg)
		token = client.Publish("msh/2/2/json/LongFast/!broadcast", 0, false, telemetryJSON)
		require.True(t, token.WaitTimeout(5*time.Second))
		require.NoError(t, token.Error())
	}

	time.Sleep(500 * time.Millisecond)

	// Проверить метрики
	metricsURL := fmt.Sprintf("http://localhost:%d/metrics", httpPort)
	// #nosec G107 - URL is constructed from test port, safe for testing
	resp, err := http.Get(metricsURL)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	metrics := string(body)

	// Проверить метрики для всех узлов
	for i, nodeID := range nodes {
		expectedBattery := 80.0 + float64(i*5)
		expectedTemp := 20.0 + float64(i*2)

		assert.Contains(t, metrics, fmt.Sprintf(`meshtastic_battery_level_percent{node_id="%d"} %g`, nodeID, expectedBattery))
		assert.Contains(t, metrics, fmt.Sprintf(`meshtastic_temperature_celsius{node_id="%d"} %g`, nodeID, expectedTemp))
	}

	client.Disconnect(250)
	server.Close()
}

func findFreePort(t *testing.T) int {
	return findFreePortExcluding(t, 0)
}

func findFreePortExcluding(t *testing.T, excludePort int) int {
	for i := 0; i < 20; i++ {
		// #nosec G102 - Binding to all interfaces is needed for testing
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			continue
		}
		port := listener.Addr().(*net.TCPAddr).Port
		listener.Close()

		if port == excludePort {
			continue
		}

		// Проверяем, что порт действительно свободен
		time.Sleep(10 * time.Millisecond)
		conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			// Порт свободен
			return port
		}
		conn.Close()
	}
	t.Fatal("could not find free port after 20 attempts")
	return 0
}

func waitForHTTPServer(t *testing.T, port int) {
	url := fmt.Sprintf("http://localhost:%d/health", port)
	client := &http.Client{Timeout: 500 * time.Millisecond}

	for i := 0; i < 50; i++ {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("HTTP server did not start on port %d within 5s", port)
}
