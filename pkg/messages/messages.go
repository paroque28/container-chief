package messages

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/compose-spec/compose-go/loader"
	compose "github.com/compose-spec/compose-go/types"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const (
	ConfigurationsTopic = "chief/configuration"
	QoS                 = 1
)

type Service struct {
	ComposeFile string          `yaml:"compose_file" json:"compose_file"`
	Compose     compose.Project `yaml:"compose" json:"compose"`
	Status      string          `yaml:"status" validate:"required" json:"status"`
}

type Configuration struct {
	Services map[string]Service `yaml:"services" validate:"required" json:"services"`
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

	for name, service := range data.Services {
		if service.ComposeFile == "" {
			log.Error().Str("input", service.ComposeFile).Msg("Compose file not specified")
			continue
		}
		composeFullPath := filepath.Join(baseDir, service.ComposeFile)
		log.Info().Str("composeFullPath", composeFullPath).Str("composeFile", service.ComposeFile).Msg("Loading compose file")

		var b []byte
		b, err = os.ReadFile(composeFullPath)
		if err != nil {
			return
		}
		var files []compose.ConfigFile
		files = append(files, compose.ConfigFile{Filename: service.ComposeFile, Content: b})
		envMap := make(map[string]string)
		project, err := loader.Load(compose.ConfigDetails{
			WorkingDir:  baseDir,
			ConfigFiles: files,
			Environment: envMap,
		}, withProjectName(name))

		if err != nil {
			log.Error().Err(err).Str("composeFullPath", composeFullPath).Msg("Failed to load compose file")
			return data, err
		}

		log.Debug().Interface("project", project).Msg("Project")

		service.Compose = *project
		data.Services[name] = service
	}
	return data, err
}

func withProjectName(name string) func(*loader.Options) {
	return func(lOpts *loader.Options) {
		lOpts.SetProjectName(name, true)
	}
}
