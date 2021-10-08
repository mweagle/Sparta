package resources

import (
	"context"
	"encoding/json"
	"fmt"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2SNS "github.com/aws/aws-sdk-go-v2/service/sns"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	"github.com/rs/zerolog"
)

// SNSLambdaEventSourceResourceRequest defines the request properties to configure
// SNS
type SNSLambdaEventSourceResourceRequest struct {
	CustomResourceRequest
	LambdaTargetArn string
	SNSTopicArn     string
}

// SNSLambdaEventSourceResource is a simple POC showing how to create custom resources
type SNSLambdaEventSourceResource struct {
	gof.CustomResource
}

func (command SNSLambdaEventSourceResource) updateRegistration(isTargetActive bool,
	awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {

	request := SNSLambdaEventSourceResourceRequest{}
	unmarshalErr := json.Unmarshal(event.ResourceProperties, &request)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	// Get the current subscriptions...
	snsSvc := awsv2SNS.NewFromConfig(awsConfig)
	snsInput := &awsv2SNS.ListSubscriptionsByTopicInput{
		TopicArn: awsv2.String(request.SNSTopicArn),
	}
	listSubscriptions, listSubscriptionsErr := snsSvc.ListSubscriptionsByTopic(context.Background(), snsInput)
	if listSubscriptionsErr != nil {
		return nil, listSubscriptionsErr
	}
	var lambdaSubscriptionArn string
	for _, eachSubscription := range listSubscriptions.Subscriptions {
		if *eachSubscription.Protocol == "lambda" &&
			*eachSubscription.Endpoint == request.LambdaTargetArn {
			if lambdaSubscriptionArn != "" {
				return nil, fmt.Errorf("multiple SNS %s registrations found for lambda: %s",
					*snsInput.TopicArn,
					request.LambdaTargetArn)
			}
			lambdaSubscriptionArn = *eachSubscription.SubscriptionArn
		}
	}
	// Just log it...
	logger.Info().
		Interface("SNSTopicArn", request.SNSTopicArn).
		Interface("LambdaArn", request.LambdaTargetArn).
		Interface("ExistingSubscriptionArn", lambdaSubscriptionArn).
		Msg("Current SNS subscription status")

	var opErr error
	if isTargetActive && lambdaSubscriptionArn == "" {
		subscribeInput := &awsv2SNS.SubscribeInput{
			Protocol: awsv2.String("lambda"),
			TopicArn: awsv2.String(request.SNSTopicArn),
			Endpoint: awsv2.String(request.LambdaTargetArn),
		}
		_, opErr = snsSvc.Subscribe(context.Background(), subscribeInput)
	} else if !isTargetActive && lambdaSubscriptionArn != "" {
		unsubscribeInput := &awsv2SNS.UnsubscribeInput{
			SubscriptionArn: awsv2.String(lambdaSubscriptionArn),
		}
		_, opErr = snsSvc.Unsubscribe(context.Background(), unsubscribeInput)
	} else {
		// Just log it...
		logger.Info().
			Interface("Command", command).
			Msg("No SNS operation required")
	}

	return nil, opErr
}

// IAMPrivileges returns the IAM privs for this custom action
func (command *SNSLambdaEventSourceResource) IAMPrivileges() []string {
	return []string{"sns:ConfirmSubscription",
		"sns:GetTopicAttributes",
		"sns:ListSubscriptionsByTopic",
		"sns:Subscribe",
		"sns:Unsubscribe"}
}

// Create implements the custom resource create operation
func (command SNSLambdaEventSourceResource) Create(awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(true, awsConfig, event, logger)
}

// Update implements the custom resource update operation
func (command SNSLambdaEventSourceResource) Update(awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(true, awsConfig, event, logger)
}

// Delete implements the custom resource delete operation
func (command SNSLambdaEventSourceResource) Delete(awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(false, awsConfig, event, logger)
}
