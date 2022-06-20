package manager

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

type Manager struct {
	backend        string
	composeManager *ComposeManager
	// kubernetesManager *KubernetesManager
}

func NewManager(backend string) (manager *Manager) {
	manager = &Manager{backend: backend}
	manager.composeManager = NewComposeManager()
	return manager
}

func (manager *Manager) MessageHandler(client mqtt.Client, msg mqtt.Message) {
	log.Info().Str("topic", string(msg.Topic())).Msg("Manager received message")
	if manager.backend == "docker-compose" {
		manager.composeManager.MessageHandler(client, msg)
	} else if manager.backend == "kubernetes" {
		log.Info().Str("payload", string(msg.Payload())).Msg("Not Implemented")
	} else {
		log.Info().Str("payload", string(msg.Payload())).Msg("Not Implemented")
	}

}
