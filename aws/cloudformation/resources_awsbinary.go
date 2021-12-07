//go:build lambdabinary

package cloudformation

import (
	gof "github.com/awslabs/goformation/v5/cloudformation"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// ResourceOutputs is responsible for returning the
// set of CloudFormation outputs for a given resource type. These are
// produced from the schema that has been previously downloaded and
// embedded into the binary.
func ResourceOutputs(resourceName string,
	resource gof.Resource,
	logger *zerolog.Logger) ([]string, error) {

	return nil, errors.New("CloudFormation ResourceOutputs not supported outside of AWS Lambda environment")
}
