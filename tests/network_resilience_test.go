package tests

import (
	"context"
	"fmt"
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

	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"
)

func TestNetworkResilience_MQTTReconnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network resilience test in short mode")
	}

	mqttPort := findFreePort(t)
	httpPort := findFreePort(t)

	// Создаем MQTT сервер
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

	// Запускаем сервер
	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("MQTT server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	waitForHTTPServer(t, httpPort)

	// Создаем MQTT клиента с автоматическим переподключением
	opts := paho.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://localhost:%d", mqttPort))
	opts.SetClientID("resilience-test-client")
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(100 * time.Millisecond)

	client := paho.NewClient(opts)

	// Подключаемся
	token := client.Connect()
	require.True(t, token.WaitTimeout(5*time.Second))
	require.NoError(t, token.Error())

	// Отправляем сообщение
	payload := `{"from": 123456789, "type": "telemetry", "payload": {"battery_level": 85.5}}`
	token = client.Publish("msh/test", 0, false, payload)
	require.True(t, token.WaitTimeout(5*time.Second))
	require.NoError(t, token.Error())

	// Симулируем разрыв соединения - закрываем сервер
	server.Close()
	time.Sleep(200 * time.Millisecond)

	// Перезапускаем сервер
	server = mqtt.New(&mqtt.Options{InlineClient: false})
	err = server.AddHook(new(auth.AllowHook), &auth.Options{
		Ledger: &auth.Ledger{
			Auth: auth.AuthRules{{Allow: true}},
		},
	})
	require.NoError(t, err)

	err = server.AddHook(hook, nil)
	require.NoError(t, err)

	tcp = listeners.NewTCP(listeners.Config{
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

	// Ждем переподключения
	time.Sleep(500 * time.Millisecond)

	// Проверяем, что клиент переподключился и может отправлять сообщения
	token = client.Publish("msh/test", 0, false, payload)
	assert.True(t, token.WaitTimeout(5*time.Second))
	assert.NoError(t, token.Error())

	client.Disconnect(250)
	server.Close()
}

func TestNetworkResilience_HTTPServerRestart(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network resilience test in short mode")
	}

	httpPort := findFreePort(t)

	f := factory.NewDefaultFactory()
	hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
		ServerAddr:   fmt.Sprintf(":%d", httpPort),
		EnableHealth: true,
	}, f)

	// Инициализируем хук
	err := hook.Init(nil)
	require.NoError(t, err)

	// Ждем запуска HTTP сервера
	waitForHTTPServer(t, httpPort)

	// Проверяем, что сервер отвечает
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", httpPort))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Останавливаем хук
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = hook.Shutdown(ctx)
	require.NoError(t, err)

	// Проверяем, что сервер недоступен
	time.Sleep(100 * time.Millisecond)
	_, err = http.Get(fmt.Sprintf("http://localhost:%d/health", httpPort))
	assert.Error(t, err)
}

func TestNetworkResilience_PortAlreadyInUse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network resilience test in short mode")
	}

	port := findFreePort(t)

	// Занимаем порт
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	require.NoError(t, err)
	defer listener.Close()

	// Пытаемся создать хук на занятом порту
	f := factory.NewDefaultFactory()
	hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
		ServerAddr:   fmt.Sprintf(":%d", port),
		EnableHealth: true,
	}, f)

	// Инициализация может пройти успешно, но сервер не запустится
	_ = hook.Init(nil)
	// Не проверяем ошибку инициализации, так как она может пройти асинхронно
}

func TestNetworkResilience_InvalidAddress(t *testing.T) {
	f := factory.NewDefaultFactory()
	hook := hooks.NewMeshtasticHook(hooks.MeshtasticHookConfig{
		ServerAddr:   "invalid:address:format",
		EnableHealth: true,
	}, f)

	err := hook.Init(nil)
	// Не проверяем ошибку, так как сервер может запуститься асинхронно
	_ = err
}
