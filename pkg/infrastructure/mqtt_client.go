package infrastructure

import (
	"context"
	"crypto/tls"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"

	"meshtastic-exporter/pkg/domain"
	"meshtastic-exporter/pkg/logger"
)

type MQTTClient struct {
	config    domain.MQTTConfig
	processor domain.MessageProcessor
	client    mqtt.Client
	logger    zerolog.Logger
}

func NewMQTTClient(config domain.MQTTConfig, processor domain.MessageProcessor) *MQTTClient {
	return &MQTTClient{
		config:    config,
		processor: processor,
		logger:    logger.ComponentLogger("mqtt-client"),
	}
}

func (c *MQTTClient) Connect() error {
	opts := mqtt.NewClientOptions()

	broker := c.buildBrokerURL()
	opts.AddBroker(broker)
	opts.SetClientID("meshtastic-exporter-standalone")
	opts.SetKeepAlive(c.config.GetKeepAlive())
	opts.SetPingTimeout(domain.DefaultMQTTPingTimeout)
	opts.SetConnectTimeout(domain.DefaultMQTTConnTimeout)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(domain.DefaultMQTTReconnectInt)

	tlsConfig := c.config.GetTLSConfig()
	if tlsConfig.GetEnabled() {
		opts.SetTLSConfig(&tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		})
	}

	users := c.config.GetUsers()
	if len(users) > 0 {
		opts.SetUsername(users[0].GetUsername())
		opts.SetPassword(users[0].GetPassword())
	}

	opts.SetOnConnectHandler(c.onConnect)
	opts.SetConnectionLostHandler(c.onConnectionLost)

	c.client = mqtt.NewClient(opts)

	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to mqtt: %w", token.Error())
	}

	c.logger.Info().Str("broker", broker).Msg("connected to mqtt broker")
	return nil
}

func (c *MQTTClient) buildBrokerURL() string {
	tlsConfig := c.config.GetTLSConfig()
	if tlsConfig.GetEnabled() {
		return fmt.Sprintf("ssl://%s:%d", c.config.GetHost(), tlsConfig.GetPort())
	}
	return fmt.Sprintf("tcp://%s:%d", c.config.GetHost(), c.config.GetPort())
}

func (c *MQTTClient) onConnect(client mqtt.Client) {
	topics := []string{"msh/+/+/json/+/+", "msh/2/json/+/+"}
	for _, topic := range topics {
		if token := client.Subscribe(topic, 0, c.messageHandler); token.Wait() && token.Error() != nil {
			c.logger.Error().Err(token.Error()).Str("topic", topic).Msg("failed to subscribe")
		} else {
			c.logger.Info().Str("topic", topic).Msg("subscribed to topic")
		}
	}
}

func (c *MQTTClient) onConnectionLost(_ mqtt.Client, err error) {
	c.logger.Error().Err(err).Msg("connection lost")
}

func (c *MQTTClient) messageHandler(_ mqtt.Client, msg mqtt.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), domain.DefaultTimeout)
	defer cancel()

	if err := c.processor.ProcessMessage(ctx, msg.Topic(), msg.Payload()); err != nil {
		c.logger.Error().Err(err).Str("topic", msg.Topic()).Msg("message processing failed")
	}
}

func (c *MQTTClient) Disconnect() {
	if c.client != nil && c.client.IsConnected() {
		c.client.Disconnect(domain.DefaultMQTTDisconnectMs)
		c.logger.Info().Msg("disconnected from mqtt broker")
	}
}
