package sparta

import (
	"net/http"
	"testing"
)

func TestAPIGateway(t *testing.T) {
	stage := NewStage("v1")
	apiGateway := NewAPIGateway("SpartaAPIGateway", stage)
	lambdaFn, _ := NewAWSLambda(LambdaName(mockLambda1),
		mockLambda1,
		IAMRoleDefinition{})

	// Register the function with the API Gateway
	apiGatewayResource, _ := apiGateway.NewResource("/test", lambdaFn)
	apiGatewayResource.NewMethod("GET", http.StatusOK)

	testProvisionEx(t,
		[]*LambdaAWSInfo{lambdaFn},
		apiGateway,
		nil,
		nil,
		false,
		nil)
}
