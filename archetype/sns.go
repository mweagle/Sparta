package archetype

import (
	"context"
	"reflect"
	"runtime"

	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
	"github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
)

// SNSReactor represents a lambda function that responds to typical SNS events
type SNSReactor interface {
	// OnSNSEvent when an SNS event occurs. Check the snsEvent field
	// for the specific event
	OnSNSEvent(ctx context.Context, snsEvent awsLambdaEvents.SNSEvent) (interface{}, error)
}

// SNSReactorFunc is a free function that adapts a SNSReactor
// compliant signature into a function that exposes an OnEvent
// function
type SNSReactorFunc func(ctx context.Context,
	snsEvent awsLambdaEvents.SNSEvent) (interface{}, error)

// OnSNSEvent satisfies the SNSReactor interface
func (reactorFunc SNSReactorFunc) OnSNSEvent(ctx context.Context,
	snsEvent awsLambdaEvents.SNSEvent) (interface{}, error) {
	return reactorFunc(ctx, snsEvent)
}

// ReactorName provides the name of the reactor func
func (reactorFunc SNSReactorFunc) ReactorName() string {
	return runtime.FuncForPC(reflect.ValueOf(reactorFunc).Pointer()).Name()
}

// NewSNSReactor returns an SNS reactor lambda function
func NewSNSReactor(reactor SNSReactor,
	snsTopic gocf.Stringable,
	additionalLambdaPermissions []sparta.IAMRolePrivilege) (*sparta.LambdaAWSInfo, error) {

	reactorLambda := func(ctx context.Context, snsEvent awsLambdaEvents.SNSEvent) (interface{}, error) {
		return reactor.OnSNSEvent(ctx, snsEvent)
	}

	lambdaFn, lambdaFnErr := sparta.NewAWSLambda(reactorName(reactor),
		reactorLambda,
		sparta.IAMRoleDefinition{})
	if lambdaFnErr != nil {
		return nil, errors.Wrapf(lambdaFnErr, "attempting to create reactor")
	}

	lambdaFn.Permissions = append(lambdaFn.Permissions, sparta.SNSPermission{
		BasePermission: sparta.BasePermission{
			SourceArn: snsTopic,
		},
	})
	if len(additionalLambdaPermissions) != 0 {
		lambdaFn.RoleDefinition.Privileges = additionalLambdaPermissions
	}
	return lambdaFn, nil
}
