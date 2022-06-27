package messages

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/compose-spec/compose-go/loader"
	compose "github.com/compose-spec/compose-go/types"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const (
	DevicesTopic = "chief/devices/"
	QoS          = 1
)

func GetConfigurationsTopic(device string) string {
	return DevicesTopic + device + "/configurations"
}

type Project struct {
	ComposeFile string `yaml:"compose_file" json:"compose_file"`
	Compose     string `yaml:"compose" json:"compose"`
	Status      string `yaml:"status" validate:"required" json:"status"`
}

type Configuration struct {
	Projects map[string]Project `yaml:"projects" validate:"required" json:"projects"`
}

func JsonToConfiguration(input []byte) (configuration Configuration, err error) {
	var data map[string]interface{}
	err = json.Unmarshal(input, &data)
	if err != nil {
		log.Error().Err(err).Str("input", string(input)).Msg("Failed to unmarshal JSON")
	}
	log.Debug().Interface("project", data).Msg("Parsing JSON")
	if err := json.Unmarshal(input, &configuration); err != nil {
		log.Error().Err(err).Str("input", string(input)).Msg("Failed to unmarshal JSON")
		log.Info().Interface("input", configuration).Msg("Failed to unmarshal JSON")
	}
	log.Debug().Interface("configuration", configuration).Msg("Configuration")
	return configuration, err
}

func ReadYaml(path string) (data Configuration, err error) {
	fullPath, err := filepath.Abs(path)
	baseDir := filepath.Dir(fullPath)
	if err != nil {
		return
	}
	log.Debug().Str("fullPath", fullPath).Str("path", path).Msg("Reading YAML file")
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return data, err
	}

	err = yaml.Unmarshal([]byte(file), &data)
	if err != nil {
		log.Error().Err(err).Str("input", string(file)).Msg("Failed to unmarshal YAML file")
	}
	log.Debug().Interface("data", data).Str("input", string(file)).Msg("YAML file read")

	for name, project := range data.Projects {
		if project.ComposeFile == "" {
			log.Error().Str("input", project.ComposeFile).Msg("Compose file not specified")
			continue
		}
		composeFullPath := filepath.Join(baseDir, project.ComposeFile)
		log.Debug().Str("composeFullPath", composeFullPath).Str("composeFile", project.ComposeFile).Msg("Loading compose file")

		var b []byte
		b, err = os.ReadFile(composeFullPath)
		if err != nil {
			return
		}
		var files []compose.ConfigFile
		files = append(files, compose.ConfigFile{Filename: project.ComposeFile, Content: b})
		envMap := make(map[string]string)
		// Validate Compose File
		_, err := loader.Load(compose.ConfigDetails{
			WorkingDir:  baseDir,
			ConfigFiles: files,
			Environment: envMap,
		}, withProjectName(name))

		if err != nil {
			log.Error().Err(err).Str("composeFullPath", composeFullPath).Msg("Failed to load compose file")
			return data, err
		}

		log.Debug().Interface("project", project).Msg("Project")
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal compose file")
			return data, err
		}
		project.Compose = string(b)
		data.Projects[name] = project
	}
	return data, err
}

func withProjectName(name string) func(*loader.Options) {
	return func(lOpts *loader.Options) {
		lOpts.SetProjectName(name, true)
	}
}
