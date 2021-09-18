package resources

import (
	"encoding/json"
	"os"
	"testing"
	"time"

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
	resHello := CustomResourceForType(HelloWorld)
	customResource := resHello.(*HelloWorldResource)
	customResource.Message = "Hello world"
}

func TestCreateHelloWorldNewInstances(t *testing.T) {
	resHello1 := CustomResourceForType(HelloWorld)
	customResource1 := resHello1.(*HelloWorldResource)

	resHello2 := CustomResourceForType(HelloWorld)
	customResource2 := resHello2.(*HelloWorldResource)

	if &customResource1 == &customResource2 {
		t.Errorf("CustomResourceForType failed to make new instances")
	}
}

func TestExecuteCreateHelloWorld(t *testing.T) {
	resHello1 := CustomResourceForType(HelloWorld)
	customResource1 := resHello1.(*HelloWorldResource)
	customResource1.Message = "Create resource here"

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	awsSession := awsSession(&logger)
	createOutputs, createError := customResource1.Create(awsSession,
		mockHelloWorldResourceEvent(t),
		&logger)
	if nil != createError {
		t.Errorf("Failed to create HelloWorldResource: %s", createError)
	}
	t.Logf("HelloWorldResource outputs: %s", createOutputs)
}
