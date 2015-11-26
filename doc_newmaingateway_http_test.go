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

func echoAPIGatewayHTTPEvent(event *json.RawMessage,
	context *sparta.LambdaContext,
	w http.ResponseWriter,
	logger *logrus.Logger) {

	var lambdaEvent sparta.APIGatewayLambdaJSONEvent
	err := json.Unmarshal([]byte(*event), &lambdaEvent)
	if err != nil {
		logger.Error("Failed to unmarshal event data: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responseBody, err := json.Marshal(lambdaEvent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		fmt.Fprint(w, string(responseBody))
	}
}

// Should be main() in your application
func ExampleMain_apiGatewayHTTPSEvent() {

	// Create the MyEchoAPI API Gateway, with stagename /test.  The associated
	// Stage reesource will cause the API to be deployed.
	apiGateway := sparta.NewAPIGateway("MyEchoHTTPAPI", stage)
	stage := sparta.NewStage("v1")

	// Create a lambda function
	echoAPIGatewayLambdaFn := NewLambda(sparta.IAMRoleDefinition{}, echoAPIGatewayEvent, nil)

	// Associate a URL path component with the Lambda function
	apiGatewayResource, _ := api.NewResource("/echoHelloWorld", echoAPIGatewayLambdaFn)

	// Associate 1 or more HTTP methods with the Resource.
	method := apiGatewayResource.NewMethod("GET")

	// Whitelist query parameters that should be passed to lambda function
	method.Parameters["method.request.querystring.myKey"] = true
	method.Parameters["method.request.querystring.myOtherKey"] = true

	// Start
	Main("HelloWorldLambdaHTTPSService", "Description for Hello World HTTPS Lambda", []*LambdaAWSInfo{echoAPIGatewayLambdaFn}, apiGateway)
}
