package config

type OctolintConfig struct {
	Url           string
	Space         string
	ApiKey        string
	SkipTests     string
	VerboseErrors bool
	Version       bool
	Spinner       bool
	ConfigFile    string
	ConfigPath    string
	Verbose       bool

	// These values are used to configure individual checks
	MaxEnvironments             int
	ContainerImageRegex         string
	VariableNameRegex           string
	TargetNameRegex             string
	WorkerNameRegex             string
	WorkerPoolNameRegex         string
	TargetRoleRegex             string
	ProjectReleaseTemplateRegex string
	ProjectStepWorkerPoolRegex  string
	SpaceNameRegex              string
	LibraryVariableSetNameRegex string
	TenantNameRegex             string
	TagSetNameRegex             string
	TagNameRegex                string
	FeedNameRegex               string
	AccountNameRegex            string
	MachinePolicyNameRegex      string
	CertificateNameRegex        string
	GitCredentialNameRegex      string
	ScriptModuleNameRegex       string
	ProjectGroupNameRegex       string
	ProjectNameRegex            string
	LifecycleNameRegex          string
}
