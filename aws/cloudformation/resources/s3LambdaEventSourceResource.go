package resources

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	"github.com/rs/zerolog"
)

// S3LambdaEventSourceResourceRequest is what the UserProperties
// should be set to in the CustomResource invocation
type S3LambdaEventSourceResourceRequest struct {
	CustomResourceRequest
	BucketArn       string
	Events          []string
	LambdaTargetArn string
	Filter          *s3.NotificationConfigurationFilter `json:"Filter,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// TODO - update all the custom resources to use this approach so that
// the properties object is propertly serialized. We'll also need to deserialize
// the request for the custom handler.
////////////////////////////////////////////////////////////////////////////////

// S3LambdaEventSourceResource manages registering a Lambda function with S3 event
type S3LambdaEventSourceResource struct {
	gof.CustomResource
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
	logger *zerolog.Logger) (map[string]interface{}, error) {

	s3EventRequest := S3LambdaEventSourceResourceRequest{}
	unmarshalErr := json.Unmarshal(event.ResourceProperties, &s3EventRequest)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	s3Svc := s3.New(session)
	bucketParts := strings.Split(s3EventRequest.BucketArn, ":")
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
		if *eachLambdaConfig.LambdaFunctionArn != s3EventRequest.LambdaTargetArn {
			lambdaConfigurations = append(lambdaConfigurations, eachLambdaConfig)
		}
	}

	if isTargetActive {
		var eventPtrs []*string
		for _, eachString := range s3EventRequest.Events {
			eventPtrs = append(eventPtrs, aws.String(eachString))
		}
		commandConfig := &s3.LambdaFunctionConfiguration{
			LambdaFunctionArn: aws.String(s3EventRequest.LambdaTargetArn),
			Events:            eventPtrs,
		}
		if s3EventRequest.Filter != nil {
			commandConfig.Filter = s3EventRequest.Filter
		}
		lambdaConfigurations = append(lambdaConfigurations, commandConfig)
	}
	config.LambdaFunctionConfigurations = lambdaConfigurations

	putBucketNotificationConfigurationInput := &s3.PutBucketNotificationConfigurationInput{
		Bucket:                    aws.String(bucketName),
		NotificationConfiguration: config,
	}

	logger.Debug().
		Interface("PutBucketNotificationConfigurationInput", putBucketNotificationConfigurationInput).
		Msg("Updating bucket configuration")

	_, putErr := s3Svc.PutBucketNotificationConfiguration(putBucketNotificationConfigurationInput)
	return nil, putErr
}

// Create implements the custom resource create operation
func (command S3LambdaEventSourceResource) Create(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateNotification(true, awsSession, event, logger)
}

// Update implements the custom resource update operation
func (command S3LambdaEventSourceResource) Update(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateNotification(true, awsSession, event, logger)
}

// Delete implements the custom resource delete operation
func (command S3LambdaEventSourceResource) Delete(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateNotification(false, awsSession, event, logger)
}
