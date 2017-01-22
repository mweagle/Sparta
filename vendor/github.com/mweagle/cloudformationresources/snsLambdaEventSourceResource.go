package cloudformationresources

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"

	gocf "github.com/crewjam/go-cloudformation"
)

// SNSLambdaEventSourceResource is a simple POC showing how to create custom resources
type SNSLambdaEventSourceResource struct {
	GoAWSCustomResource
	LambdaTargetArn *gocf.StringExpr
	SNSTopicArn     *gocf.StringExpr
}

func (command SNSLambdaEventSourceResource) updateRegistration(isTargetActive bool,
	session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {

	// Get the current subscriptions...
	snsSvc := sns.New(session)
	snsInput := &sns.ListSubscriptionsByTopicInput{
		TopicArn: aws.String(command.SNSTopicArn.Literal),
	}
	listSubscriptions, listSubscriptionsErr := snsSvc.ListSubscriptionsByTopic(snsInput)
	if nil != listSubscriptionsErr {
		return nil, listSubscriptionsErr
	}
	var lambdaSubscriptionArn string
	for _, eachSubscription := range listSubscriptions.Subscriptions {
		if *eachSubscription.Protocol == "lambda" && *eachSubscription.Endpoint == command.LambdaTargetArn.Literal {
			if "" != lambdaSubscriptionArn {
				return nil, fmt.Errorf("Multiple SNS %s registrations found for lambda: %s",
					*snsInput.TopicArn,
					command.LambdaTargetArn.Literal)
			}
			lambdaSubscriptionArn = *eachSubscription.SubscriptionArn
		}
	}
	// Just log it...
	logger.WithFields(logrus.Fields{
		"SNSTopicArn":             command.SNSTopicArn,
		"LambdaArn":               command.LambdaTargetArn,
		"ExistingSubscriptionArn": lambdaSubscriptionArn,
	}).Info("Current SNS subscription status")

	var opErr error
	if isTargetActive && "" == lambdaSubscriptionArn {
		subscribeInput := &sns.SubscribeInput{
			Protocol: aws.String("lambda"),
			TopicArn: aws.String(command.SNSTopicArn.Literal),
			Endpoint: aws.String(command.LambdaTargetArn.Literal),
		}
		_, opErr = snsSvc.Subscribe(subscribeInput)
	} else if !isTargetActive && "" != lambdaSubscriptionArn {
		unsubscribeInput := &sns.UnsubscribeInput{
			SubscriptionArn: aws.String(lambdaSubscriptionArn),
		}
		_, opErr = snsSvc.Unsubscribe(unsubscribeInput)
	} else {
		// Just log it...
		logger.WithFields(logrus.Fields{
			"Command": command,
		}).Info("No SNS operation required")
	}

	return nil, opErr
}
func (command SNSLambdaEventSourceResource) create(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(true, session, logger)
}

func (command SNSLambdaEventSourceResource) update(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(true, session, logger)
}

func (command SNSLambdaEventSourceResource) delete(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(false, session, logger)
}
