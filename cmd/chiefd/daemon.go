package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/akamensky/argparse"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/paroque28/container-chief/pkg/config"
	"github.com/paroque28/container-chief/pkg/manager"
	"github.com/paroque28/container-chief/pkg/messages"
	"github.com/paroque28/container-chief/pkg/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Log setup
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Read arguments
	parser := argparse.NewParser(filepath.Base(os.Args[0]), "Container Chief Daemon")
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}
	configuration, err := config.FetchDaemonConfiguration()
	if err != nil {
		log.Error().Err(err).Msg("Failed to read configuration")
		return
	}

	// Start Manager
	mgr := manager.NewManager(configuration)
	messagePubHandler := func(client mqtt.Client, msg mqtt.Message) {
		if msg == nil {
			log.Error().Msg("Message is empty")
			return
		}
		mgr.MessageHandler(client, msg)
	}

	// Connect to MQTT
	svr := new(server.Server)
	svr.Connect(configuration.Mqtt.CHIEF_MQTT_BROKER, configuration.Mqtt.CHIEF_MQTT_PORT, messagePubHandler, "client-"+uuid.New().String())

	svr.Subscribe(messages.GetConfigurationsTopic(configuration.Backend.CHIEF_DEVICE_ID), messages.QoS, nil)
	<-make(chan int)
}
