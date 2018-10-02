package archetype

import (
	"context"
	"fmt"

	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
	"github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
)

// KinesisReactor represents a lambda function that responds to Kinesis messages
type KinesisReactor interface {
	// OnEvent when an SNS event occurs. Check the snsEvent field
	// for the specific event
	OnKinesisMessage(ctx context.Context,
		kinesisEvent awsLambdaEvents.KinesisEvent) (interface{}, error)
}

// KinesisReactorFunc is a free function that adapts a KinesisReactor
// compliant signature into a function that exposes an OnEvent
// function
type KinesisReactorFunc func(ctx context.Context,
	kinesisEvent awsLambdaEvents.KinesisEvent) (interface{}, error)

// OnKinesisMessage satisfies the KinesisReactor interface
func (reactorFunc KinesisReactorFunc) OnKinesisMessage(ctx context.Context,
	kinesisEvent awsLambdaEvents.KinesisEvent) (interface{}, error) {
	return reactorFunc(ctx, kinesisEvent)
}

// NewKinesisReactor returns an Kinesis reactor lambda function
func NewKinesisReactor(reactor KinesisReactor,
	kinesisStream gocf.Stringable,
	startingPosition string,
	batchSize int64,
	additionalLambdaPermissions []sparta.IAMRolePrivilege) (*sparta.LambdaAWSInfo, error) {

	reactorLambda := func(ctx context.Context, kinesisEvent awsLambdaEvents.KinesisEvent) (interface{}, error) {
		return reactor.OnKinesisMessage(ctx, kinesisEvent)
	}

	lambdaFn := sparta.HandleAWSLambda(fmt.Sprintf("%T", reactor),
		reactorLambda,
		sparta.IAMRoleDefinition{})

	lambdaFn.EventSourceMappings = append(lambdaFn.EventSourceMappings,
		&sparta.EventSourceMapping{
			EventSourceArn:   kinesisStream,
			StartingPosition: startingPosition,
			BatchSize:        batchSize,
		})
	if len(additionalLambdaPermissions) != 0 {
		lambdaFn.RoleDefinition.Privileges = additionalLambdaPermissions
	}
	return lambdaFn, nil
}
