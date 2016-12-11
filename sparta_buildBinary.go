// +build !lambdabinary

package sparta

// Defines functions that are only valid in the context of the build
// binary
import (
	"fmt"
)

var codePipelineEnvironments map[string]map[string]string

func init() {
	codePipelineEnvironments = make(map[string]map[string]string, 0)
}

// RegisterCodePipelineEnvironment is part of a CodePipeline deployment
// and defines the environments available for deployment
func RegisterCodePipelineEnvironment(environmentName string, environmentVariables map[string]string) error {
	if _, exists := codePipelineEnvironments[environmentName]; exists {
		return fmt.Errorf("Environment (%s) has already been defined", environmentName)
	}
	codePipelineEnvironments[environmentName] = environmentVariables
	return nil
}
