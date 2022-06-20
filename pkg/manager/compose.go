package manager

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

type ComposeManager struct {
}

func NewComposeManager() *ComposeManager {
	return &ComposeManager{}
}

func (manager *ComposeManager) MessageHandler(client mqtt.Client, msg mqtt.Message) {
	log.Info().Str("topic", string(msg.Topic())).Msg("ComposeManager received message")
}
