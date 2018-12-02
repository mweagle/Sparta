package archetype

import (
	"context"
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

func TestCloudWatchEvented(t *testing.T) {
	testStruct := &archetypeTest{}

	lambdaFn, lambdaFnErr := NewCloudWatchEventedReactor(testStruct,
		map[string]map[string]interface{}{
			"events": map[string]interface{}{
				"source":      []string{"aws.ec2"},
				"detail-type": []string{"EC2 Instance state change"},
			},
		},
		nil)
	if lambdaFnErr != nil {
		t.Fatalf("Failed to instantiate NewCloudWatchEventedReactor: %s", lambdaFnErr.Error())
	}
	spartaTesting.Provision(t, []*sparta.LambdaAWSInfo{lambdaFn}, nil)

	lambdaFn, lambdaFnErr = NewCloudWatchEventedReactor(CloudWatchReactorFunc(testStruct.OnCloudWatchMessage),
		map[string]map[string]interface{}{
			"events": map[string]interface{}{
				"source":      []string{"aws.ec2"},
				"detail-type": []string{"EC2 Instance state change"},
			},
		},
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
