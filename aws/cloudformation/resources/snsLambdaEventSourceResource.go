package resources

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/rs/zerolog"
)

// SNSLambdaEventSourceResourceRequest defines the request properties to configure
// SNS
type SNSLambdaEventSourceResourceRequest struct {
	LambdaTargetArn *gocf.StringExpr
	SNSTopicArn     *gocf.StringExpr
}

// SNSLambdaEventSourceResource is a simple POC showing how to create custom resources
type SNSLambdaEventSourceResource struct {
	gocf.CloudFormationCustomResource
	SNSLambdaEventSourceResourceRequest
}

func (command SNSLambdaEventSourceResource) updateRegistration(isTargetActive bool,
	session *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {

	unmarshalErr := json.Unmarshal(event.ResourceProperties, &command)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	// Get the current subscriptions...
	snsSvc := sns.New(session)
	snsInput := &sns.ListSubscriptionsByTopicInput{
		TopicArn: aws.String(command.SNSTopicArn.Literal),
	}
	listSubscriptions, listSubscriptionsErr := snsSvc.ListSubscriptionsByTopic(snsInput)
	if listSubscriptionsErr != nil {
		return nil, listSubscriptionsErr
	}
	var lambdaSubscriptionArn string
	for _, eachSubscription := range listSubscriptions.Subscriptions {
		if *eachSubscription.Protocol == "lambda" &&
			*eachSubscription.Endpoint == command.LambdaTargetArn.Literal {
			if lambdaSubscriptionArn != "" {
				return nil, fmt.Errorf("multiple SNS %s registrations found for lambda: %s",
					*snsInput.TopicArn,
					command.LambdaTargetArn.Literal)
			}
			lambdaSubscriptionArn = *eachSubscription.SubscriptionArn
		}
	}
	// Just log it...
	logger.Info().
		Interface("SNSTopicArn", command.SNSTopicArn).
		Interface("LambdaArn", command.LambdaTargetArn).
		Interface("ExistingSubscriptionArn", lambdaSubscriptionArn).
		Msg("Current SNS subscription status")

	var opErr error
	if isTargetActive && lambdaSubscriptionArn == "" {
		subscribeInput := &sns.SubscribeInput{
			Protocol: aws.String("lambda"),
			TopicArn: aws.String(command.SNSTopicArn.Literal),
			Endpoint: aws.String(command.LambdaTargetArn.Literal),
		}
		_, opErr = snsSvc.Subscribe(subscribeInput)
	} else if !isTargetActive && lambdaSubscriptionArn != "" {
		unsubscribeInput := &sns.UnsubscribeInput{
			SubscriptionArn: aws.String(lambdaSubscriptionArn),
		}
		_, opErr = snsSvc.Unsubscribe(unsubscribeInput)
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
func (command SNSLambdaEventSourceResource) Create(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(true, awsSession, event, logger)
}

// Update implements the custom resource update operation
func (command SNSLambdaEventSourceResource) Update(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(true, awsSession, event, logger)
}

// Delete implements the custom resource delete operation
func (command SNSLambdaEventSourceResource) Delete(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(false, awsSession, event, logger)
}
