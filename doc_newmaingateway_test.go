package sparta

import (
	"Sparta"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
)

// NOTE: your application MUST use `package main` and define a `main()` function.  The
// example text is to make the documentation compatible with godoc.

func echoAPIGatewayEvent(event *json.RawMessage,
	context *sparta.LambdaContext,
	w http.ResponseWriter,
	logger *logrus.Logger) {
	logger.WithFields(logrus.Fields{
		"RequestID": context.AWSRequestID,
		"Event":     string(*event),
	}).Info("Request received")

	fmt.Fprintf(w, "Hello World!")
}

// Should be main() in your application
func ExampleMain_apiGateway() {

	// Create the MyEchoAPI API Gateway, with stagename /test.  The associated
	// Stage reesource will cause the API to be deployed.
	apiGateway := sparta.NewAPIGateway("MyEchoAPI", stage)
	stage := sparta.NewStage("test")

	// Create a lambda function
	echoAPIGatewayLambdaFn := NewLambda(sparta.IAMRoleDefinition{}, echoAPIGatewayEvent, nil)

	// Associate a URL path component with the Lambda function
	apiGatewayResource, _ := api.NewResource("/echoHelloWorld", echoAPIGatewayLambdaFn)

	// Associate 1 or more HTTP methods with the Resource.
	apiGatewayResource.NewMethod("GET")

	// After the stack is deployed, the
	// echoAPIGatewayEvent lambda function will be available at:
	// https://{RestApiID}.execute-api.{AWSRegion}.amazonaws.com/test
	//
	// The dynamically generated URL will be written to STDOUT as part of stack provisioning as in:
	//
	//	Outputs: [{
	//      Description: "API Gateway URL",
	//      OutputKey: "URL",
	//      OutputValue: "https://zdjfwrcao7.execute-api.us-west-2.amazonaws.com/test"
	//    }]
	// eg:
	// 	curl -vs https://zdjfwrcao7.execute-api.us-west-2.amazonaws.com/test/echoHelloWorld

	// Start
	Main("HelloWorldLambdaService", "Description for Hello World Lambda", []*LambdaAWSInfo{echoAPIGatewayLambdaFn}, apiGateway)
}
