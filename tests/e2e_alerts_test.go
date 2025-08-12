package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"
	"meshtastic-exporter/pkg/infrastructure"
)

func TestE2E_AlertManagerWebhook(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Найти свободный порт для AlertManager
	alertPort := findFreePort(t)

	// Создать MQTT сервер (минимальный)
	server := mqtt.New(&mqtt.Options{InlineClient: false})

	err := server.AddHook(new(auth.AllowHook), &auth.Options{
		Ledger: &auth.Ledger{
			Auth: auth.AuthRules{{Allow: true}},
		},
	})
	require.NoError(t, err)

	// Создать фабрику для тестов
	f := factory.NewDefaultFactory()

	// Создать MeshtasticHook с AlertManager
	hookConfig := hooks.MeshtasticHookConfig{
		ServerAddr:   fmt.Sprintf("localhost:%d", alertPort),
		EnableHealth: false,
		AlertPath:    "/alerts/webhook",
	}
	alertHook := hooks.NewMeshtasticHook(hookConfig, f)

	err = server.AddHook(alertHook, nil)
	require.NoError(t, err)

	// Запустить сервер
	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	// Дождаться запуска AlertManager HTTP сервера
	time.Sleep(200 * time.Millisecond)

	// Отправить AlertManager webhook
	alertPayload := infrastructure.AlertPayload{
		Alerts: []infrastructure.AlertItem{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
					"instance":  "server-01",
				},
				Annotations: map[string]string{
					"summary":     "CPU usage is above 90%",
					"description": "Server server-01 has high CPU usage",
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(alertPayload)
	require.NoError(t, err)

	// Отправить POST запрос к webhook
	webhookURL := fmt.Sprintf("http://localhost:%d/alerts/webhook", alertPort)
	// #nosec G107 - URL is constructed from test port, safe for testing
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Проверить ответ webhook - это главное, что нужно протестировать
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Проверить, что в логах есть сообщение об отправке алерта
	// (в реальном тесте можно было бы проверить MQTT сообщения через mock или spy)
	t.Log("AlertManager webhook processed successfully")

	// Очистка
	server.Close()
}

func TestE2E_AlertManagerMultipleAlerts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	alertPort := findFreePort(t)

	server := mqtt.New(&mqtt.Options{InlineClient: false})

	err := server.AddHook(new(auth.AllowHook), &auth.Options{
		Ledger: &auth.Ledger{
			Auth: auth.AuthRules{{Allow: true}},
		},
	})
	require.NoError(t, err)

	f := factory.NewDefaultFactory()

	hookConfig := hooks.MeshtasticHookConfig{
		ServerAddr:   fmt.Sprintf("localhost:%d", alertPort),
		EnableHealth: false,
		AlertPath:    "/webhook",
	}
	alertHook := hooks.NewMeshtasticHook(hookConfig, f)

	err = server.AddHook(alertHook, nil)
	require.NoError(t, err)

	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	time.Sleep(200 * time.Millisecond)

	// Отправить множественные алерты
	alertPayload := infrastructure.AlertPayload{
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
				Status: "resolved",
				Labels: map[string]string{
					"alertname": "ServiceDown",
					"severity":  "critical",
				},
				Annotations: map[string]string{
					"summary": "Service is back online",
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(alertPayload)
	require.NoError(t, err)

	webhookURL := fmt.Sprintf("http://localhost:%d/webhook", alertPort)
	// #nosec G107 - URL is constructed from test port, safe for testing
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	t.Log("Multiple alerts processed successfully")

	server.Close()
}

func TestE2E_AlertManagerInvalidRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	alertPort := findFreePort(t)

	server := mqtt.New(&mqtt.Options{InlineClient: false})

	err := server.AddHook(new(auth.AllowHook), &auth.Options{
		Ledger: &auth.Ledger{
			Auth: auth.AuthRules{{Allow: true}},
		},
	})
	require.NoError(t, err)

	f := factory.NewDefaultFactory()

	hookConfig := hooks.MeshtasticHookConfig{
		ServerAddr:   fmt.Sprintf("localhost:%d", alertPort),
		EnableHealth: false,
		AlertPath:    "/alerts",
	}
	alertHook := hooks.NewMeshtasticHook(hookConfig, f)

	err = server.AddHook(alertHook, nil)
	require.NoError(t, err)

	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	webhookURL := fmt.Sprintf("http://localhost:%d/alerts", alertPort)

	testCases := []struct {
		name           string
		method         string
		contentType    string
		body           string
		expectedStatus int
	}{
		{
			name:           "GET request should be rejected",
			method:         "GET",
			contentType:    "",
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid JSON should be rejected",
			method:         "POST",
			contentType:    "application/json",
			body:           `{invalid json`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Valid JSON should be accepted",
			method:         "POST",
			contentType:    "application/json",
			body:           `{"alerts":[]}`,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, webhookURL, strings.NewReader(tc.body))
			require.NoError(t, err)

			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}

	server.Close()
}
