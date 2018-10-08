package archetype

import (
	"context"
	"fmt"

	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
	"github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
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

// NewSNSReactor returns an SNS reactor lambda function
func NewSNSReactor(reactor SNSReactor,
	snsTopic gocf.Stringable,
	additionalLambdaPermissions []sparta.IAMRolePrivilege) (*sparta.LambdaAWSInfo, error) {

	reactorLambda := func(ctx context.Context, snsEvent awsLambdaEvents.SNSEvent) (interface{}, error) {
		return reactor.OnSNSEvent(ctx, snsEvent)
	}

	lambdaFn := sparta.HandleAWSLambda(fmt.Sprintf("%T", reactor),
		reactorLambda,
		sparta.IAMRoleDefinition{})

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
