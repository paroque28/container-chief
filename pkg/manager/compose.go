package manager

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/paroque28/container-chief/pkg/config"
	"github.com/paroque28/container-chief/pkg/messages"
	"github.com/rs/zerolog/log"
)

type ComposeManager struct {
	servicesPath string
}

func NewComposeManager(configuration config.DaemonConfigurations) *ComposeManager {
	os.MkdirAll(configuration.DockerCompose.CHIEF_SERVICES_PATH, os.ModePerm)
	return &ComposeManager{servicesPath: configuration.DockerCompose.CHIEF_SERVICES_PATH}
}

func (manager *ComposeManager) MessageHandler(configuration messages.Configuration) (err error) {
	for service_name, service := range configuration.Projects {
		// Create folder for service
		servicepath := filepath.Join(manager.servicesPath, service_name)
		if err := os.MkdirAll(servicepath, os.ModePerm); err != nil {
			log.Error().Err(err).Msg("Failed to create service directory")
			continue
		}
		// Create service file
		err = ioutil.WriteFile(servicepath+"/docker-compose.yml", []byte(service.Compose), 0644)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create service file")
			continue
		}

	}
	return err

}
