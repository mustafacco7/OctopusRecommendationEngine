package client_wrapper

import (
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/newclient"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/projects"
)

func GetProjects(limit int, client newclient.Client, spaceID string) ([]*projects.Project, error) {
	if limit == 0 {
		return projects.GetAll(client, spaceID)
	}

	result, err := projects.Get(client, spaceID, projects.ProjectsQuery{
		Take: limit,
	})

	if err != nil {
		return nil, err
	}

	return result.Items, nil
}
