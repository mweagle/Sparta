package archetype

import (
	"context"
	"reflect"
	"runtime"

	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
	"github.com/mweagle/Sparta"
	"github.com/pkg/errors"
)

// CloudWatchLogsReactor represents a lambda function that responds to CW log messages
type CloudWatchLogsReactor interface {
	// OnLogMessage when an SNS event occurs. Check the snsEvent field
	// for the specific event
	OnLogMessage(ctx context.Context,
		cwLogs awsLambdaEvents.CloudwatchLogsEvent) (interface{}, error)
}

// CloudWatchLogsReactorFunc is a free function that adapts a CloudWatchLogsReactor
// compliant signature into a function that exposes an OnEvent
// function
type CloudWatchLogsReactorFunc func(ctx context.Context,
	cwLogs awsLambdaEvents.CloudwatchLogsEvent) (interface{}, error)

// OnLogMessage satisfies the CloudWatchLogsReactor interface
func (reactorFunc CloudWatchLogsReactorFunc) OnLogMessage(ctx context.Context,
	cwLogs awsLambdaEvents.CloudwatchLogsEvent) (interface{}, error) {
	return reactorFunc(ctx, cwLogs)
}

// ReactorName provides the name of the reactor func
func (reactorFunc CloudWatchLogsReactorFunc) ReactorName() string {
	return runtime.FuncForPC(reflect.ValueOf(reactorFunc).Pointer()).Name()
}

// NewCloudWatchLogsReactor returns a CloudWatch logs reactor lambda function
func NewCloudWatchLogsReactor(reactor CloudWatchLogsReactor,
	subscriptions map[string]sparta.CloudWatchEventsRule,
	additionalLambdaPermissions []sparta.IAMRolePrivilege) (*sparta.LambdaAWSInfo, error) {
	if len(subscriptions) <= 0 {
		return nil, errors.Errorf("CloudWatchLogs subscription map must not be empty")
	}

	reactorLambda := func(ctx context.Context, cwLogs awsLambdaEvents.CloudwatchLogsEvent) (interface{}, error) {
		return reactor.OnLogMessage(ctx, cwLogs)
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
