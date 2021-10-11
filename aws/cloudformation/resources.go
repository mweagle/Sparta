package cloudformation

import (
	"encoding/json"
	"fmt"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	"github.com/jmespath/go-jmespath"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

//go:embed cloudformation-schema.json
var cloudformationSchema string

// ResourceOutputs is responsible for returning the conditional
// set of CloudFormation outputs for a given resource type. These are
// produced from the schema
func ResourceOutputs(resourceName string,
	resource gof.Resource,
	logger *zerolog.Logger) ([]string, error) {

	var rawData interface{}
	unmarshalErr := json.Unmarshal([]byte(cloudformationSchema), &rawData)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	// Issue the JMES query to find this resource in the schema...
	jmesQuery := fmt.Sprintf("ResourceTypes.\"%s\".Attributes", resource.AWSCloudFormationType())
	result, resultErr := jmespath.Search(jmesQuery, rawData)
	if resultErr != nil {
		return nil, resultErr
	}

	resultMap, resultMapOk := result.(map[string]interface{})
	if !resultMapOk {
		// If this a custom resource, there are no outputs...
		if resource.AWSCloudFormationType() == "AWS::CloudFormation::CustomResource" {
			return nil, nil
		}
		return nil, errors.Errorf("Failed to extract outputs for resource type: %s", resource.AWSCloudFormationType())
	}

	vals := []string{}
	for eachKey := range resultMap {
		vals = append(vals, eachKey)
	}
	return vals, nil
}
