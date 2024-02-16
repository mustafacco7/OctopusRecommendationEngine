package checks

import (
	"golang.org/x/exp/slices"
	"strings"
)

// SpecialVars is a list of known debug or system variables that are not expected to appear in the project itself
var SpecialVars = []string{
	"OctopusPrintVariables",
	"OctopusPrintEvaluatedVariables",
}

// IgnoreVariable is used to find special variables that the end user does not control the naming or use of
func IgnoreVariable(name string) bool {
	if slices.Index(SpecialVars[:], name) != -1 {
		return true
	}

	// Ignore variables that look like JSON substitutions
	if strings.Index(name, ":") != -1 {
		return true
	}

	// Ignore variables that look like groups
	if strings.Index(name, "[") != -1 {
		return true
	}

	return false
}
