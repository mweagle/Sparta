package resources

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws/session"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	"github.com/rs/zerolog"
)

// HelloWorldResourceRequest is what the UserProperties
// should be set to in the CustomResource invocation
type HelloWorldResourceRequest struct {
	CustomResourceRequest
	Message string
}

// HelloWorldResource is a simple POC showing how to create custom resources
type HelloWorldResource struct {
	gof.CustomResource
}

// IAMPrivileges returns the IAM privs for this custom action
func (command *HelloWorldResource) IAMPrivileges() []string {
	return []string{}
}

// Create implements resource create
func (command HelloWorldResource) Create(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	request := HelloWorldResourceRequest{}
	requestPropsErr := json.Unmarshal(event.ResourceProperties, &request)
	if requestPropsErr != nil {
		return nil, requestPropsErr
	}
	logger.Info().Msgf("create: Hello %s", request.Message)
	return map[string]interface{}{
		"Resource": "Created message: " + request.Message,
	}, nil
}

// Update implements resource update
func (command HelloWorldResource) Update(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	request := HelloWorldResourceRequest{}
	requestPropsErr := json.Unmarshal(event.ResourceProperties, &request)
	if requestPropsErr != nil {
		return nil, requestPropsErr
	}
	logger.Info().Msgf("update:  %s", request.Message)
	return nil, nil
}

// Delete implements resource delete
func (command HelloWorldResource) Delete(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	request := HelloWorldResourceRequest{}

	requestPropsErr := json.Unmarshal(event.ResourceProperties, &request)
	if requestPropsErr != nil {
		return nil, requestPropsErr
	}
	logger.Info().Msgf("delete: %s", request.Message)
	return nil, nil
}
