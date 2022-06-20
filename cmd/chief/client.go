package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/akamensky/argparse"
	"github.com/google/uuid"
	"github.com/paroque28/container-chief/pkg/config"
	"github.com/paroque28/container-chief/pkg/messages"
	"github.com/paroque28/container-chief/pkg/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func main() {
	// Log setup
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	viper.SetConfigName("config/default")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Read application configuration
	viper.SetConfigType("yml")
	var configuration config.ClientConfigurations
	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Msg("Failed to read configuration")
	}
	err := viper.Unmarshal(&configuration)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal configuration")
	}
	log.Info().Str("broker", configuration.Mqtt.CHIEF_MQTT_BROKER).Int("port", configuration.Mqtt.CHIEF_MQTT_PORT).Msg("MQTT configuration")

	// Parser
	parser := argparse.NewParser(filepath.Base(os.Args[0]), "Container Chief Client")
	input := parser.String("i", "input", &argparse.Options{Required: true, Help: "Path to the input file"})
	err = parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	// Read input file
	data, err := messages.ReadYaml(*input)
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading YAML file")
		return
	}
	log.Debug().Interface("data", data).Msg("YAML file read")
	body, err := json.Marshal(data)
	if err != nil {
		log.Fatal().Err(err).Msg("Error marshalling YAML file")
		return
	}

	// Connect to MQTT
	svr := new(server.Server)
	svr.Connect(configuration.Mqtt.CHIEF_MQTT_BROKER, configuration.Mqtt.CHIEF_MQTT_PORT, nil, "client-"+uuid.New().String())

	svr.Publish(messages.ConfigurationsTopic, messages.QoS, string(body))
}
