package naming

import (
	"errors"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/client"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/checks"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/config"
	"regexp"
	"strings"
)

const OctoLintInvalidTargetNames = "OctoLintInvalidTargetNames"

// OctopusInvalidTargetName find targets that have not been healthy in the last 30 days.
type OctopusInvalidTargetName struct {
	client       *client.Client
	errorHandler checks.OctopusClientErrorHandler
	config       *config.OctolintConfig
}

func NewOctopusInvalidTargetName(client *client.Client, config *config.OctolintConfig, errorHandler checks.OctopusClientErrorHandler) OctopusInvalidTargetName {
	return OctopusInvalidTargetName{
		client:       client,
		errorHandler: errorHandler,
		config:       config,
	}
}

func (o OctopusInvalidTargetName) Id() string {
	return OctoLintInvalidTargetNames
}

func (o OctopusInvalidTargetName) Execute() (checks.OctopusCheckResult, error) {
	if o.client == nil {
		return nil, errors.New("octoclient is nil")
	}

	if strings.TrimSpace(o.config.TargetNameRegex) == "" {
		return nil, nil
	}

	regex, err := regexp.Compile(o.config.TargetNameRegex)

	if err != nil {
		return checks.NewOctopusCheckResultImpl(
			"The supplied regex "+o.config.TargetNameRegex+" does not compile",
			o.Id(),
			"",
			checks.Error,
			checks.Naming), nil
	}

	allMachines, err := o.client.Machines.GetAll()

	if err != nil {
		return o.errorHandler.HandleError(o.Id(), checks.Organization, err)
	}

	responses := []string{}
	for _, m := range allMachines {
		if !regex.Match([]byte(m.Name)) {
			responses = append(responses, m.Name)
		}
	}

	if len(responses) > 0 {
		return checks.NewOctopusCheckResultImpl(
			"The following target names do not match the regex "+o.config.TargetNameRegex+":\n"+strings.Join(responses, "\n"),
			o.Id(),
			"",
			checks.Warning,
			checks.Organization), nil
	}

	return checks.NewOctopusCheckResultImpl(
		"All targets match the regex "+o.config.TargetNameRegex,
		o.Id(),
		"",
		checks.Ok,
		checks.Organization), nil
}
