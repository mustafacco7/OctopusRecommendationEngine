package organization

import (
	"errors"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/client"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/projects"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/tasks"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/checks"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/config"
	"go.uber.org/zap"
	"strings"
	"time"
)

const OctopusUnusedProjectsCheckName = "OctoLintUnusedProjects"

// OctopusUnusedProjectsCheck find projects that have not had a deployment in the last 30 days
type OctopusUnusedProjectsCheck struct {
	client       *client.Client
	errorHandler checks.OctopusClientErrorHandler
	config       *config.OctolintConfig
}

func NewOctopusUnusedProjectsCheck(client *client.Client, config *config.OctolintConfig, errorHandler checks.OctopusClientErrorHandler) OctopusUnusedProjectsCheck {
	return OctopusUnusedProjectsCheck{config: config, client: client, errorHandler: errorHandler}
}

func (o OctopusUnusedProjectsCheck) Id() string {
	return OctopusUnusedProjectsCheckName
}

func (o OctopusUnusedProjectsCheck) Execute() (checks.OctopusCheckResult, error) {
	if o.client == nil {
		return nil, errors.New("octoclient is nil")
	}

	if o.config.Verbose {
		zap.L().Info("Starting check " + o.Id())
	}

	defer func() {
		if o.config.Verbose {
			zap.L().Info("Ended check " + o.Id())
		}
	}()

	projects, err := projects.GetAll(o.client, o.client.GetSpaceID())

	if err != nil {
		return o.errorHandler.HandleError(o.Id(), checks.Organization, err)
	}

	unusedProjects := []string{}
	for _, project := range projects {

		projectHasTask := false

		tasks, err := o.client.Tasks.Get(tasks.TasksQuery{
			Environment:             "",
			HasPendingInterruptions: false,
			HasWarningsOrErrors:     false,
			IncludeSystem:           true,
			IsActive:                false,
			IsRunning:               false,
			Project:                 project.ID,
			Skip:                    0,
			Take:                    100,
		})

		if err != nil {
			return o.errorHandler.HandleError(o.Id(), checks.Organization, err)
		}

		for _, task := range tasks.Items {
			if task.StartTime != nil && task.StartTime.After(time.Now().Add(-time.Hour*24*time.Duration(o.config.MaxDaysSinceLastTask))) {
				projectHasTask = true
				break
			}
		}

		if !projectHasTask {
			unusedProjects = append(unusedProjects, project.Name)
		}
	}

	if len(unusedProjects) > 0 {
		return checks.NewOctopusCheckResultImpl(
			"The following projects have not had any tasks 30 days:\n"+strings.Join(unusedProjects, "\n"),
			o.Id(),
			"",
			checks.Warning,
			checks.Organization), nil
	}

	return checks.NewOctopusCheckResultImpl(
		"There are no projects that have not had any tasks in the last 30 days",
		o.Id(),
		"",
		checks.Ok,
		checks.Organization), nil
}
