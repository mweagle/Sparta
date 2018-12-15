package archetype

import (
	"context"
	"reflect"
	"runtime"

	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
	sparta "github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
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

// ReactorName provides the name of the reactor func
func (reactorFunc DynamoDBReactorFunc) ReactorName() string {
	return runtime.FuncForPC(reflect.ValueOf(reactorFunc).Pointer()).Name()
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

	lambdaFn, lambdaFnErr := sparta.NewAWSLambda(reactorName(reactor),
		reactorLambda,
		sparta.IAMRoleDefinition{})
	if lambdaFnErr != nil {
		return nil, errors.Wrapf(lambdaFnErr, "attempting to create reactor")
	}

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
