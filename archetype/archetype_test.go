package archetype

import (
	"context"
	"testing"

	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
	gocf "github.com/mweagle/go-cloudformation"
)

type archetypeTest struct {
}

func (at *archetypeTest) OnS3Event(ctx context.Context, event awsLambdaEvents.S3Event) (interface{}, error) {
	return nil, nil
}

func (at *archetypeTest) OnSNSEvent(ctx context.Context, snsEvent awsLambdaEvents.SNSEvent) (interface{}, error) {
	return nil, nil
}

func (at *archetypeTest) OnDynamoEvent(ctx context.Context,
	dynamoEvent awsLambdaEvents.DynamoDBEvent) (interface{}, error) {
	return nil, nil
}

func (at *archetypeTest) OnKinesisMessage(ctx context.Context,
	kinesisEvent awsLambdaEvents.KinesisEvent) (interface{}, error) {
	return nil, nil
}

func TestS3Archetype(t *testing.T) {
	testStruct := &archetypeTest{}

	_, lambdaFnErr := NewS3Reactor(testStruct,
		gocf.String("s3Bucket"),
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate S3Reactor: %s", lambdaFnErr.Error())
	}

	_, lambdaFnErr = NewS3Reactor(S3ReactorFunc(testStruct.OnS3Event),
		gocf.String("s3Bucket"),
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate S3Reactor: %s", lambdaFnErr.Error())
	}
}

func TestSNSArchetype(t *testing.T) {
	testStruct := &archetypeTest{}

	_, lambdaFnErr := NewSNSReactor(testStruct,
		gocf.String("snsTopic"),
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate SNSReactor: %s", lambdaFnErr.Error())
	}

	_, lambdaFnErr = NewSNSReactor(SNSReactorFunc(testStruct.OnSNSEvent),
		gocf.String("s3Bucket"),
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate SNSReactor: %s", lambdaFnErr.Error())
	}
}

func TestDynamoDBArchetype(t *testing.T) {
	testStruct := &archetypeTest{}

	_, lambdaFnErr := NewDynamoDBReactor(testStruct,
		gocf.String("arn:dynamo"),
		"TRIM_HORIZON",
		10,
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate DynamoDBReactor: %s", lambdaFnErr.Error())
	}

	_, lambdaFnErr = NewDynamoDBReactor(DynamoDBReactorFunc(testStruct.OnDynamoEvent),
		gocf.String("arn:dynamo"),
		"TRIM_HORIZON",
		10,
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate DynamoDBReactor: %s", lambdaFnErr.Error())
	}
}

func TestKinesisArchetype(t *testing.T) {
	testStruct := &archetypeTest{}

	_, lambdaFnErr := NewKinesisReactor(testStruct,
		gocf.String("arn:kinesis"),
		"TRIM_HORIZON",
		10,
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate KinesisReactor: %s", lambdaFnErr.Error())
	}

	_, lambdaFnErr = NewKinesisReactor(KinesisReactorFunc(testStruct.OnKinesisMessage),
		gocf.String("arn:kinesis"),
		"TRIM_HORIZON",
		10,
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate KinesisReactor: %s", lambdaFnErr.Error())
	}
}
