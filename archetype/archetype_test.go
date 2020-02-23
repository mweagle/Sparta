package archetype

import (
	"context"
	"encoding/json"
	"testing"

	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
	sparta "github.com/mweagle/Sparta"
	spartaTesting "github.com/mweagle/Sparta/testing"
	gocf "github.com/mweagle/go-cloudformation"
)

func TestReactorName(t *testing.T) {
	reactor := func() {

	}
	testName := reactorName(reactor)
	if testName == "" {
		t.Fatalf("Failed to create reactor name")
	}
	t.Logf("Created reactor name: %s", testName)
	testName = reactorName(nil)
	if testName == "" {
		t.Fatalf("Failed toc reate reactor name for nil arg")
	}
	t.Logf("Created reactor name: %s", testName)
}

type archetypeTest struct {
}

func (at *archetypeTest) OnS3Event(ctx context.Context,
	event awsLambdaEvents.S3Event) (interface{}, error) {
	return nil, nil
}

func (at *archetypeTest) OnSNSEvent(ctx context.Context,
	snsEvent awsLambdaEvents.SNSEvent) (interface{}, error) {
	return nil, nil
}
func (at *archetypeTest) OnCloudWatchMessage(ctx context.Context,
	cwEvent awsLambdaEvents.CloudwatchLogsEvent) (interface{}, error) {
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

func (at *archetypeTest) OnEventBridgeBroadcast(ctx context.Context,
	msg json.RawMessage) (interface{}, error) {
	return nil, nil
}

////////////////////////////////////////////////////////////////////////////////
/*
  ___ ____
 / __|__ /
 \__ \|_ \
 |___/___/
*/
////////////////////////////////////////////////////////////////////////////////
func TestS3Archetype(t *testing.T) {
	testStruct := &archetypeTest{}

	lambdaFn, lambdaFnErr := NewS3Reactor(testStruct,
		gocf.String("s3Bucket"),
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate S3Reactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)

	lambdaFn, lambdaFnErr = NewS3Reactor(S3ReactorFunc(testStruct.OnS3Event),
		gocf.String("s3Bucket"),
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate S3Reactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)
}

func TestS3ScopedArchetype(t *testing.T) {
	testStruct := &archetypeTest{}

	lambdaFn, lambdaFnErr := NewS3ScopedReactor(testStruct,
		gocf.String("s3Bucket"),
		"/input",
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate S3Reactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)

	lambdaFn, lambdaFnErr = NewS3ScopedReactor(S3ReactorFunc(testStruct.OnS3Event),
		gocf.String("s3Bucket"),
		"/input",
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate S3Reactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)
}

////////////////////////////////////////////////////////////////////////////////
/*
  ___ _  _ ___
 / __| \| / __|
 \__ \ .` \__ \
 |___/_|\_|___/
*/
////////////////////////////////////////////////////////////////////////////////

func TestSNSArchetype(t *testing.T) {
	testStruct := &archetypeTest{}

	lambdaFn, lambdaFnErr := NewSNSReactor(testStruct,
		gocf.String("snsTopic"),
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate SNSReactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)

	lambdaFn, lambdaFnErr = NewSNSReactor(SNSReactorFunc(testStruct.OnSNSEvent),
		gocf.String("s3Bucket"),
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate SNSReactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)

}

////////////////////////////////////////////////////////////////////////////////
/*
  ___                           ___  ___
 |   \ _  _ _ _  __ _ _ __  ___|   \| _ )
 | |) | || | ' \/ _` | '  \/ _ \ |) | _ \
 |___/ \_, |_||_\__,_|_|_|_\___/___/|___/
	   |__/
*/
////////////////////////////////////////////////////////////////////////////////

func TestDynamoDBArchetype(t *testing.T) {
	testStruct := &archetypeTest{}

	lambdaFn, lambdaFnErr := NewDynamoDBReactor(testStruct,
		gocf.String("arn:dynamo"),
		"TRIM_HORIZON",
		10,
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate DynamoDBReactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)

	lambdaFn, lambdaFnErr = NewDynamoDBReactor(DynamoDBReactorFunc(testStruct.OnDynamoEvent),
		gocf.String("arn:dynamo"),
		"TRIM_HORIZON",
		10,
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate DynamoDBReactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)
}

////////////////////////////////////////////////////////////////////////////////
/*
  _  ___             _
 | |/ (_)_ _  ___ __(_)___
 | ' <| | ' \/ -_|_-< (_-<
 |_|\_\_|_||_\___/__/_/__/
*/
////////////////////////////////////////////////////////////////////////////////
func TestKinesisArchetype(t *testing.T) {
	testStruct := &archetypeTest{}

	lambdaFn, lambdaFnErr := NewKinesisReactor(testStruct,
		gocf.String("arn:kinesis"),
		"TRIM_HORIZON",
		10,
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate KinesisReactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)

	lambdaFn, lambdaFnErr = NewKinesisReactor(KinesisReactorFunc(testStruct.OnKinesisMessage),
		gocf.String("arn:kinesis"),
		"TRIM_HORIZON",
		10,
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate KinesisReactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)
}

////////////////////////////////////////////////////////////////////////////////
/*
   ___ _             ___      __    _      _
  / __| |___ _  _ __| \ \    / /_ _| |_ __| |_
 | (__| / _ \ || / _` |\ \/\/ / _` |  _/ _| ' \
  \___|_\___/\_,_\__,_| \_/\_/\__,_|\__\__|_||_|
*/
////////////////////////////////////////////////////////////////////////////////
func TestCloudWatchEvented(t *testing.T) {
	testStruct := &archetypeTest{}

	lambdaFn, lambdaFnErr := NewCloudWatchEventedReactor(testStruct,
		map[string]map[string]interface{}{
			"events": {
				"source":      []string{"aws.ec2"},
				"detail-type": []string{"EC2 Instance state change"},
			}},
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate NewCloudWatchEventedReactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)

	lambdaFn, lambdaFnErr = NewCloudWatchEventedReactor(CloudWatchReactorFunc(testStruct.OnCloudWatchMessage),
		map[string]map[string]interface{}{
			"events": {
				"source":      []string{"aws.ec2"},
				"detail-type": []string{"EC2 Instance state change"},
			}},
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate NewCloudWatchEventedReactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)
}

func TestCloudWatchScheduled(t *testing.T) {
	testStruct := &archetypeTest{}

	lambdaFn, lambdaFnErr := NewCloudWatchScheduledReactor(testStruct,
		map[string]string{
			"every5Mins": "rate(5 minutes)",
		},
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate NewCloudWatchScheduledReactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)

	lambdaFn, lambdaFnErr = NewCloudWatchScheduledReactor(CloudWatchReactorFunc(testStruct.OnCloudWatchMessage),
		map[string]string{
			"every5Mins": "rate(5 minutes)",
		},
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate NewCloudWatchScheduledReactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)
}

////////////////////////////////////////////////////////////////////////////////
/*
  ___             _   ___     _    _
 | __|_ _____ _ _| |_| _ )_ _(_)__| |__ _ ___
 | _|\ V / -_) ' \  _| _ \ '_| / _` / _` / -_)
 |___|\_/\___|_||_\__|___/_| |_\__,_\__, \___|
									|___/
*/
////////////////////////////////////////////////////////////////////////////////
func TestEventBridgeScheduled(t *testing.T) {
	testStruct := &archetypeTest{}

	lambdaFn, lambdaFnErr := NewEventBridgeScheduledReactor(testStruct,
		"rate(5 minutes)",
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate NewEventBridgeScheduledReactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)

	lambdaFn, lambdaFnErr = NewEventBridgeScheduledReactor(EventBridgeReactorFunc(testStruct.OnEventBridgeBroadcast),
		"rate(5 minutes)",
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate NewEventBridgeScheduledReactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)
}

func TestEventBridgeEvented(t *testing.T) {
	testStruct := &archetypeTest{}

	lambdaFn, lambdaFnErr := NewEventBridgeEventReactor(testStruct,
		map[string]interface{}{
			"source": []string{"aws.ec2"},
		},
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate NewEventBridgeEventReactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)

	lambdaFn, lambdaFnErr = NewEventBridgeEventReactor(EventBridgeReactorFunc(testStruct.OnEventBridgeBroadcast),
		map[string]interface{}{
			"source": []string{"aws.ec2"},
		}, nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate NewEventBridgeEventReactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)
}
