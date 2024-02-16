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

	// These values are used to configure individual checks
	MaxEnvironments      int
	ContainerImageRegex  string
	VariableNameRegex    string
	TargetNameRegex      string
	TargetRoleRegex      string
	ReleaseTemplateRegex string
}
