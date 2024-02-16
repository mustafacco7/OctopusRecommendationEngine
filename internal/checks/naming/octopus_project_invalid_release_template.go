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

const OctoLintProjectReleaseTemplate = "OctoLintProjectReleaseTemplate"

// OctopusProjectReleaseTemplateRegex checks to see if any project has too many steps.
type OctopusProjectReleaseTemplateRegex struct {
	client       *client.Client
	errorHandler checks.OctopusClientErrorHandler
	config       *config.OctolintConfig
}

func NewOctopusProjectReleaseTemplateRegex(client *client.Client, config *config.OctolintConfig, errorHandler checks.OctopusClientErrorHandler) OctopusProjectReleaseTemplateRegex {
	return OctopusProjectReleaseTemplateRegex{
		client:       client,
		errorHandler: errorHandler,
		config:       config,
	}
}

func (o OctopusProjectReleaseTemplateRegex) Id() string {
	return OctoLintProjectReleaseTemplate
}

func (o OctopusProjectReleaseTemplateRegex) Execute() (checks.OctopusCheckResult, error) {
	if o.client == nil {
		return nil, errors.New("octoclient is nil")
	}

	if strings.TrimSpace(o.config.ReleaseTemplateRegex) == "" {
		return nil, nil
	}

	regex, err := regexp.Compile(o.config.ReleaseTemplateRegex)

	if err != nil {

		return checks.NewOctopusCheckResultImpl(
			"The supplied regex "+o.config.ReleaseTemplateRegex+" does not compile",
			o.Id(),
			"",
			checks.Error,
			checks.Naming), nil
	}

	projects, err := o.client.Projects.GetAll()

	if err != nil {
		return o.errorHandler.HandleError(o.Id(), checks.Organization, err)
	}

	results := []string{}
	for _, p := range projects {
		if p.VersioningStrategy != nil && !regex.Match([]byte(p.VersioningStrategy.Template)) {
			results = append(results, p.Name+" - "+p.VersioningStrategy.Template)
		}
	}

	if len(results) > 0 {
		return checks.NewOctopusCheckResultImpl(
			"The following project release templates do not match the regex "+o.config.ReleaseTemplateRegex+":\n"+strings.Join(results, "\n"),
			o.Id(),
			"",
			checks.Warning,
			checks.Organization), nil
	}

	return checks.NewOctopusCheckResultImpl(
		"All projects match the release templates regex "+o.config.ReleaseTemplateRegex,
		o.Id(),
		"",
		checks.Ok,
		checks.Organization), nil
}

func (o OctopusProjectReleaseTemplateRegex) stepsInDeploymentProcess(deploymentProcessID string) (*deployments.DeploymentProcess, error) {
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
