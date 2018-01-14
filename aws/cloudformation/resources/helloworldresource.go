package resources

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws/session"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/sirupsen/logrus"
)

// HelloWorldResourceRequest is what the UserProperties
// should be set to in the CustomResource invocation
type HelloWorldResourceRequest struct {
	Message *gocf.StringExpr
}

// HelloWorldResource is a simple POC showing how to create custom resources
type HelloWorldResource struct {
	gocf.CloudFormationCustomResource
	HelloWorldResourceRequest
}

func (command HelloWorldResource) create(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *logrus.Logger) (map[string]interface{}, error) {

	requestPropsErr := json.Unmarshal(event.ResourceProperties, &command)
	if requestPropsErr != nil {
		return nil, requestPropsErr
	}
	logger.Info("create: Hello ", command.Message.Literal)
	return map[string]interface{}{
		"Resource": "Created message: " + command.Message.Literal,
	}, nil
}

func (command HelloWorldResource) update(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *logrus.Logger) (map[string]interface{}, error) {
	requestPropsErr := json.Unmarshal(event.ResourceProperties, &command)
	if requestPropsErr != nil {
		return nil, requestPropsErr
	}

	if requestPropsErr != nil {
		return nil, requestPropsErr
	}

	logger.Info("update: ", command.Message.Literal)
	return nil, nil
}

func (command HelloWorldResource) delete(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *logrus.Logger) (map[string]interface{}, error) {
	requestPropsErr := json.Unmarshal(event.ResourceProperties, &command)
	if requestPropsErr != nil {
		return nil, requestPropsErr
	}
	if requestPropsErr != nil {
		return nil, requestPropsErr
	}
	logger.Info("delete: ", command.Message.Literal)
	return nil, nil
}
