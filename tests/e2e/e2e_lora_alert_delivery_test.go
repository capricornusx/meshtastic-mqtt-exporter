package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"
	"meshtastic-exporter/pkg/infrastructure"
)

var testMQTTDownlinkTopic = infrastructure.GetDefaultMQTTDownlinkTopic()

type LoRaMessageCapture struct {
	mu       sync.RWMutex
	messages []CapturedMessage
}

type CapturedMessage struct {
	Topic   string
	Payload string
	QoS     byte
	Retain  bool
}

func (c *LoRaMessageCapture) AddMessage(topic, payload string, qos byte, retain bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = append(c.messages, CapturedMessage{
		Topic:   topic,
		Payload: payload,
		QoS:     qos,
		Retain:  retain,
	})
}

func (c *LoRaMessageCapture) GetMessages() []CapturedMessage {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]CapturedMessage, len(c.messages))
	copy(result, c.messages)
	return result
}

func (c *LoRaMessageCapture) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = nil
}

func TestE2E_LoRaAlertDelivery_CriticalAlert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	mqttPort := findFreePort(t)
	httpPort := findFreePortExcluding(t, mqttPort)
	capture := &LoRaMessageCapture{}

	// Создаем MQTT сервер с перехватчиком сообщений
	server := createMQTTServerWithCapture(t, mqttPort, httpPort, capture)
	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	waitForHTTPServer(t, httpPort)

	// Отправляем критический алерт
	criticalAlert := infrastructure.AlertPayload{
		Alerts: []infrastructure.AlertItem{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "NodeOffline",
					"severity":  "critical",
					"node_id":   "123456789",
				},
				Annotations: map[string]string{
					"summary":     "Node 123456789 is offline",
					"description": "Critical node has been offline for 5 minutes",
				},
			},
		},
	}

	// Отправляем webhook
	sendAlertWebhook(t, httpPort, criticalAlert)
	time.Sleep(500 * time.Millisecond)

	// Проверяем, что сообщение отправлено в LoRa сеть
	messages := capture.GetMessages()
	require.Greater(t, len(messages), 0, "Должно быть отправлено хотя бы одно LoRa сообщение")

	// Проверяем формат сообщения
	found := false
	for _, msg := range messages {
		if msg.Topic == testMQTTDownlinkTopic {
			assert.Contains(t, msg.Payload, "NodeOffline")
			assert.Contains(t, msg.Payload, "firing")
			assert.Contains(t, msg.Payload, "sendtext")
			found = true
			break
		}
	}
	assert.True(t, found, "Критический алерт должен быть отправлен в MQTT топик")

	server.Close()
}

func TestE2E_LoRaAlertDelivery_WarningAlert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	mqttPort := findFreePort(t)
	httpPort := findFreePortExcluding(t, mqttPort)
	capture := &LoRaMessageCapture{}

	server := createMQTTServerWithCapture(t, mqttPort, httpPort, capture)
	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	waitForHTTPServer(t, httpPort)

	// Отправляем предупреждающий алерт
	warningAlert := infrastructure.AlertPayload{
		Alerts: []infrastructure.AlertItem{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighBatteryUsage",
					"severity":  "warning",
					"node_id":   "987654321",
				},
				Annotations: map[string]string{
					"summary":     "High battery usage detected",
					"description": "Battery level is below 20%",
				},
			},
		},
	}

	sendAlertWebhook(t, httpPort, warningAlert)
	time.Sleep(500 * time.Millisecond)

	messages := capture.GetMessages()
	require.Greater(t, len(messages), 0, "Должно быть отправлено хотя бы одно LoRa сообщение")

	// Проверяем, что warning алерт отправлен правильно
	found := false
	for _, msg := range messages {
		if msg.Topic == testMQTTDownlinkTopic {
			assert.Contains(t, msg.Payload, "HighBatteryUsage")
			assert.Contains(t, msg.Payload, "firing")
			assert.Contains(t, msg.Payload, "sendtext")
			found = true
			break
		}
	}
	assert.True(t, found, "Warning алерт должен быть отправлен в LoRa сеть")

	server.Close()
}

func TestE2E_LoRaAlertDelivery_MultipleAlerts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	mqttPort := findFreePort(t)
	httpPort := findFreePortExcluding(t, mqttPort)
	capture := &LoRaMessageCapture{}

	server := createMQTTServerWithCapture(t, mqttPort, httpPort, capture)
	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	waitForHTTPServer(t, httpPort)

	// Отправляем множественные алерты
	multipleAlerts := infrastructure.AlertPayload{
		Alerts: []infrastructure.AlertItem{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "DiskSpaceLow",
					"severity":  "warning",
				},
				Annotations: map[string]string{
					"summary": "Disk space is low",
				},
			},
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "ServiceDown",
					"severity":  "critical",
				},
				Annotations: map[string]string{
					"summary": "Service is down",
				},
			},
			{
				Status: "resolved",
				Labels: map[string]string{
					"alertname": "MemoryUsageHigh",
					"severity":  "warning",
				},
				Annotations: map[string]string{
					"summary": "Memory usage is back to normal",
				},
			},
		},
	}

	sendAlertWebhook(t, httpPort, multipleAlerts)
	time.Sleep(500 * time.Millisecond)

	messages := capture.GetMessages()
	require.Greater(t, len(messages), 0, "Должны быть отправлены LoRa сообщения")

	// Проверяем, что все алерты обработаны
	alertNames := []string{"DiskSpaceLow", "ServiceDown", "MemoryUsageHigh"}
	for _, alertName := range alertNames {
		found := false
		for _, msg := range messages {
			if contains(msg.Payload, alertName) {
				found = true
				break
			}
		}
		assert.True(t, found, fmt.Sprintf("Алерт %s должен быть отправлен", alertName))
	}

	server.Close()
}

func TestE2E_LoRaAlertDelivery_AlertFormatting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	mqttPort := findFreePort(t)
	httpPort := findFreePortExcluding(t, mqttPort)
	capture := &LoRaMessageCapture{}

	server := createMQTTServerWithCapture(t, mqttPort, httpPort, capture)
	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	waitForHTTPServer(t, httpPort)

	// Алерт с длинным описанием для проверки форматирования
	longAlert := infrastructure.AlertPayload{
		Alerts: []infrastructure.AlertItem{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "VeryLongAlertNameThatShouldBeTruncated",
					"severity":  "critical",
					"instance":  "very-long-instance-name-that-might-be-truncated",
				},
				Annotations: map[string]string{
					"summary":     "This is a very long summary that should be truncated to fit LoRa message size limits",
					"description": "This is an extremely long description that definitely exceeds the typical LoRa message size limits and should be handled appropriately by the alert formatting system",
				},
			},
		},
	}

	sendAlertWebhook(t, httpPort, longAlert)
	time.Sleep(500 * time.Millisecond)

	messages := capture.GetMessages()
	require.Greater(t, len(messages), 0, "Должно быть отправлено LoRa сообщение")

	// Проверяем, что сообщение не превышает разумные размеры для LoRa
	for _, msg := range messages {
		if msg.Topic == testMQTTDownlinkTopic {
			assert.Less(t, len(msg.Payload), 240, "LoRa сообщение должно быть меньше 240 байт")
			assert.Contains(t, msg.Payload, "VeryLongAlert") // Часть имени должна остаться
			break
		}
	}

	server.Close()
}

func TestE2E_LoRaAlertDelivery_ConcurrentAlerts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	mqttPort := findFreePort(t)
	httpPort := findFreePortExcluding(t, mqttPort)
	capture := &LoRaMessageCapture{}

	server := createMQTTServerWithCapture(t, mqttPort, httpPort, capture)
	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	waitForHTTPServer(t, httpPort)

	// Отправляем множественные алерты параллельно
	const numAlerts = 10
	var wg sync.WaitGroup

	wg.Add(numAlerts)
	for i := 0; i < numAlerts; i++ {
		go func(alertIdx int) {
			defer wg.Done()

			alert := infrastructure.AlertPayload{
				Alerts: []infrastructure.AlertItem{
					{
						Status: "firing",
						Labels: map[string]string{
							"alertname": fmt.Sprintf("ConcurrentAlert%d", alertIdx),
							"severity":  "warning",
						},
						Annotations: map[string]string{
							"summary": fmt.Sprintf("Concurrent alert number %d", alertIdx),
						},
					},
				},
			}

			sendAlertWebhook(t, httpPort, alert)
		}(i)
	}

	wg.Wait()
	time.Sleep(1 * time.Second)

	messages := capture.GetMessages()
	require.Greater(t, len(messages), 0, "Должны быть отправлены LoRa сообщения")

	// Проверяем, что все алерты обработаны
	processedAlerts := 0
	for i := 0; i < numAlerts; i++ {
		alertName := fmt.Sprintf("ConcurrentAlert%d", i)
		for _, msg := range messages {
			if contains(msg.Payload, alertName) {
				processedAlerts++
				break
			}
		}
	}

	assert.Greater(t, processedAlerts, numAlerts/2, "Должно быть обработано хотя бы половина параллельных алертов")

	server.Close()
}

// Вспомогательные функции

func createMQTTServerWithCapture(t *testing.T, mqttPort, httpPort int, capture *LoRaMessageCapture) *mqtt.Server {
	server := mqtt.New(&mqtt.Options{InlineClient: true})

	err := server.AddHook(new(auth.AllowHook), &auth.Options{
		Ledger: &auth.Ledger{
			Auth: auth.AuthRules{{Allow: true}},
		},
	})
	require.NoError(t, err)

	// Добавляем хук для перехвата сообщений
	captureHook := &MessageCaptureHook{capture: capture}
	err = server.AddHook(captureHook, nil)
	require.NoError(t, err)

	f := factory.NewDefaultFactory()

	// Создаем хук с AlertSender, который использует MQTT сервер
	hook := hooks.NewMeshtasticHookWithMQTT(hooks.MeshtasticHookConfig{
		ServerAddr:   fmt.Sprintf(":%d", httpPort),
		EnableHealth: true,
		AlertPath:    "/alerts/webhook",
		TopicPrefix:  "msh/",
	}, f, server)

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

func sendAlertWebhook(t *testing.T, httpPort int, alert infrastructure.AlertPayload) {
	payloadBytes, err := json.Marshal(alert)
	require.NoError(t, err)

	webhookURL := fmt.Sprintf("http://localhost:%d/alerts/webhook", httpPort)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// MessageCaptureHook перехватывает все MQTT сообщения
type MessageCaptureHook struct {
	mqtt.HookBase
	capture *LoRaMessageCapture
}

func (h *MessageCaptureHook) ID() string {
	return "message-capture"
}

func (h *MessageCaptureHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnPublish,
	}, []byte{b})
}

func (h *MessageCaptureHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	// Перехватываем только исходящие сообщения в MQTT топик
	if pk.TopicName == testMQTTDownlinkTopic {
		h.capture.AddMessage(pk.TopicName, string(pk.Payload), pk.FixedHeader.Qos, pk.FixedHeader.Retain)
	}
	return pk, nil
}

func (h *MessageCaptureHook) Init(config any) error {
	return nil
}

func (h *MessageCaptureHook) Stop() error {
	return nil
}

func (h *MessageCaptureHook) Shutdown(ctx context.Context) error {
	return nil
}
