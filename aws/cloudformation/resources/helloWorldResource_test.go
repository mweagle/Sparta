package resources

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	cwCustomProvider "github.com/mweagle/Sparta/aws/cloudformation/provider"

	"github.com/rs/zerolog"
)

func mockHelloWorldResourceEvent(t *testing.T) *CloudFormationLambdaEvent {
	props := map[string]interface{}{
		"ServiceToken": "arn:aws:lambda:us-west-2:123412341234:function:SpartaApplication-S3CustomResourced9468234fca3ffb5-18V7808Y2VSHY",
		"Message":      "World",
	}
	bytes, bytesErr := json.Marshal(props)
	if bytesErr != nil {
		t.Fatalf("Failed to serialize mock custom resource event")
	}

	return &CloudFormationLambdaEvent{
		RequestType:        CreateOperation,
		RequestID:          time.Now().String(),
		StackID:            "1234567890",
		LogicalResourceID:  "logicalID",
		ResourceProperties: json.RawMessage(bytes),
	}
}

func TestCreateHelloWorld(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	resHello, _ := cwCustomProvider.NewCloudFormationCustomResource(HelloWorld, &logger)
	customResource := resHello.(*HelloWorldResource)
	customResource.Properties = ToCustomResourceProperties(&HelloWorldResourceRequest{
		Message: "Hello world",
	})
}

func TestCreateHelloWorldNewInstances(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	resHello1, _ := cwCustomProvider.NewCloudFormationCustomResource(HelloWorld, &logger)
	customResource1 := resHello1.(*HelloWorldResource)

	resHello2, _ := cwCustomProvider.NewCloudFormationCustomResource(HelloWorld, &logger)
	customResource2 := resHello2.(*HelloWorldResource)

	if &customResource1 == &customResource2 {
		t.Errorf("CustomResourceForType failed to make new instances")
	}
}

func TestExecuteCreateHelloWorld(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	resHello1, _ := cwCustomProvider.NewCloudFormationCustomResource(HelloWorld, &logger)
	customResource1 := resHello1.(*HelloWorldResource)
	customResource1.Properties = ToCustomResourceProperties(&HelloWorldResourceRequest{
		Message: "Hello world",
	})

	awsConfig := newAWSConfig(&logger)
	createOutputs, createError := customResource1.Create(awsConfig,
		mockHelloWorldResourceEvent(t),
		&logger)
	if nil != createError {
		t.Errorf("Failed to create HelloWorldResource: %s", createError)
	}
	t.Logf("HelloWorldResource outputs: %s", createOutputs)
}
