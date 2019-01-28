package resources

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/sirupsen/logrus"
)

// S3LambdaEventSourceResourceRequest is what the UserProperties
// should be set to in the CustomResource invocation
type S3LambdaEventSourceResourceRequest struct {
	BucketArn       *gocf.StringExpr
	Events          []string
	LambdaTargetArn *gocf.StringExpr
	Filter          *s3.NotificationConfigurationFilter `json:"Filter,omitempty"`
}

// S3LambdaEventSourceResource manages registering a Lambda function with S3 event
type S3LambdaEventSourceResource struct {
	gocf.CloudFormationCustomResource
	S3LambdaEventSourceResourceRequest
}

// IAMPrivileges returns the IAM privs for this custom action
func (command *S3LambdaEventSourceResource) IAMPrivileges() []string {
	return []string{"s3:GetBucketLocation",
		"s3:GetBucketNotification",
		"s3:PutBucketNotification",
		"s3:GetBucketNotificationConfiguration",
		"s3:PutBucketNotificationConfiguration"}
}

func (command S3LambdaEventSourceResource) updateNotification(isTargetActive bool,
	session *session.Session,
	event *CloudFormationLambdaEvent,
	logger *logrus.Logger) (map[string]interface{}, error) {

	unmarshalErr := json.Unmarshal(event.ResourceProperties, &command)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

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
		if command.Filter != nil {
			commandConfig.Filter = command.Filter
		}
		lambdaConfigurations = append(lambdaConfigurations, commandConfig)
	}
	config.LambdaFunctionConfigurations = lambdaConfigurations

	putBucketNotificationConfigurationInput := &s3.PutBucketNotificationConfigurationInput{
		Bucket:                    aws.String(bucketName),
		NotificationConfiguration: config,
	}

	logger.WithFields(logrus.Fields{
		"PutBucketNotificationConfigurationInput": putBucketNotificationConfigurationInput,
	}).Debug("Updating bucket configuration")

	_, putErr := s3Svc.PutBucketNotificationConfiguration(putBucketNotificationConfigurationInput)
	return nil, putErr
}

// Create implements the custom resource create operation
func (command S3LambdaEventSourceResource) Create(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.updateNotification(true, awsSession, event, logger)
}

// Update implements the custom resource update operation
func (command S3LambdaEventSourceResource) Update(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.updateNotification(true, awsSession, event, logger)
}

// Delete implements the custom resource delete operation
func (command S3LambdaEventSourceResource) Delete(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.updateNotification(false, awsSession, event, logger)
}
