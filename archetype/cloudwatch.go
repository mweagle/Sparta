package archetype

import (
	"context"
	"reflect"
	"runtime"

	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
	sparta "github.com/mweagle/Sparta"
	"github.com/pkg/errors"
)

// CloudWatchReactor represents a lambda function that responds to CW messages
type CloudWatchReactor interface {
	// OnLogMessage when an SNS event occurs. Check the snsEvent field
	// for the specific event
	OnCloudWatchMessage(ctx context.Context,
		cwLogs awsLambdaEvents.CloudwatchLogsEvent) (interface{}, error)
}

// CloudWatchReactorFunc is a free function that adapts a CloudWatchReactor
// compliant signature into a function that exposes an OnEvent
// function
type CloudWatchReactorFunc func(ctx context.Context,
	cwLogs awsLambdaEvents.CloudwatchLogsEvent) (interface{}, error)

// OnCloudWatchMessage satisfies the CloudWatchReactor interface
func (reactorFunc CloudWatchReactorFunc) OnCloudWatchMessage(ctx context.Context,
	cwLogs awsLambdaEvents.CloudwatchLogsEvent) (interface{}, error) {
	return reactorFunc(ctx, cwLogs)
}

// ReactorName provides the name of the reactor func
func (reactorFunc CloudWatchReactorFunc) ReactorName() string {
	return runtime.FuncForPC(reflect.ValueOf(reactorFunc).Pointer()).Name()
}

// NewCloudWatchEventedReactor returns a CloudWatch logs reactor lambda function
// that executes in response to the given events. The eventPatterns map is a map of names
// to map[string]interface{} values that represents the events to listen to. See
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/CloudWatchEventsandEventPatterns.html
// for the proper syntax. Example:
// 	map[string]interface{}{
//		"source":      []string{"aws.ec2"},
//		"detail-type": []string{"EC2 Instance state change"},
//	}
func NewCloudWatchEventedReactor(reactor CloudWatchReactor,
	eventPatterns map[string]map[string]interface{},
	additionalLambdaPermissions []sparta.IAMRolePrivilege) (*sparta.LambdaAWSInfo, error) {

	subscriptions := make(map[string]sparta.CloudWatchEventsRule)
	for eachName, eachPattern := range eventPatterns {
		subscriptions[eachName] = sparta.CloudWatchEventsRule{
			EventPattern: eachPattern,
		}
	}
	return NewCloudWatchReactor(reactor, subscriptions, additionalLambdaPermissions)
}

// NewCloudWatchScheduledReactor returns a CloudWatch logs reactor lambda function
// that executes with the given schedule. The cronSchedules map is a map of names
// to ScheduleExpressions. See
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html#RateExpressions
// for the proper syntax. Example:
// 	"rate(5 minutes)"
//
func NewCloudWatchScheduledReactor(reactor CloudWatchReactor,
	cronSchedules map[string]string,
	additionalLambdaPermissions []sparta.IAMRolePrivilege) (*sparta.LambdaAWSInfo, error) {

	subscriptions := make(map[string]sparta.CloudWatchEventsRule)
	for eachName, eachSchedule := range cronSchedules {
		subscriptions[eachName] = sparta.CloudWatchEventsRule{
			ScheduleExpression: eachSchedule,
		}
	}
	return NewCloudWatchReactor(reactor, subscriptions, additionalLambdaPermissions)
}

// NewCloudWatchReactor returns a CloudWatch logs reactor lambda function
func NewCloudWatchReactor(reactor CloudWatchReactor,
	subscriptions map[string]sparta.CloudWatchEventsRule,
	additionalLambdaPermissions []sparta.IAMRolePrivilege) (*sparta.LambdaAWSInfo, error) {
	if len(subscriptions) <= 0 {
		return nil, errors.Errorf("CloudWatchLogs subscription map must not be empty")
	}

	reactorLambda := func(ctx context.Context, cwLogs awsLambdaEvents.CloudwatchLogsEvent) (interface{}, error) {
		return reactor.OnCloudWatchMessage(ctx, cwLogs)
	}
	lambdaFn, lambdaFnErr := sparta.NewAWSLambda(reactorName(reactor),
		reactorLambda,
		sparta.IAMRoleDefinition{})
	if lambdaFnErr != nil {
		return nil, errors.Wrapf(lambdaFnErr, "attempting to create reactor")
	}
	cloudWatchEventsPermission := sparta.CloudWatchEventsPermission{}
	cloudWatchEventsPermission.Rules = make(map[string]sparta.CloudWatchEventsRule, 0)
	for eachRuleName, eachRule := range subscriptions {
		cloudWatchEventsPermission.Rules[eachRuleName] = eachRule
	}
	lambdaFn.Permissions = append(lambdaFn.Permissions, cloudWatchEventsPermission)

	if len(additionalLambdaPermissions) != 0 {
		lambdaFn.RoleDefinition.Privileges = additionalLambdaPermissions
	}
	return lambdaFn, nil
}
