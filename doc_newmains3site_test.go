package sparta

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

// NOTE: your application MUST use `package main` and define a `main()` function.  The
// example text is to make the documentation compatible with godoc.

func echoS3SiteAPIGatewayEvent(ctx context.Context,
	props map[string]interface{}) (map[string]interface{}, error) {
	lambdaCtx, _ := lambdacontext.FromContext(ctx)
	Logger().Info().
		Str("RequestID", lambdaCtx.AwsRequestID).
		Interface("Properties", props).
		Msg("Lambda event")
	return props, nil
}

// Should be main() in your application
func ExampleMain_s3Site() {

	// Create an API Gateway
	apiStage := NewStage("v1")
	apiGateway := NewAPIGateway("SpartaS3Site", apiStage)
	apiGateway.CORSEnabled = true

	// Create a lambda function
	echoS3SiteAPIGatewayEventLambdaFn, _ := NewAWSLambda(LambdaName(echoS3SiteAPIGatewayEvent),
		echoS3SiteAPIGatewayEvent,
		IAMRoleDefinition{})
	apiGatewayResource, _ := apiGateway.NewResource("/hello", echoS3SiteAPIGatewayEventLambdaFn)
	_, err := apiGatewayResource.NewMethod("GET", http.StatusOK)
	if nil != err {
		panic("Failed to create GET resource")
	}
	// Create an S3 site from the contents in ./site
	s3Site, _ := NewS3Site("./site")

	// Provision everything
	Main("HelloWorldS3SiteService", "Description for S3Site", []*LambdaAWSInfo{echoS3SiteAPIGatewayEventLambdaFn}, apiGateway, s3Site)
}
