package config

type OctolintConfig struct {
	Url           string
	Space         string
	ApiKey        string
	SkipTests     string
	VerboseErrors bool
	Version       bool
	Spinner       bool

	// These values are used to configure individual checks
	MaxEnvironments int

	// Container image regex
	ContainerImageRegex string
}
