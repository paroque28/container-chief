package manager

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/paroque28/container-chief/pkg/config"
	"github.com/paroque28/container-chief/pkg/messages"
	"github.com/rs/zerolog/log"
)

type ComposeManager struct {
	servicesPath string
	cli          command.Cli
}

func NewComposeManager(configuration config.DaemonConfigurations) *ComposeManager {
	os.MkdirAll(configuration.DockerCompose.CHIEF_SERVICES_PATH, os.ModePerm)
	cli, err := command.NewDockerCli(command.WithStandardStreams())
	cli.Initialize(flags.NewClientOptions())
	if err != nil {
		log.Error().Err(err).Msg("Failed to create docker cli")
		return nil
	}
	return &ComposeManager{servicesPath: configuration.DockerCompose.CHIEF_SERVICES_PATH, cli: cli}
}

func (manager *ComposeManager) MessageHandler(configuration messages.Configuration) (err error) {
	services, err := manager.ListDockerCompose()

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

		if strings.Contains(service.Status, "up") {
			// Start or update project
			err = manager.StartProject(servicepath + "/docker-compose.yml")
		} else if strings.Contains(service.Status, "pull") {
			// Pull project
			err = manager.PullProject(servicepath + "/docker-compose.yml")
			err = manager.StopProject(servicepath + "/docker-compose.yml")
		} else if strings.Contains(service.Status, "down") {
			// Stop project
			err = manager.StopProject(servicepath + "/docker-compose.yml")
		} else {
			log.Info().Str("service", service_name).Str("status", service.Status).Msg("Unknown status")
		}
		if err != nil {
			log.Error().Err(err).Str("service", service_name).Str("status", service.Status).Msg("Failed to start service")
			return err
		}

	}

	err = manager.removeOrphanProjects(services, configuration.Projects)

	return err

}

func (manager *ComposeManager) removeOrphanProjects(stack []api.Stack, projects map[string]messages.Project) (err error) {
	for _, composeProject := range stack {
		log.Info().Str("project", composeProject.Name).Msg("removeOrphanProjects")
		for projectName := range projects {
			if composeProject.Name == projectName {
				continue
			}
		}
		log.Info().Str("project", composeProject.Name).Msg("Remove orphan project")
		err = manager.StopProject(composeProject.ConfigFiles)
		err = manager.RemoveProject(composeProject.ConfigFiles)
		if err != nil {
			log.Error().Err(err).Msg("Failed to remove orphan project")
			return err
		}
	}
	return err
}

func withProjectName(name string) func(*loader.Options) {
	return func(lOpts *loader.Options) {
		lOpts.SetProjectName(name, true)
	}
}

func (manager *ComposeManager) readComposeFile(composeFile string) (project *types.Project, err error) {
	fullPath, err := filepath.Abs(composeFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get absolute path")
	}
	baseDir := filepath.Dir(fullPath)
	var b []byte
	b, err = os.ReadFile(fullPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read compose file")
		return project, err
	}
	var files []types.ConfigFile
	files = append(files, types.ConfigFile{Filename: composeFile, Content: b})
	envMap := make(map[string]string)
	// Read Compose File
	project, err = loader.Load(types.ConfigDetails{
		WorkingDir:  baseDir,
		ConfigFiles: files,
		Environment: envMap,
	}, withProjectName(path.Base(baseDir)))
	for i := range project.Services {
		if project.Services[i].CustomLabels == nil {
			project.Services[i].CustomLabels = make(map[string]string)
		}
		project.Services[i].CustomLabels["chief.project"] = project.Name
		project.Services[i].CustomLabels["com.docker.compose.project"] = project.Name
		project.Services[i].CustomLabels["com.docker.compose.project.config_files"] = fullPath
		project.Services[i].CustomLabels["com.docker.compose.project.working_dir"] = baseDir

	}
	log.Debug().Interface("project", project).Msg("readComposeFile")
	return project, err
}

func (manager *ComposeManager) PullProject(composeFile string) (err error) {
	ctx := context.TODO()
	project, err := manager.readComposeFile(composeFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read compose file")
		return err
	}
	opts := api.PullOptions{Quiet: false, IgnoreFailures: false}

	composeService := compose.NewComposeService(manager.cli)
	err = composeService.Pull(ctx, project, opts)
	return err
}

func (manager *ComposeManager) StartProject(composeFile string) (err error) {
	log.Info().Str("composeFile", composeFile).Msg("Start project")
	ctx := context.TODO()
	project, err := manager.readComposeFile(composeFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read compose file")
		return err
	}

	createOpts := api.CreateOptions{RemoveOrphans: true, IgnoreOrphans: true, QuietPull: false, Inherit: false, Recreate: "force"}
	startOpts := api.StartOptions{Project: project, CascadeStop: false, Wait: false}

	opts := api.UpOptions{Start: startOpts, Create: createOpts}
	log.Debug().Interface("opts", opts).Msg("Up options")
	composeService := compose.NewComposeService(manager.cli)
	log.Info().Str("composeFile", composeFile).Interface("project", *project).Msg("Creating project")
	err = composeService.Create(ctx, project, createOpts)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create project")
		return err
	}
	log.Info().Str("composeFile", composeFile).Msg("Starting project")
	err = composeService.Up(ctx, project, opts)
	return err
}

func (manager *ComposeManager) StopProject(composeFile string) (err error) {
	log.Info().Str("composeFile", composeFile).Msg("Stop project")
	ctx := context.TODO()
	project, err := manager.readComposeFile(composeFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read compose file")
		return err
	}
	opts := api.DownOptions{RemoveOrphans: true, Project: project, Images: "local"}
	composeService := compose.NewComposeService(manager.cli)
	err = composeService.Down(ctx, project.Name, opts)
	return err
}

func (manager *ComposeManager) RemoveProject(composeFile string) (err error) {
	log.Info().Str("composeFile", composeFile).Msg("Remove project")
	ctx := context.TODO()
	project, err := manager.readComposeFile(composeFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read compose file")
		return err
	}
	opts := api.RemoveOptions{Force: true}
	composeService := compose.NewComposeService(manager.cli)
	err = composeService.Remove(ctx, project.Name, opts)
	return err
}

func (manager *ComposeManager) ListDockerCompose() (stack []api.Stack, err error) {
	ctx := context.TODO()
	opts := api.ListOptions{All: true}

	composeService := compose.NewComposeService(manager.cli)
	stack, err = composeService.List(ctx, opts)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list services")
		return stack, err
	}
	log.Info().Str("version", manager.cli.Client().ClientVersion()).Interface("stack", stack).Msg("Stack results")
	return stack, err
}
