package archetype

import (
	"context"
	"fmt"

	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
	"github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
)

// DynamoDBReactor represents a lambda function that responds to Dynamo  messages
type DynamoDBReactor interface {
	// OnEvent when an SNS event occurs. Check the snsEvent field
	// for the specific event
	OnDynamoEvent(ctx context.Context,
		dynamoEvent awsLambdaEvents.DynamoDBEvent) (interface{}, error)
}

// DynamoDBReactorFunc is a free function that adapts a DynamoDBReactor
// compliant signature into a function that exposes an OnEvent
// function
type DynamoDBReactorFunc func(ctx context.Context,
	dynamoEvent awsLambdaEvents.DynamoDBEvent) (interface{}, error)

// OnDynamoEvent satisfies the DynamoDBReactor interface
func (reactorFunc DynamoDBReactorFunc) OnDynamoEvent(ctx context.Context,
	dynamoEvent awsLambdaEvents.DynamoDBEvent) (interface{}, error) {
	return reactorFunc(ctx, dynamoEvent)
}

// NewDynamoDBReactor returns an Kinesis reactor lambda function
func NewDynamoDBReactor(reactor DynamoDBReactor,
	dynamoDBARN gocf.Stringable,
	startingPosition string,
	batchSize int64,
	additionalLambdaPermissions []sparta.IAMRolePrivilege) (*sparta.LambdaAWSInfo, error) {

	reactorLambda := func(ctx context.Context, dynamoEvent awsLambdaEvents.DynamoDBEvent) (interface{}, error) {
		return reactor.OnDynamoEvent(ctx, dynamoEvent)
	}

	lambdaFn := sparta.HandleAWSLambda(fmt.Sprintf("%T", reactor),
		reactorLambda,
		sparta.IAMRoleDefinition{})

	lambdaFn.EventSourceMappings = append(lambdaFn.EventSourceMappings,
		&sparta.EventSourceMapping{
			EventSourceArn:   dynamoDBARN,
			StartingPosition: startingPosition,
			BatchSize:        batchSize,
		})
	if len(additionalLambdaPermissions) != 0 {
		lambdaFn.RoleDefinition.Privileges = additionalLambdaPermissions
	}
	return lambdaFn, nil
}
