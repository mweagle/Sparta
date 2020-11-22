package sparta

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	spartaAPIGateway "github.com/mweagle/Sparta/aws/apigateway"
	spartaAWSEvents "github.com/mweagle/Sparta/aws/events"
	"github.com/rs/zerolog"
)

var randVal string

func init() {
	randVal = time.Now().UTC().String()
}

type testRequest struct {
	Message string
	Request spartaAWSEvents.APIGatewayRequest
}

func testAPIGatewayLambda(ctx context.Context,
	gatewayEvent spartaAWSEvents.APIGatewayRequest) (interface{}, error) {
	logger, loggerOk := ctx.Value(ContextKeyLogger).(*zerolog.Logger)
	if loggerOk {
		logger.Info().Msg("Hello world structured log message")
	}

	// Return a message, together with the incoming input...
	return spartaAPIGateway.NewResponse(http.StatusOK, &testRequest{
		Message: fmt.Sprintf("Test %s", randVal),
		Request: gatewayEvent,
	}), nil
}

func TestAPIGatewayRequest(t *testing.T) {
	requestBody := &testRequest{
		Message: randVal,
	}
	mockRequest, mockRequestErr := spartaAWSEvents.NewAPIGatewayMockRequest("helloWorld",
		http.MethodGet,
		nil,
		requestBody)
	if mockRequestErr != nil {
		t.Fatal(mockRequestErr)
	}
	resp, respErr := testAPIGatewayLambda(context.Background(), *mockRequest)
	if respErr != nil {
		t.Fatal(respErr)
	} else {
		t.Log(fmt.Sprintf("%#v", resp))
	}
}

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

func TestAPIV2Gateway(t *testing.T) {
	stage, _ := NewAPIV2Stage("v1")
	apiGateway, _ := NewAPIV2(Websocket,
		"sample",
		"$request.body.message",
		stage)
	lambdaFn, _ := NewAWSLambda(LambdaName(mockLambda1),
		mockLambda1,
		IAMRoleDefinition{})
	apiv2Route, _ := apiGateway.NewAPIV2Route("$connect",
		lambdaFn)
	apiv2Route.OperationName = "ConnectRoute"

	lambdaFn2, _ := NewAWSLambda(LambdaName(mockLambda2),
		mockLambda2,
		IAMRoleDefinition{})
	apiv2Route2, _ := apiGateway.NewAPIV2Route("$disconnect",
		lambdaFn2)
	apiv2Route2.OperationName = "DisconnectRoute"

	testProvisionEx(t,
		[]*LambdaAWSInfo{lambdaFn, lambdaFn2},
		apiGateway,
		nil,
		nil,
		false,
		nil)
}
