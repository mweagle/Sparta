package resources

import (
	"encoding/json"
	"testing"
	"time"

	gocf "github.com/mweagle/go-cloudformation"
	"github.com/sirupsen/logrus"
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
	resHello := gocf.NewResourceByType(HelloWorld)
	customResource := resHello.(*HelloWorldResource)
	customResource.Message = gocf.String("Hello world")
}

func TestCreateHelloWorldNewInstances(t *testing.T) {
	resHello1 := gocf.NewResourceByType(HelloWorld)
	customResource1 := resHello1.(*HelloWorldResource)

	resHello2 := gocf.NewResourceByType(HelloWorld)
	customResource2 := resHello2.(*HelloWorldResource)

	if &customResource1 == &customResource2 {
		t.Errorf("gocf.NewResourceByType failed to make new instances")
	}
}

func TestExecuteCreateHelloWorld(t *testing.T) {
	resHello1 := gocf.NewResourceByType(HelloWorld)
	customResource1 := resHello1.(*HelloWorldResource)
	customResource1.Message = gocf.String("Create resource here")

	logger := logrus.New()
	awsSession := awsSession(logger)
	createOutputs, createError := customResource1.Create(awsSession,
		mockHelloWorldResourceEvent(t),
		logger)
	if nil != createError {
		t.Errorf("Failed to create HelloWorldResource: %s", createError)
	}
	t.Logf("HelloWorldResource outputs: %s", createOutputs)
}
