package naming

import (
	"errors"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/client"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/core"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/deployments"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/checks"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/config"
	"golang.org/x/exp/slices"
	"strings"
)

const OctoLintProjectDefaultStepNames = "OctoLintProjectDefaultStepNames"

// OctopusProjectDefaultStepNames checks to see if any project has too many steps.
type OctopusProjectDefaultStepNames struct {
	client       *client.Client
	errorHandler checks.OctopusClientErrorHandler
	config       *config.OctolintConfig
}

func NewOctopusProjectDefaultStepNames(client *client.Client, config *config.OctolintConfig, errorHandler checks.OctopusClientErrorHandler) OctopusProjectDefaultStepNames {
	return OctopusProjectDefaultStepNames{
		client:       client,
		errorHandler: errorHandler,
		config:       config,
	}
}

func (o OctopusProjectDefaultStepNames) Id() string {
	return OctoLintProjectDefaultStepNames
}

func (o OctopusProjectDefaultStepNames) Execute() (checks.OctopusCheckResult, error) {
	if o.client == nil {
		return nil, errors.New("octoclient is nil")
	}

	projects, err := o.client.Projects.GetAll()

	if err != nil {
		return o.errorHandler.HandleError(o.Id(), checks.Naming, err)
	}

	actionsWithDefaultNames := []string{}
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
				if slices.Index(checks.DefaultStepNames, a.Name) != -1 {
					actionsWithDefaultNames = append(actionsWithDefaultNames, p.Name+"/"+a.Name)
				}
			}
		}

	}

	if len(actionsWithDefaultNames) > 0 {
		return checks.NewOctopusCheckResultImpl(
			"The following project actions use the default step names:\n"+strings.Join(actionsWithDefaultNames, "\n"),
			o.Id(),
			"",
			checks.Warning,
			checks.Organization), nil
	}

	return checks.NewOctopusCheckResultImpl(
		"There are no project actions default step names",
		o.Id(),
		"",
		checks.Ok,
		checks.Organization), nil
}

func (o OctopusProjectDefaultStepNames) stepsInDeploymentProcess(deploymentProcessID string) (*deployments.DeploymentProcess, error) {
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
