package naming

import (
	"errors"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/client"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/core"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/deployments"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/checks"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/config"
	"regexp"
	"strings"
)

var OctoLintContainerImageName = "OctoLintContainerImageName"

// OctopusProjectContainerImageRegex checks to see if any project has too many steps.
type OctopusProjectContainerImageRegex struct {
	client       *client.Client
	errorHandler checks.OctopusClientErrorHandler
	config       *config.OctolintConfig
}

func NewOctopusProjectContainerImageRegex(client *client.Client, config *config.OctolintConfig, errorHandler checks.OctopusClientErrorHandler) OctopusProjectContainerImageRegex {
	return OctopusProjectContainerImageRegex{
		client:       client,
		errorHandler: errorHandler,
		config:       config,
	}
}

func (o OctopusProjectContainerImageRegex) Id() string {
	return "OctoLintContainerImageName"
}

func (o OctopusProjectContainerImageRegex) Execute() (checks.OctopusCheckResult, error) {
	if o.client == nil {
		return nil, errors.New("octoclient is nil")
	}

	if strings.TrimSpace(o.config.ContainerImageRegex) == "" {
		return nil, nil
	}

	regex, err := regexp.Compile(o.config.ContainerImageRegex)

	if err != nil {

		return checks.NewOctopusCheckResultImpl(
			"The supplied regex "+o.config.ContainerImageRegex+" does not compile",
			o.Id(),
			"",
			checks.Error,
			checks.Naming), nil
	}

	projects, err := o.client.Projects.GetAll()

	if err != nil {
		return o.errorHandler.HandleError(o.Id(), checks.Organization, err)
	}

	actionsWithInvalidImages := []string{}
	for _, p := range projects {
		deploymentProcess, err := o.stepsInDeploymentProcess(p.DeploymentProcessID)

		if err != nil {
			if !o.errorHandler.ShouldContinue(err) {
				return nil, err
			}
			continue
		}

		if deploymentProcess == nil {
			continue
		}

		for _, s := range deploymentProcess.Steps {
			for _, a := range s.Actions {
				if a.Container == nil || strings.TrimSpace(a.Container.Image) == "" {
					continue
				}

				if !regex.Match([]byte(a.Container.Image)) {
					actionsWithInvalidImages = append(actionsWithInvalidImages, p.Name+"/"+a.Name+": "+a.Container.Image)
				}
			}
		}

	}

	if len(actionsWithInvalidImages) > 0 {
		return checks.NewOctopusCheckResultImpl(
			"The following project actions have invalid container images:\n"+strings.Join(actionsWithInvalidImages, "\n"),
			o.Id(),
			"",
			checks.Warning,
			checks.Organization), nil
	}

	return checks.NewOctopusCheckResultImpl(
		"There are no project actions with invalid container images",
		o.Id(),
		"",
		checks.Ok,
		checks.Organization), nil
}

func (o OctopusProjectContainerImageRegex) stepsInDeploymentProcess(deploymentProcessID string) (*deployments.DeploymentProcess, error) {
	if deploymentProcessID == "" {
		return nil, nil
	}

	resource, err := o.client.DeploymentProcesses.GetByID(deploymentProcessID)

	if err != nil {
		// If we can't find the deployment process, assume zero steps
		if err.(*core.APIError).StatusCode == 404 {
			return nil, nil
		}
		return nil, err
	}

	return resource, nil
}
