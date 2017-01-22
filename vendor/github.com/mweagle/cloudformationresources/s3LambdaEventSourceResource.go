package cloudformationresources

import (
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	gocf "github.com/crewjam/go-cloudformation"
)

// S3LambdaEventSourceResource manages registering a Lambda function with S3 event
type S3LambdaEventSourceResource struct {
	GoAWSCustomResource
	BucketArn       *gocf.StringExpr
	Events          []string
	LambdaTargetArn *gocf.StringExpr
	Filter          *s3.NotificationConfigurationFilter `json:"Filter,omitempty"`
}

func (command S3LambdaEventSourceResource) updateNotification(isTargetActive bool,
	session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {

	s3Svc := s3.New(session)
	bucketParts := strings.Split(command.BucketArn.Literal, ":")
	bucketName := bucketParts[len(bucketParts)-1]

	params := &s3.GetBucketNotificationConfigurationRequest{
		Bucket: aws.String(bucketName),
	}
	config, configErr := s3Svc.GetBucketNotificationConfiguration(params)
	if nil != configErr {
		return nil, configErr
	}
	// First thing, eliminate existing references...
	var lambdaConfigurations []*s3.LambdaFunctionConfiguration
	for _, eachLambdaConfig := range config.LambdaFunctionConfigurations {
		if *eachLambdaConfig.LambdaFunctionArn != command.LambdaTargetArn.Literal {
			lambdaConfigurations = append(lambdaConfigurations, eachLambdaConfig)
		}
	}

	if isTargetActive {
		var eventPtrs []*string
		for _, eachString := range command.Events {
			eventPtrs = append(eventPtrs, aws.String(eachString))
		}
		commandConfig := &s3.LambdaFunctionConfiguration{
			LambdaFunctionArn: aws.String(command.LambdaTargetArn.Literal),
			Events:            eventPtrs,
		}
		if nil != command.Filter {
			commandConfig.Filter = command.Filter
		}
		lambdaConfigurations = append(lambdaConfigurations, commandConfig)
	}
	config.LambdaFunctionConfigurations = lambdaConfigurations

	putBucketNotificationConfigurationInput := &s3.PutBucketNotificationConfigurationInput{
		Bucket: aws.String(bucketName),
		NotificationConfiguration: config,
	}

	logger.WithFields(logrus.Fields{
		"PutBucketNotificationConfigurationInput": putBucketNotificationConfigurationInput,
	}).Debug("Updating bucket configuration")

	_, putErr := s3Svc.PutBucketNotificationConfiguration(putBucketNotificationConfigurationInput)
	return nil, putErr
}

func (command S3LambdaEventSourceResource) create(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.updateNotification(true, session, logger)
}

func (command S3LambdaEventSourceResource) update(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.updateNotification(true, session, logger)
}

func (command S3LambdaEventSourceResource) delete(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.updateNotification(false, session, logger)
}
