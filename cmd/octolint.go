package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/resources"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/spaces"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/checks"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/checks/factory"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/executor"
	"github.com/OctopusSolutionsEngineering/OctopusRecommendationEngine/internal/reporters"
	"github.com/OctopusSolutionsEngineering/OctopusTerraformTestFramework/octoclient"
	"github.com/briandowns/spinner"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var Version = "development"

type octolintConfig struct {
	Url           string
	Space         string
	ApiKey        string
	SkipTests     string
	VerboseErrors bool
	Version       bool
	Spinner       bool
}

func main() {
	config, err := parseArgs()

	if err != nil {
		errorExit(err.Error())
	}

	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)

	if config.Spinner {
		s.Start()
	}

	if config.Version {
		fmt.Println("Version: " + Version)
		os.Exit(0)
	}

	if config.Url == "" {
		errorExit("You must specify the URL with the -url argument")
	}

	if config.ApiKey == "" {
		errorExit("You must specify the API key with the -apiKey argument")
	}

	if !strings.HasPrefix(config.Space, "Spaces-") {
		spaceId, err := lookupSpaceAsName(config.Url, config.Space, config.ApiKey)

		if err != nil {
			errorExit("Failed to create the Octopus client")
		}

		config.Space = spaceId
	}

	client, err := octoclient.CreateClient(config.Url, config.Space, config.ApiKey)

	if err != nil {
		errorExit("Failed to create the Octopus client. Check that the url, api key, and space are correct.")
	}

	factory := factory.NewOctopusCheckFactory(client, config.Url, config.Space)
	checkCollection, err := factory.BuildAllChecks(config.SkipTests)

	if err != nil {
		errorExit("Failed to create the checks")
	}

	executor := executor.NewOctopusCheckExecutor()
	results, err := executor.ExecuteChecks(checkCollection, func(check checks.OctopusCheck, err error) error {
		fmt.Fprintf(os.Stderr, "Failed to execute check "+check.Id())
		if config.VerboseErrors {
			fmt.Println("##octopus[stdout-verbose]")
			fmt.Println(err.Error())
			fmt.Println("##octopus[stdout-default]")
		} else {
			fmt.Fprintf(os.Stderr, err.Error()+"\n")
		}
		return nil
	})

	if err != nil {
		errorExit("Failed to run the checks")
	}

	reporter := reporters.NewOctopusPlainCheckReporter(checks.Warning)
	report, err := reporter.Generate(results)

	if err != nil {
		errorExit("Failed to generate the report")
	}

	if config.Spinner {
		s.Stop()
	}

	fmt.Println(report)
}

func errorExit(message string) {
	fmt.Println(message)
	os.Exit(1)
}

func parseArgs() (*octolintConfig, error) {
	config := octolintConfig{}

	flag.StringVar(&config.Url, "url", "", "The Octopus URL e.g. https://myinstance.octopus.app")
	flag.StringVar(&config.Space, "space", "", "The Octopus space name or ID")
	flag.StringVar(&config.ApiKey, "apiKey", "", "The Octopus api key")
	flag.StringVar(&config.SkipTests, "skipTests", "", "A comma separated list of tests to skip")
	flag.BoolVar(&config.VerboseErrors, "verboseErrors", false, "Print error details as verbose logs in Octopus")
	flag.BoolVar(&config.Version, "version", false, "Print the version")
	flag.BoolVar(&config.Spinner, "spinner", true, "Display the spinner")

	flag.Parse()

	err := overrideArgs()

	if err != nil {
		return nil, err
	}

	if config.Url == "" {
		config.Url = os.Getenv("OCTOPUS_CLI_SERVER")
	}

	if config.ApiKey == "" {
		config.ApiKey = os.Getenv("OCTOPUS_CLI_API_KEY")
	}

	return &config, nil
}

// Inspired by https://github.com/carolynvs/stingoftheviper
// Viper needs manual handling to implement reading settings from env vars, config files, and from the command line
func overrideArgs() error {
	v := viper.New()

	// Set the base name of the config file, without the file extension.
	v.SetConfigName("octolint")

	// Set as many paths as you like where viper should look for the
	// config file. We are only looking in the current working directory.
	v.AddConfigPath(".")

	// Attempt to read the config file, gracefully ignoring errors
	// caused by a config file not being found. Return an error
	// if we cannot parse the config file.
	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --number
	// binds to an environment variable STING_NUMBER. This helps
	// avoid conflicts.
	v.SetEnvPrefix("octolint")

	// Environment variables can't have dashes in them, so bind them to their equivalent
	// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	return bindFlags(v)
}

// Bind each flag to its associated viper configuration (config file and environment variable)
func bindFlags(v *viper.Viper) (funErr error) {
	var funcError error = nil

	flag.VisitAll(func(allFlags *flag.Flag) {
		defined := false
		flag.Visit(func(definedFlag *flag.Flag) {
			if definedFlag.Name == allFlags.Name {
				defined = true
			}
		})

		if !defined && v.IsSet(allFlags.Name) {
			configName := strings.ReplaceAll(allFlags.Name, "-", "")
			err := flag.Set(allFlags.Name, v.GetString(configName))
			funcError = errors.Join(funcError, err)
		}
	})

	return funcError
}

func lookupSpaceAsName(octopusUrl string, spaceName string, apiKey string) (string, error) {
	if len(strings.TrimSpace(spaceName)) == 0 {
		return "", errors.New("space can not be empty")
	}

	requestURL := fmt.Sprintf("%s/api/Spaces?take=1000&partialName=%s", octopusUrl, url.QueryEscape(spaceName))

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)

	if err != nil {
		return "", err
	}

	if apiKey != "" {
		req.Header.Set("X-Octopus-ApiKey", apiKey)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", nil
	}
	defer res.Body.Close()

	collection := resources.Resources[spaces.Space]{}
	err = json.NewDecoder(res.Body).Decode(&collection)

	if err != nil {
		return "", err
	}

	for _, space := range collection.Items {
		if space.Name == spaceName {
			return space.ID, nil
		}
	}

	return "", errors.New("did not find space with name " + spaceName)
}
