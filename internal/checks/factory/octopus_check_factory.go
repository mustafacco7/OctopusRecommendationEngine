package factory

import (
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/client"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/checks"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/checks/naming"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/checks/organization"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/checks/performance"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/checks/security"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/config"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
	"strings"
)

// OctopusCheckFactory builds all the lint checks. This is where you can customize things like error handlers.
type OctopusCheckFactory struct {
	client       *client.Client
	errorHandler checks.OctopusClientErrorHandler
	url          string
	space        string
}

func NewOctopusCheckFactory(client *client.Client, url string, space string) OctopusCheckFactory {
	return OctopusCheckFactory{client: client, url: url, space: space, errorHandler: checks.OctopusClientPermissiveErrorHandler{}}
}

// BuildAllChecks creates new instances of all the checks and returns them as an array.
func (o OctopusCheckFactory) BuildAllChecks(config *config.OctolintConfig) ([]checks.OctopusCheck, error) {
	skipChecksSlice := lo.Map(strings.Split(config.SkipTests, ","), func(item string, index int) string {
		return strings.TrimSpace(item)
	})

	allChecks := []checks.OctopusCheck{
		security.NewOctopusUnrotatedAccountsCheck(o.client, o.errorHandler),
		security.NewOctopusDeploymentQueuedByAdminCheck(o.client, o.errorHandler),
		security.NewOctopusPerpetualApiKeysCheck(o.client, o.errorHandler),
		security.NewOctopusDuplicatedGitCredentialsCheck(o.client, o.errorHandler),
		security.NewOctopusInsecureK8sCheck(o.client, o.errorHandler),
		security.NewOctopusInsecureFeedsCheck(o.client, o.errorHandler),
		security.NewOctopusInsecureSubscriptionsCheck(o.client, o.errorHandler),
		organization.NewOctopusEnvironmentCountCheck(o.client, config, o.errorHandler),
		organization.NewOctopusDefaultProjectGroupCountCheck(o.client, o.errorHandler),
		organization.NewOctopusEmptyProjectCheck(o.client, o.errorHandler),
		organization.NewOctopusUnusedVariablesCheck(o.client, o.errorHandler),
		organization.NewOctopusDuplicatedVariablesCheck(o.client, o.errorHandler),
		organization.NewOctopusProjectTooManyStepsCheck(o.client, o.errorHandler),
		organization.NewOctopusLifecycleRetentionPolicyCheck(o.client, o.errorHandler),
		organization.NewOctopusUnusedTargetsCheck(o.client, o.errorHandler),
		organization.NewOctopusProjectSpecificEnvironmentCheck(o.client, o.errorHandler),
		organization.NewOctopusTenantsInsteadOfTagsCheck(o.client, o.errorHandler),
		organization.NewOctopusProjectGroupsWithExclusiveEnvironmentsCheck(o.client, o.errorHandler),
		organization.NewOctopusUnhealthyTargetCheck(o.client, o.errorHandler),
		performance.NewOctopusDeploymentQueuedTimeCheck(o.client, o.url, o.space, o.errorHandler),
		naming.NewOctopusProjectContainerImageRegex(o.client, config, o.errorHandler),
		naming.NewOctopusInvalidVariableNameCheck(o.client, config, o.errorHandler),
		naming.NewOctopusInvalidTargetName(o.client, config, o.errorHandler),
		naming.NewOctopusInvalidTargetRole(o.client, config, o.errorHandler),
		naming.NewOctopusProjectReleaseTemplateRegex(o.client, config, o.errorHandler),
		naming.NewOctopusProjectWorkerPoolRegex(o.client, config, o.errorHandler),
		naming.NewOctopusInvalidLifecycleName(o.client, config, o.errorHandler),
		naming.NewOctopusProjectDefaultStepNames(o.client, config, o.errorHandler),
	}

	return lo.Filter(allChecks, func(item checks.OctopusCheck, index int) bool {
		return slices.Index(skipChecksSlice, item.Id()) == -1
	}), nil
}
