package commands

import (
	"fmt"
	"kool-dev/kool/core/environment"
	"kool-dev/kool/core/shell"
	"kool-dev/kool/services/cloud"
	"kool-dev/kool/services/compose"
	"os"
	"strings"

	"github.com/spf13/cobra"
	yaml3 "gopkg.in/yaml.v3"
)

// KoolCloudSetup holds handlers and functions for setting up deployment configuration
type KoolCloudSetup struct {
	DefaultKoolService

	promptSelect shell.PromptSelect
	env          environment.EnvStorage
}

// NewSetupCommand initializes new kool deploy Cobra command
func NewSetupCommand(setup *KoolCloudSetup) *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Set up local configuration files for deployment",
		RunE:  DefaultCommandRunFunction(setup),
		Args:  cobra.NoArgs,

		DisableFlagsInUseLine: true,
	}
}

// NewKoolCloudSetup factories new KoolCloudSetup instance pointer
func NewKoolCloudSetup() *KoolCloudSetup {
	return &KoolCloudSetup{
		*newDefaultKoolService(),

		shell.NewPromptSelect(),
		environment.NewEnvStorage(),
	}
}

// Execute runs the setup logic.
func (s *KoolCloudSetup) Execute(args []string) (err error) {
	var (
		composeConfig *compose.DockerComposeConfig
		serviceName   string

		deployConfig *cloud.DeployConfig = &cloud.DeployConfig{
			Services: make(map[string]*cloud.DeployConfigService),
		}

		postInstructions []func()
	)

	if !s.Shell().IsTerminal() {
		err = fmt.Errorf("setup command is not available in non-interactive mode")
		return
	}

	s.Shell().Warning("Warning: auto-setup is an experimental feature. Review all the generated configuration files before deploying.")
	s.Shell().Info("Loading docker compose configuration...")

	if composeConfig, err = compose.ParseConsolidatedDockerComposeConfig(s.env.Get("PWD")); err != nil {
		return
	}

	s.Shell().Info("Docker compose configuration loaded. Starting interactive setup:")

	for serviceName = range composeConfig.Services {
		var (
			answer string

			composeService = composeConfig.Services[serviceName]
		)

		if answer, err = s.promptSelect.Ask(fmt.Sprintf("Do you want to deploy the service container '%s'?", serviceName), []string{"Yes", "No"}); err != nil {
			return
		}

		if answer == "No" {
			s.Shell().Warning(fmt.Sprintf("Not going to deploy service container '%s'", serviceName))
			continue
		}

		s.Shell().Info(fmt.Sprintf("Setting up service container '%s' for deployment", serviceName))
		deployConfig.Services[serviceName] = &cloud.DeployConfigService{}

		// handle image/build config
		if len(composeService.Volumes) == 0 && composeService.Build == nil {
			// the simple-path - we have an image only and that is what we want to deploy
			if image, isString := (*composeService.Image).(string); isString {
				deployConfig.Services[serviceName].Image = new(string)
				*deployConfig.Services[serviceName].Image = image
			} else {
				err = fmt.Errorf("unable to parse image configuration for service '%s'", serviceName)
				return
			}
		} else {
			// OK there's something for us to build... maybe the user is already building it?
			// in case there's a build config, we'll use that
			if composeService.Build != nil {
				// if it's a string, that should be the build path...
				if build, isString := (*composeService.Build).(string); isString {
					if build != "." {
						err = fmt.Errorf("service '%s' got a build dockerfile on path '%s'. Please move to the root folder/context to be able to deploy.", serviceName, build)
						return
					}
					deployConfig.Services[serviceName].Build = new(string)
					*deployConfig.Services[serviceName].Build = "Dockerfile"
				} else if buildConfig, isMap := (*composeService.Build).(map[string]interface{}); isMap {
					if ctx, exists := buildConfig["context"].(string); exists && ctx != "." {
						err = fmt.Errorf("service '%s' got a build dockerfile on path '%s'. Please move to the root folder/context to be able to deploy.", serviceName, build)
						return
					}

					if dockerfile, exists := buildConfig["dockerfile"].(string); exists {
						deployConfig.Services[serviceName].Build = new(string)
						*deployConfig.Services[serviceName].Build = dockerfile
					} else {
						err = fmt.Errorf("could not tell Dockerfile for service '%s'", serviceName)
						return
					}
				}
			} else {
				// no build config, so we'll have to build it
				deployConfig.Services[serviceName].Build = new(string)
				*deployConfig.Services[serviceName].Build = "Dockerfile"

				postInstructions = append(postInstructions, func() {
					s.Shell().Info(fmt.Sprintf("⇒ Service '%s' needs to be built. Make sure to create the necessary Dockerfile.", serviceName))
				})
			}
		}

		// handle port/public config
		ports := composeService.Ports
		if len(ports) > 0 {
			potentialPorts := []string{}
			for i := range ports {
				mappedPorts := strings.Split(ports[i], ":")

				potentialPorts = append(potentialPorts, mappedPorts[len(mappedPorts)-1])
			}

			if len(potentialPorts) > 1 {
				if answer, err = s.promptSelect.Ask("Which port do you want to make public?", potentialPorts); err != nil {
					return
				}
			} else {
				answer = potentialPorts[0]
			}

			deployConfig.Services[serviceName].Port = new(string)
			*deployConfig.Services[serviceName].Port = answer

			public := &cloud.DeployConfigPublicEntry{}
			public.Port = new(string)
			*public.Port = answer

			deployConfig.Services[serviceName].Public = append(deployConfig.Services[serviceName].Public, public)
		}
	}

	var yaml []byte
	if yaml, err = yaml3.Marshal(deployConfig); err != nil {
		return
	}

	if err = os.WriteFile(koolDeployFile, yaml, 0644); err != nil {
		return
	}

	s.Shell().Println("")

	for _, instruction := range postInstructions {
		instruction()
	}

	s.Shell().Println("")
	s.Shell().Println("")
	s.Shell().Success("Setup completed. Please review the generated configuration file before deploying.")
	s.Shell().Println("")

	return
}
