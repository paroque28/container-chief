package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type ClientConfigurations struct {
	Mqtt MqttConfigurations
}

type DaemonConfigurations struct {
	Mqtt          MqttConfigurations
	Backend       BackendConfigurations
	DockerCompose DockerComposeConfigurations
}

type MqttConfigurations struct {
	CHIEF_MQTT_BROKER string
	CHIEF_MQTT_PORT   int
}
type BackendConfigurations struct {
	CHIEF_BACKEND   string
	CHIEF_DEVICE_ID string
}
type DockerComposeConfigurations struct {
	CHIEF_SERVICES_PATH string
}

func setupViper() {
	viper.SetConfigName("default")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/chiefd")
	viper.AutomaticEnv()
	viper.SetConfigType("yml")
}

func FetchDaemonConfiguration() (configuration DaemonConfigurations, err error) {
	setupViper()
	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Msg("Failed to read configuration")
		return configuration, err
	}
	viper.SetDefault("backend.CHIEF_BACKEND", "docker-compose")
	viper.SetDefault("dockerCompose.CHIEF_SERVICES_PATH", "/var/lib/chief/servicess")
	err = viper.Unmarshal(&configuration)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal configuration")
		return configuration, err
	}
	log.Info().Str("backend", configuration.Backend.CHIEF_BACKEND).Msg("Configuration")
	log.Info().Str("broker", configuration.Mqtt.CHIEF_MQTT_BROKER).Int("port", configuration.Mqtt.CHIEF_MQTT_PORT).Msg("MQTT configuration")
	return configuration, err
}

func FetchClientConfiguration() (configuration ClientConfigurations, err error) {
	setupViper()
	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Msg("Failed to read configuration")
		return configuration, err
	}
	err = viper.Unmarshal(&configuration)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal configuration")
		return configuration, err
	}
	log.Debug().Str("broker", configuration.Mqtt.CHIEF_MQTT_BROKER).Int("port", configuration.Mqtt.CHIEF_MQTT_PORT).Msg("MQTT configuration")
	return configuration, err
}
