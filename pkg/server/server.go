package server

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

type Server struct {
	client mqtt.Client
	Broker string
	Port   int
}

var defaultMessagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Info().Str("payload", string(msg.Payload())).Str("topic", msg.Topic()).Msg("Received message")
}

func (server *Server) Connect(broker string, port int, messagePubHandler mqtt.MessageHandler, clientID string) {
	server.Broker = broker
	server.Port = port
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", server.Broker, server.Port))
	opts.SetClientID(clientID)
	// opts.SetUsername("emqx")
	// opts.SetPassword("public")
	if messagePubHandler != nil {
		opts.SetDefaultPublishHandler(messagePubHandler)
	} else {
		opts.SetDefaultPublishHandler(defaultMessagePubHandler)
	}
	opts.OnConnect = func(client mqtt.Client) {
		server.client = client
		log.Debug().Str("broker", server.Broker).Int("port", server.Port).Str("cliendid", clientID).Msg("Connected to MQTT broker")
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		server.client = nil
		log.Err(err).Msg("Connection lost")
	}
	server.client = mqtt.NewClient(opts)
	if server.client == nil {
		log.Fatal().Msg("Failed to create client")
	}
	log.Debug().Str("broker", server.Broker).Int("port", server.Port).Msg("Connecting to MQTT broker")
	if token := server.client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

}

func (server *Server) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) {
	if server.client == nil {
		log.Fatal().Msg("Client not connected")
	}
	log.Info().Str("topic", topic).Msg("Subscribing to topic")
	token := server.client.Subscribe(topic, qos, callback)
	token.Wait()
}

func (server *Server) Publish(topic string, qos byte, payload string) {
	if server.client == nil {
		log.Fatal().Msg("Client not connected")
	}
	token := server.client.Publish(topic, qos, false, payload)
	token.Wait()
	log.Debug().Str("topic", topic).Str("payload", payload).Msg("Published message")
}
