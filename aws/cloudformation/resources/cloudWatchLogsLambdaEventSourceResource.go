package resources

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/rs/zerolog"
)

// CloudWatchLogsLambdaEventSourceFilter represents a filter for a cloudwatchlogs
// stream
type CloudWatchLogsLambdaEventSourceFilter struct {
	Name         *gocf.StringExpr
	Pattern      *gocf.StringExpr
	LogGroupName *gocf.StringExpr
}

// CloudWatchEventSourceResourceRequest is what the UserProperties
// should be set to in the CustomResource invocation
type CloudWatchEventSourceResourceRequest struct {
	LambdaTargetArn *gocf.StringExpr
	Filters         []*CloudWatchLogsLambdaEventSourceFilter
	RoleARN         *gocf.StringExpr `json:",omitempty"`
}

// CloudWatchLogsLambdaEventSourceResource is a simple POC showing how to create custom resources
type CloudWatchLogsLambdaEventSourceResource struct {
	gocf.CloudFormationCustomResource
	CloudWatchEventSourceResourceRequest
}

// IAMPrivileges returns the IAM privs for this custom action
func (command *CloudWatchLogsLambdaEventSourceResource) IAMPrivileges() []string {
	return []string{"logs:DescribeSubscriptionFilters",
		"logs:DeleteSubscriptionFilter",
		"logs:PutSubscriptionFilter"}
}

func cloudWatchEventSourceProperties(event *CloudFormationLambdaEvent) (*CloudWatchEventSourceResourceRequest, error) {
	eventProperties := CloudWatchEventSourceResourceRequest{}
	unmarshalErr := json.Unmarshal(event.ResourceProperties, &eventProperties)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	return &eventProperties, nil
}

func (command CloudWatchLogsLambdaEventSourceResource) updateRegistration(isTargetActive bool,
	session *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {

	requestProps, requestPropsErr := cloudWatchEventSourceProperties(event)
	if requestPropsErr != nil {
		return nil, requestPropsErr
	}

	var opErr error
	cwLogsSvc := cloudwatchlogs.New(session)
	for _, eachFilter := range requestProps.Filters {

		// Always delete the filter by name if we can find it...
		deleteSubscriptionInput := &cloudwatchlogs.DeleteSubscriptionFilterInput{
			FilterName:   aws.String(eachFilter.Name.Literal),
			LogGroupName: aws.String(eachFilter.LogGroupName.Literal),
		}
		deleteResult, deleteErr := cwLogsSvc.DeleteSubscriptionFilter(deleteSubscriptionInput)
		logger.Debug().
			Interface("DeleteInput", deleteSubscriptionInput).
			Interface("Result", deleteResult).
			Interface("Error", deleteErr).
			Msg("DeleteSubscriptionFilter result")

		if nil != deleteErr && strings.Contains(deleteErr.Error(), "ResourceNotFoundException") {
			deleteErr = nil
		}
		opErr = deleteErr

		// Conditionally create
		if isTargetActive && nil == opErr {
			// Put the subscription filter
			putSubscriptionInput := &cloudwatchlogs.PutSubscriptionFilterInput{
				DestinationArn: aws.String(requestProps.LambdaTargetArn.Literal),
				FilterName:     aws.String(eachFilter.Name.Literal),
				FilterPattern:  aws.String(eachFilter.Pattern.Literal),
				LogGroupName:   aws.String(eachFilter.LogGroupName.Literal),
			}
			if nil != requestProps.RoleARN {
				putSubscriptionInput.RoleArn = aws.String(requestProps.RoleARN.Literal)
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
					opErr = fmt.Errorf("conflict with differently named subscription on prexisting LogGroupName: %s",
						eachFilter.LogGroupName.Literal)

					logger.Error().
						Interface("DescribeSubscriptionResult", describeResult).
						Interface("PutSubscriptionInput", putSubscriptionInput).
						Interface("LogGroupName", eachFilter.LogGroupName).
						Msg(opErr.Error())
				}
			}
		}
		if nil != opErr {
			return nil, opErr
		}
	}
	return nil, opErr
}

// Create implements the create operation
func (command CloudWatchLogsLambdaEventSourceResource) Create(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(true, awsSession, event, logger)
}

// Update implements the update operation
func (command CloudWatchLogsLambdaEventSourceResource) Update(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(true, awsSession, event, logger)
}

// Delete implements the delete operation
func (command CloudWatchLogsLambdaEventSourceResource) Delete(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.updateRegistration(false, awsSession, event, logger)
}
