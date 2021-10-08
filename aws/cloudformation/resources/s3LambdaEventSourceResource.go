package resources

import (
	"context"
	"encoding/json"
	"strings"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2S3 "github.com/aws/aws-sdk-go-v2/service/s3"
	awsv2S3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"

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
	Filter          *awsv2S3Types.NotificationConfigurationFilter `json:"Filter,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// TODO - update all the custom resources to use this approach so that
// the properties object is properly serialized. We'll also need to deserialize
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
	awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {

	s3EventRequest := S3LambdaEventSourceResourceRequest{}
	unmarshalErr := json.Unmarshal(event.ResourceProperties, &s3EventRequest)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	s3Svc := awsv2S3.NewFromConfig(awsConfig)
	bucketParts := strings.Split(s3EventRequest.BucketArn, ":")
	bucketName := bucketParts[len(bucketParts)-1]

	params := &awsv2S3.GetBucketNotificationConfigurationInput{
		Bucket: awsv2.String(bucketName),
	}
	config, configErr := s3Svc.GetBucketNotificationConfiguration(context.Background(), params)
	if nil != configErr {
		return nil, configErr
	}
	// First thing, eliminate existing references...
	var lambdaConfigurations []awsv2S3Types.LambdaFunctionConfiguration
	for _, eachLambdaConfig := range config.LambdaFunctionConfigurations {
		if *eachLambdaConfig.LambdaFunctionArn != s3EventRequest.LambdaTargetArn {
			lambdaConfigurations = append(lambdaConfigurations, eachLambdaConfig)
		}
	}

	if isTargetActive {
		var eventPtrs []awsv2S3Types.Event
		for _, eachString := range s3EventRequest.Events {
			eventPtrs = append(eventPtrs, awsv2S3Types.Event(eachString))
		}
		commandConfig := awsv2S3Types.LambdaFunctionConfiguration{
			LambdaFunctionArn: awsv2.String(s3EventRequest.LambdaTargetArn),
			Events:            eventPtrs,
		}
		if s3EventRequest.Filter != nil {
			commandConfig.Filter = s3EventRequest.Filter
		}
		lambdaConfigurations = append(lambdaConfigurations, commandConfig)
	}

	putBucketNotificationConfigurationInput := &awsv2S3.PutBucketNotificationConfigurationInput{
		Bucket: awsv2.String(bucketName),
		NotificationConfiguration: &awsv2S3Types.NotificationConfiguration{
			LambdaFunctionConfigurations: lambdaConfigurations,
		},
	}

	logger.Debug().
		Interface("PutBucketNotificationConfigurationInput", putBucketNotificationConfigurationInput).
		Msg("Updating bucket configuration")

	_, putErr := s3Svc.PutBucketNotificationConfiguration(context.Background(), putBucketNotificationConfigurationInput)
	return nil, putErr
}

// Create implements the custom resource create operation
func (command S3LambdaEventSourceResource) Create(awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateNotification(true, awsConfig, event, logger)
}

// Update implements the custom resource update operation
func (command S3LambdaEventSourceResource) Update(awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateNotification(true, awsConfig, event, logger)
}

// Delete implements the custom resource delete operation
func (command S3LambdaEventSourceResource) Delete(awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateNotification(false, awsConfig, event, logger)
}
