package hooks

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meshtastic-exporter/pkg/factory"
)

func TestNewMeshtasticHook(t *testing.T) {
	f := factory.NewDefaultFactory() // Mock factory
	hook := NewMeshtasticHook(MeshtasticHookConfig{
		ServerAddr:  ":9090",
		TopicPrefix: "test/",
	}, f)

	assert.Equal(t, ":9090", hook.config.ServerAddr)
	assert.Equal(t, "test/", hook.config.TopicPrefix)
	assert.Equal(t, 30*time.Minute, hook.config.MetricsTTL)
}

func TestNewMeshtasticHookSimple(t *testing.T) {
	f := factory.NewDefaultFactory() // Mock factory
	hook := NewMeshtasticHook(MeshtasticHookConfig{
		ServerAddr:   "", // Disabled for test
		EnableHealth: true,
	}, f)

	assert.Equal(t, "", hook.config.ServerAddr)
	assert.Equal(t, "msh/", hook.config.TopicPrefix)
	assert.True(t, hook.config.EnableHealth)
}

func TestMeshtasticHook_ID(t *testing.T) {
	f := factory.NewDefaultFactory() // Mock factory
	hook := NewMeshtasticHook(MeshtasticHookConfig{ServerAddr: ""}, f)
	assert.Equal(t, "meshtastic", hook.ID())
}

func TestMeshtasticHook_Provides(t *testing.T) {
	f := factory.NewDefaultFactory() // Mock factory
	hook := NewMeshtasticHook(MeshtasticHookConfig{ServerAddr: ""}, f)

	assert.True(t, hook.Provides(mqtt.OnPublish))
	assert.True(t, hook.Provides(mqtt.OnConnect))
	assert.True(t, hook.Provides(mqtt.OnDisconnect))
	assert.False(t, hook.Provides(mqtt.OnSubscribe))
}

func TestMeshtasticHook_OnPublish(t *testing.T) {
	f := factory.NewDefaultFactory() // Mock factory
	hook := NewMeshtasticHook(MeshtasticHookConfig{ServerAddr: ""}, f)

	tests := []struct {
		name     string
		topic    string
		payload  map[string]interface{}
		expected bool // whether message should be processed
	}{
		{
			name:  "valid_meshtastic_message",
			topic: "msh/test/data",
			payload: map[string]interface{}{
				"from": float64(123456),
				"type": "telemetry",
				"payload": map[string]interface{}{
					"battery_level": float64(85.5),
					"temperature":   float64(23.4),
				},
			},
			expected: true,
		},
		{
			name:     "non_meshtastic_topic",
			topic:    "other/topic",
			payload:  map[string]interface{}{"test": "data"},
			expected: false,
		},
		{
			name:     "invalid_json_ignored",
			topic:    "msh/test",
			payload:  nil, // Will be marshaled as null, then fail to unmarshal
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payload []byte
			if tt.payload != nil {
				payload, _ = json.Marshal(tt.payload)
			} else {
				payload = []byte("invalid json")
			}

			pk := packets.Packet{
				TopicName: tt.topic,
				Payload:   payload,
			}

			result, err := hook.OnPublish(nil, pk)
			require.NoError(t, err)
			assert.Equal(t, tt.topic, result.TopicName)
		})
	}
}

func TestMeshtasticHook_Init(t *testing.T) {
	f := factory.NewDefaultFactory() // Mock factory
	hook := NewMeshtasticHook(MeshtasticHookConfig{ServerAddr: ""}, f)
	err := hook.Init(nil)
	require.NoError(t, err)
}

func TestOnConnect(t *testing.T) {
	f := factory.NewDefaultFactory() // Mock factory
	hook := NewMeshtasticHook(MeshtasticHookConfig{ServerAddr: ""}, f)

	client := &mqtt.Client{ID: "test-client"}
	err := hook.OnConnect(client, packets.Packet{})
	require.NoError(t, err)
}

func TestOnDisconnect(t *testing.T) {
	f := factory.NewDefaultFactory() // Mock factory
	hook := NewMeshtasticHook(MeshtasticHookConfig{ServerAddr: ""}, f)

	client := &mqtt.Client{ID: "test-client"}
	// Should not panic
	hook.OnDisconnect(client, nil, false)
	hook.OnDisconnect(client, nil, true)
}

func TestMeshtasticHook_Shutdown(t *testing.T) {
	f := factory.NewDefaultFactory() // Mock factory
	hook := NewMeshtasticHook(MeshtasticHookConfig{ServerAddr: ""}, f)
	err := hook.Shutdown(context.TODO())
	require.NoError(t, err)
}
