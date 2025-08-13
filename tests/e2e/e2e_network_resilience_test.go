package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"
)

func TestE2E_NetworkResilience_MQTTServerRestart(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	mqttPort := findFreePort(t)
	httpPort := findFreePortExcluding(t, mqttPort)

	// Создаем первый MQTT сервер и отправляем начальное сообщение
	server1, client := setupInitialServer(t, mqttPort, httpPort)
	initialMetrics := getMetrics(t, httpPort)
	assert.Contains(t, initialMetrics, `meshtastic_battery_level_percent{node_id="123456789"} 85.5`)

	// Перезапускаем сервер
	server2 := restartMQTTServer(t, server1, mqttPort, httpPort)

	// Проверяем переподключение клиента
	ensureClientReconnected(t, client)
	waitForHTTPServer(t, httpPort)

	// Отправляем сообщение после переподключения
	sendMessageAfterReconnect(t, client, httpPort)

	client.Disconnect(250)
	server2.Close()
}

func setupInitialServer(t *testing.T, mqttPort, httpPort int) (*mqtt.Server, paho.Client) {
	server := createMQTTServer(t, mqttPort, httpPort)
	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	waitForHTTPServer(t, httpPort)

	client := createMQTTClient(t, mqttPort, "resilience-test")
	sendTelemetryMessage(t, client, 123456789, 85.5)
	time.Sleep(200 * time.Millisecond)

	return server, client
}

func restartMQTTServer(t *testing.T, oldServer *mqtt.Server, mqttPort, httpPort int) *mqtt.Server {
	oldServer.Close()
	time.Sleep(300 * time.Millisecond)

	newServer := createMQTTServer(t, mqttPort, httpPort)
	go func() {
		if err := newServer.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	return newServer
}

func ensureClientReconnected(t *testing.T, client paho.Client) {
	t.Logf("Waiting for client reconnection...")
	connected := waitForConnection(t, client)
	if !connected {
		connected = attemptManualReconnect(t, client)
	}
	require.True(t, connected, "Client should reconnect after server restart")
}

func waitForConnection(t *testing.T, client paho.Client) bool {
	for i := 0; i < 20; i++ {
		if client.IsConnected() {
			t.Logf("Client reconnected after %d attempts", i+1)
			return true
		}
		t.Logf("Attempt %d: client not connected yet", i+1)
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func attemptManualReconnect(t *testing.T, client paho.Client) bool {
	t.Logf("Client failed to reconnect, trying manual reconnect...")
	token := client.Connect()
	if token.WaitTimeout(5*time.Second) && token.Error() == nil {
		t.Logf("Manual reconnect successful")
		return true
	}
	t.Logf("Manual reconnect failed: %v", token.Error())
	return false
}

func sendMessageAfterReconnect(t *testing.T, client paho.Client, httpPort int) {
	t.Logf("Sending telemetry message after reconnect...")
	ensureClientConnected(t, client)

	telemetryMsg := map[string]interface{}{
		"from": uint64(987654321),
		"type": "telemetry",
		"payload": map[string]interface{}{
			"battery_level": 75.0,
		},
	}

	telemetryJSON, err := json.Marshal(telemetryMsg)
	require.NoError(t, err)
	t.Logf("Publishing message: %s", string(telemetryJSON))

	publishWithRetry(t, client, telemetryJSON, httpPort)

	finalMetrics := getMetrics(t, httpPort)
	t.Logf("Final metrics after restart: %s", finalMetrics)
	assert.Contains(t, finalMetrics, `meshtastic_battery_level_percent{node_id="987654321"} 75`)
}

func ensureClientConnected(t *testing.T, client paho.Client) {
	if !client.IsConnected() {
		t.Logf("Client disconnected before sending message, reconnecting...")
		token := client.Connect()
		require.True(t, token.WaitTimeout(5*time.Second))
		require.NoError(t, token.Error())
	}
}

func publishWithRetry(t *testing.T, client paho.Client, telemetryJSON []byte, httpPort int) {
	for attempt := 1; attempt <= 3; attempt++ {
		t.Logf("Publishing attempt %d", attempt)
		token := client.Publish("msh/2/2/json/LongFast/!broadcast", 0, false, telemetryJSON)
		require.True(t, token.WaitTimeout(5*time.Second), "Failed to publish message on attempt %d", attempt)
		require.NoError(t, token.Error(), "Error publishing message on attempt %d", attempt)
		t.Logf("Message published successfully on attempt %d", attempt)

		time.Sleep(500 * time.Millisecond)

		metricsCheck := getMetrics(t, httpPort)
		if contains(metricsCheck, `meshtastic_battery_level_percent{node_id="987654321"} 75`) {
			t.Logf("Message processed successfully on attempt %d", attempt)
			break
		}
		t.Logf("Attempt %d: message not processed yet, metrics: %s", attempt, metricsCheck)
	}
}

func TestE2E_NetworkResilience_HTTPServerRestart(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	httpPort := findFreePort(t)

	f := factory.NewDefaultFactory()
	hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
		ServerAddr:   fmt.Sprintf(":%d", httpPort),
		EnableHealth: true,
	}, f)

	// Запускаем HTTP сервер
	require.NoError(t, hook.Init(nil))
	waitForHTTPServer(t, httpPort)

	// Проверяем, что сервер работает
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", httpPort))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Останавливаем сервер
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, hook.Shutdown(ctx))

	// Проверяем, что сервер недоступен
	time.Sleep(100 * time.Millisecond)
	_, err = http.Get(fmt.Sprintf("http://localhost:%d/health", httpPort))
	assert.Error(t, err)

	// Перезапускаем сервер
	require.NoError(t, hook.Init(nil))
	waitForHTTPServer(t, httpPort)

	// Проверяем, что сервер снова работает
	resp, err = http.Get(fmt.Sprintf("http://localhost:%d/health", httpPort))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	require.NoError(t, hook.Shutdown(context.Background()))
}

func TestE2E_NetworkResilience_ConcurrentConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	mqttPort := findFreePort(t)
	httpPort := findFreePortExcluding(t, mqttPort)

	server := createMQTTServer(t, mqttPort, httpPort)
	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	waitForHTTPServer(t, httpPort)

	// Создаем множественных клиентов
	const numClients = 10
	var wg sync.WaitGroup
	clients := make([]paho.Client, numClients)

	for i := 0; i < numClients; i++ {
		clients[i] = createMQTTClient(t, mqttPort, fmt.Sprintf("client-%d", i))
	}

	// Отправляем сообщения параллельно
	wg.Add(numClients)
	for i := 0; i < numClients; i++ {
		go func(clientIdx int) {
			defer wg.Done()
			nodeID := uint64(1000000 + clientIdx)
			batteryLevel := 80.0 + float64(clientIdx)
			sendTelemetryMessage(t, clients[clientIdx], nodeID, batteryLevel)
		}(i)
	}

	wg.Wait()
	time.Sleep(500 * time.Millisecond)

	// Проверяем метрики от всех клиентов
	metrics := getMetrics(t, httpPort)
	for i := 0; i < numClients; i++ {
		nodeID := 1000000 + i
		batteryLevel := 80.0 + float64(i)
		expectedMetric := fmt.Sprintf(`meshtastic_battery_level_percent{node_id="%d"} %g`, nodeID, batteryLevel)
		assert.Contains(t, metrics, expectedMetric)
	}

	// Отключаем всех клиентов
	for _, client := range clients {
		client.Disconnect(250)
	}
	server.Close()
}

func TestE2E_NetworkResilience_SlowNetwork(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	mqttPort := findFreePort(t)
	httpPort := findFreePortExcluding(t, mqttPort)

	server := createMQTTServer(t, mqttPort, httpPort)
	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	waitForHTTPServer(t, httpPort)

	// Создаем клиента с коротким таймаутом
	opts := paho.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://localhost:%d", mqttPort))
	opts.SetClientID("slow-network-test")
	opts.SetConnectTimeout(1 * time.Second)
	opts.SetWriteTimeout(1 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetKeepAlive(2 * time.Second)

	client := paho.NewClient(opts)
	token := client.Connect()
	require.True(t, token.WaitTimeout(5*time.Second))
	require.NoError(t, token.Error())

	// Отправляем большое количество сообщений быстро
	const numMessages = 50
	for i := 0; i < numMessages; i++ {
		nodeID := uint64(2000000 + i)
		batteryLevel := 50.0 + float64(i%50)
		sendTelemetryMessage(t, client, nodeID, batteryLevel)
		time.Sleep(10 * time.Millisecond) // Небольшая задержка
	}

	time.Sleep(1 * time.Second)

	// Проверяем, что хотя бы часть сообщений обработана
	metrics := getMetrics(t, httpPort)
	messageCount := 0
	for i := 0; i < numMessages; i++ {
		nodeID := 2000000 + i
		if contains(metrics, fmt.Sprintf(`node_id="%d"`, nodeID)) {
			messageCount++
		}
	}

	assert.Greater(t, messageCount, numMessages/2, "Должно быть обработано хотя бы половина сообщений")

	client.Disconnect(250)
	server.Close()
}

func TestE2E_NetworkResilience_PortExhaustion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Занимаем много портов
	var listeners []net.Listener
	defer func() {
		for _, l := range listeners {
			l.Close()
		}
	}()

	// Занимаем 100 портов
	for i := 0; i < 100; i++ {
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			break
		}
		listeners = append(listeners, listener)
	}

	// Пытаемся найти свободный порт
	mqttPort := findFreePort(t)
	httpPort := findFreePortExcluding(t, mqttPort)

	// Создаем сервер на найденных портах
	server := createMQTTServer(t, mqttPort, httpPort)
	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	waitForHTTPServer(t, httpPort)

	client := createMQTTClient(t, mqttPort, "port-exhaustion-test")
	sendTelemetryMessage(t, client, 3000000, 90.0)
	time.Sleep(200 * time.Millisecond)

	metrics := getMetrics(t, httpPort)
	assert.Contains(t, metrics, `meshtastic_battery_level_percent{node_id="3000000"} 90`)

	client.Disconnect(250)
	server.Close()
}

// Вспомогательные функции

func createMQTTServer(t *testing.T, mqttPort, httpPort int) *mqtt.Server {
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

	return server
}

func createMQTTClient(t *testing.T, mqttPort int, clientID string) paho.Client {
	opts := paho.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://localhost:%d", mqttPort))
	opts.SetClientID(clientID)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(100 * time.Millisecond)
	opts.SetMaxReconnectInterval(1 * time.Second)

	client := paho.NewClient(opts)
	token := client.Connect()
	require.True(t, token.WaitTimeout(5*time.Second))
	require.NoError(t, token.Error())

	return client
}

func sendTelemetryMessage(t *testing.T, client paho.Client, nodeID uint64, batteryLevel float64) {
	telemetryMsg := map[string]interface{}{
		"from": nodeID,
		"type": "telemetry",
		"payload": map[string]interface{}{
			"battery_level": batteryLevel,
		},
	}

	telemetryJSON, err := json.Marshal(telemetryMsg)
	require.NoError(t, err)

	token := client.Publish("msh/2/2/json/LongFast/!broadcast", 0, false, telemetryJSON)
	require.True(t, token.WaitTimeout(5*time.Second))
	require.NoError(t, token.Error())
}

func getMetrics(t *testing.T, httpPort int) string {
	metricsURL := fmt.Sprintf("http://localhost:%d/metrics", httpPort)
	resp, err := http.Get(metricsURL)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return string(body)
}

func contains(s, substr string) bool {
	return findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
