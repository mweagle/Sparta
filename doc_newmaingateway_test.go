package sparta

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/rs/zerolog"
)

// NOTE: your application MUST use `package main` and define a `main()` function.  The
// example text is to make the documentation compatible with godoc.

func echoAPIGatewayEvent(ctx context.Context,
	props map[string]interface{}) error {
	lambdaCtx, _ := lambdacontext.FromContext(ctx)
	logger, _ := ctx.Value(ContextKeyLogger).(*zerolog.Logger)
	logger.Info().
		Str("RequestID", lambdaCtx.AwsRequestID).
		Interface("Properties", props).
		Msg("Lambda event")
	return nil
}

// Should be main() in your application
func ExampleMain_apiGateway() {

	// Create the MyEchoAPI API Gateway, with stagename /test.  The associated
	// Stage reesource will cause the API to be deployed.
	stage := NewStage("test")
	apiGateway := NewAPIGateway("MyEchoAPI", stage)

	// Create a lambda function
	echoAPIGatewayLambdaFn, _ := NewAWSLambda(LambdaName(echoAPIGatewayEvent),
		echoAPIGatewayEvent,
		IAMRoleDefinition{})

	// Associate a URL path component with the Lambda function
	apiGatewayResource, _ := apiGateway.NewResource("/echoHelloWorld", echoAPIGatewayLambdaFn)

	// Associate 1 or more HTTP methods with the Resource.
	apiGatewayResource.NewMethod("GET", http.StatusOK)

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
	Main("HelloWorldLambdaService", "Description for Hello World Lambda", []*LambdaAWSInfo{echoAPIGatewayLambdaFn}, apiGateway, nil)
}
