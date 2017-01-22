package cloudformationresources

import (
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	gocf "github.com/crewjam/go-cloudformation"
)

// CloudWatchLogsLambdaEventSourceFilter represents a filter for a cloudwatchlogs
// stream
type CloudWatchLogsLambdaEventSourceFilter struct {
	Name         *gocf.StringExpr
	Pattern      *gocf.StringExpr
	LogGroupName *gocf.StringExpr
}

// CloudWatchLogsLambdaEventSourceResource is a simple POC showing how to create custom resources
type CloudWatchLogsLambdaEventSourceResource struct {
	GoAWSCustomResource
	LambdaTargetArn *gocf.StringExpr
	Filters         []*CloudWatchLogsLambdaEventSourceFilter
	RoleARN         *gocf.StringExpr `json:",omitempty"`
}

func (command CloudWatchLogsLambdaEventSourceResource) updateRegistration(isTargetActive bool,
	session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {

	var opErr error
	cwLogsSvc := cloudwatchlogs.New(session)
	for _, eachFilter := range command.Filters {

		// Always delete the filter by name if we can find it...
		deleteSubscriptionInput := &cloudwatchlogs.DeleteSubscriptionFilterInput{
			FilterName:   aws.String(eachFilter.Name.Literal),
			LogGroupName: aws.String(eachFilter.LogGroupName.Literal),
		}
		deleteResult, deleteErr := cwLogsSvc.DeleteSubscriptionFilter(deleteSubscriptionInput)
		logger.WithFields(logrus.Fields{
			"DeleteInput": deleteSubscriptionInput,
			"Result":      deleteResult,
			"Error":       deleteErr,
		}).Debug("DeleteSubscriptionFilter result")
		if nil != deleteErr && strings.Contains(deleteErr.Error(), "ResourceNotFoundException") {
			deleteErr = nil
		}
		opErr = deleteErr

		// Conditionally create
		if isTargetActive && nil == opErr {
			// Put the subscription filter
			putSubscriptionInput := &cloudwatchlogs.PutSubscriptionFilterInput{
				DestinationArn: aws.String(command.LambdaTargetArn.Literal),
				FilterName:     aws.String(eachFilter.Name.Literal),
				FilterPattern:  aws.String(eachFilter.Pattern.Literal),
				LogGroupName:   aws.String(eachFilter.LogGroupName.Literal),
			}
			if nil != command.RoleARN {
				putSubscriptionInput.RoleArn = aws.String(command.RoleARN.Literal)
			}
			_, opErr = cwLogsSvc.PutSubscriptionFilter(putSubscriptionInput)
			// If there was an error, see if there's a differently named filter for the given
			// CloudWatchLogs stream.
			if nil != opErr {
				describeSubscriptionFilters := &cloudwatchlogs.DescribeSubscriptionFiltersInput{
					LogGroupName: aws.String(eachFilter.LogGroupName.Literal),
				}
				describeResult, describeResultErr := cwLogsSvc.DescribeSubscriptionFilters(describeSubscriptionFilters)
				if nil == describeResultErr {
					opErr = fmt.Errorf("Conflict with differently named subscription on prexisting LogGroupName: %s",
						eachFilter.LogGroupName.Literal)

					logger.WithFields(logrus.Fields{
						"DescribeSubscriptionResult": describeResult,
						"PutSubscriptionInput":       putSubscriptionInput,
						"LogGroupName":               eachFilter.LogGroupName,
					}).Error(opErr.Error())
				}
			}
		}
		if nil != opErr {
			return nil, opErr
		}
	}
	return nil, opErr
}
func (command CloudWatchLogsLambdaEventSourceResource) create(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(true, session, logger)
}

func (command CloudWatchLogsLambdaEventSourceResource) update(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(true, session, logger)
}

func (command CloudWatchLogsLambdaEventSourceResource) delete(session *session.Session,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(false, session, logger)
}
