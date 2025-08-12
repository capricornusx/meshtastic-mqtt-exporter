package adapters

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigAdapter_Validate_Success(t *testing.T) {
	config := NewConfigAdapter(
		MQTTConfigAdapter{
			Host: "localhost",
			Port: 1883,
		},
		PrometheusConfigAdapter{
			Listen: "0.0.0.0:8100",
			Path:   "/metrics",
		},
		AlertManagerConfigAdapter{
			Listen: "0.0.0.0:8100",
			Path:   "/alerts",
		},
	)

	err := config.Validate()

	assert.NoError(t, err)
}

func TestConfigAdapter_Validate_EmptyMQTTHost(t *testing.T) {
	config := NewConfigAdapter(
		MQTTConfigAdapter{
			Host: "",
			Port: 1883,
		},
		PrometheusConfigAdapter{},
		AlertManagerConfigAdapter{},
	)

	err := config.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MQTT host cannot be empty")
}

func TestConfigAdapter_Validate_InvalidMQTTPort(t *testing.T) {
	testCases := []struct {
		name string
		port int
	}{
		{"zero port", 0},
		{"negative port", -1},
		{"port too high", 65536},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NewConfigAdapter(
				MQTTConfigAdapter{
					Host: "localhost",
					Port: tc.port,
				},
				PrometheusConfigAdapter{},
				AlertManagerConfigAdapter{},
			)

			err := config.Validate()

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid MQTT port")
		})
	}
}

func TestConfigAdapter_Validate_InvalidPrometheusPort(t *testing.T) {
	config := NewConfigAdapter(
		MQTTConfigAdapter{
			Host: "localhost",
			Port: 1883,
		},
		PrometheusConfigAdapter{
			Listen: "",
		},
		AlertManagerConfigAdapter{},
	)

	err := config.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prometheus listen address cannot be empty")
}

func TestConfigAdapter_Validate_InvalidAlertManagerPort(t *testing.T) {
	config := NewConfigAdapter(
		MQTTConfigAdapter{
			Host: "localhost",
			Port: 1883,
		},
		PrometheusConfigAdapter{},
		AlertManagerConfigAdapter{
			Listen: "",
		},
	)

	err := config.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prometheus listen address cannot be empty")
}

func TestConfigAdapter_GetMethods(t *testing.T) {
	mqttConfig := MQTTConfigAdapter{
		Host:            "test-host",
		Port:            1883,
		Timeout:         30 * time.Second,
		KeepAlive:       60 * time.Second,
		MaxInflight:     1024,
		MaxQueued:       1000,
		ReceiveMaximum:  512,
		MaxQoS:          2,
		RetainAvailable: true,
		MessageExpiry:   24 * time.Hour,
		MaxClients:      1000,
		Users: []UserAuthAdapter{
			{Username: "user1", Password: "pass1"},
		},
		TLSConfig: TLSConfigAdapter{
			Enabled: true,
			Port:    8883,
		},
	}

	prometheusConfig := PrometheusConfigAdapter{
		Listen:     "prom-host:8100",
		Path:       "/metrics",
		MetricsTTL: 30 * time.Minute,
	}

	alertConfig := AlertManagerConfigAdapter{
		Listen: "alert-host:8080",
		Path:   "/alerts",
	}

	config := NewConfigAdapter(mqttConfig, prometheusConfig, alertConfig)

	mqtt := config.GetMQTTConfig()
	assert.Equal(t, "test-host", mqtt.GetHost())
	assert.Equal(t, 1883, mqtt.GetPort())
	assert.True(t, mqtt.GetTLSConfig().GetEnabled())
	assert.Equal(t, 30*time.Second, mqtt.GetTimeout())
	assert.Equal(t, 60*time.Second, mqtt.GetKeepAlive())
	assert.Equal(t, 1024, mqtt.GetMaxInflight())
	assert.Equal(t, 1000, mqtt.GetMaxQueued())
	assert.Equal(t, 512, mqtt.GetReceiveMaximum())
	assert.Equal(t, 2, mqtt.GetMaxQoS())
	assert.True(t, mqtt.GetRetainAvailable())
	assert.Equal(t, int64(86400), mqtt.GetMessageExpiry()) // 24h in seconds
	assert.Equal(t, 1000, mqtt.GetMaxClients())
	assert.Len(t, mqtt.GetUsers(), 1)
	assert.Equal(t, "user1", mqtt.GetUsers()[0].GetUsername())

	prometheus := config.GetPrometheusConfig()
	assert.Equal(t, "prom-host:8100", prometheus.GetListen())
	assert.Equal(t, "/metrics", prometheus.GetPath())
	assert.Equal(t, 30*time.Minute, prometheus.GetMetricsTTL())

	alert := config.GetAlertManagerConfig()
	assert.Equal(t, "alert-host:8080", alert.GetListen())
	assert.Equal(t, "/alerts", alert.GetPath())
}
