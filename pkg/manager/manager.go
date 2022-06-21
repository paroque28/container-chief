package manager

import (
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/paroque28/container-chief/pkg/config"
	"github.com/paroque28/container-chief/pkg/messages"
	"github.com/rs/zerolog/log"
)

type Manager struct {
	mu             sync.Mutex
	backend        string
	composeManager *ComposeManager
	// kubernetesManager *KubernetesManager
}

func NewManager(configuration config.DaemonConfigurations) (manager *Manager) {
	manager = &Manager{backend: configuration.Backend.CHIEF_BACKEND}
	manager.composeManager = NewComposeManager(configuration)
	return manager
}

func (manager *Manager) MessageHandler(client mqtt.Client, msg mqtt.Message) {
	manager.mu.Lock()
	log.Info().Str("topic", string(msg.Topic())).Msg("Manager received message")
	configuration, err := messages.JsonToConfiguration(msg.Payload())
	if err != nil {
		log.Err(err).Msg("Failed to parse message")
		return
	}
	if manager.backend == "docker-compose" {
		manager.composeManager.MessageHandler(configuration)
	} else if manager.backend == "kubernetes" {
		log.Error().Msg("Kubernetes not Implemented")
	} else {
		log.Error().Msg("Not Implemented")
	}
	manager.mu.Unlock()

}
