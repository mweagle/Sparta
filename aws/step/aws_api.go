package step

import (
	"fmt"
	"math/rand"
)

// DynamoDBGetItemState represents bindings for
// https://docs.aws.amazon.com/step-functions/latest/dg/connectors-ddb.html
type AWSSDKState struct {
	BaseTask
	serviceName               string
	apiAction                 string
	serviceIntegrationPattern string
	parameters                map[string]interface{}
}

// MarshalJSON for custom marshalling, since this will be stringified and we need it
// to turn into a stringified
// Ref: https://aws.amazon.com/blogs/aws/now-aws-step-functions-supports-200-aws-services-to-enable-easier-workflow-automation/
func (awsstate *AWSSDKState) MarshalJSON() ([]byte, error) {
	resourceURL := fmt.Sprintf("arn:aws:states:::aws-sdk:%s:%s",
		awsstate.serviceName,
		awsstate.apiAction)
	if awsstate.serviceIntegrationPattern != "" {
		resourceURL += fmt.Sprintf(".[%s]", awsstate.serviceIntegrationPattern)
	}
	return awsstate.BaseTask.marshalMergedParams(resourceURL, &awsstate.parameters)
}

// NewAWSSDKState returns an initialized AWSSDKState state
func NewAWSSDKState(stateName string,
	serviceName string,
	apiAction string,
	serviceIntegrationPattern string,
	parameters map[string]interface{}) *AWSSDKState {

	return NewAWSSDKIntegrationState(stateName,
		serviceName,
		apiAction,
		"",
		parameters)
}

// NewAWSSDKIntegrationState returns an initialized AWSSDKState state
func NewAWSSDKIntegrationState(stateName string,
	serviceName string,
	apiAction string,
	serviceIntegrationPattern string,
	parameters map[string]interface{}) *AWSSDKState {

	awssdk := &AWSSDKState{
		serviceName:               serviceName,
		apiAction:                 apiAction,
		serviceIntegrationPattern: serviceIntegrationPattern,
		BaseTask: BaseTask{
			baseInnerState: baseInnerState{
				name: stateName,
				id:   rand.Int63(),
			},
		},
		parameters: parameters,
	}
	return awssdk
}
