module github.com/paroque28/container-chief

go 1.13

require (
	github.com/akamensky/argparse v1.3.1
	github.com/compose-spec/compose-go v1.2.7
	github.com/docker/cli v20.10.12+incompatible
	github.com/docker/compose/v2 v2.6.0
	github.com/eclipse/paho.mqtt.golang v1.4.1
	github.com/google/uuid v1.3.0
	github.com/rs/zerolog v1.27.0
	github.com/spf13/viper v1.12.0
	gopkg.in/yaml.v3 v3.0.0
)

replace (
	github.com/docker/cli => github.com/docker/cli v20.10.3-0.20220309205733-2b52f62e9627+incompatible
	github.com/docker/docker => github.com/docker/docker v20.10.3-0.20220309172631-83b51522df43+incompatible

	// For k8s dependencies, we use a replace directive, to prevent them being
	// upgraded to the version specified in containerd, which is not relevant to the
	// version needed.
	// See https://github.com/docker/buildx/pull/948 for details.
	// https://github.com/docker/buildx/blob/v0.8.1/go.mod#L62-L64
	k8s.io/api => k8s.io/api v0.22.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.4
	k8s.io/client-go => k8s.io/client-go v0.22.4
)
