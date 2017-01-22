package cloudformationresources

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/session"
)

// HelloWorldResource is a simple POC showing how to create custom resources
type HelloWorldResource struct {
	GoAWSCustomResource
	Message string
}

func (command HelloWorldResource) create(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	logger.Info("create: Hello ", command.Message)
	return map[string]interface{}{
		"Resource": "Created message: " + command.Message,
	}, nil
}

func (command HelloWorldResource) update(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	logger.Info("update: ", command.Message)
	return nil, nil
}

func (command HelloWorldResource) delete(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	logger.Info("delete: ", command.Message)
	return nil, nil
}
