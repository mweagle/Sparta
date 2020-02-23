package decorator

import (
	"context"
	"net/http"
	"testing"

	sparta "github.com/mweagle/Sparta"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	spartaAWSEvents "github.com/mweagle/Sparta/aws/events"
	spartaTesting "github.com/mweagle/Sparta/testing"
	gocf "github.com/mweagle/go-cloudformation"
)

func TestAPIGatewayCustomDomain(t *testing.T) {
	helloWorld := func(ctx context.Context,
		gatewayEvent spartaAWSEvents.APIGatewayRequest) (interface{}, error) {
		return "Hello World", nil
	}
	lambdaFuncs := func(api *sparta.API) []*sparta.LambdaAWSInfo {
		var lambdaFunctions []*sparta.LambdaAWSInfo
		lambdaFn := sparta.HandleAWSLambda(sparta.LambdaName(helloWorld),
			helloWorld,
			sparta.IAMRoleDefinition{})
		apiGatewayResource, _ := api.NewResource("/hello", lambdaFn)

		// We only return http.StatusOK
		apiMethod, apiMethodErr := apiGatewayResource.NewMethod("GET",
			http.StatusOK,
			http.StatusInternalServerError)
		if nil != apiMethodErr {
			panic("Failed to create /hello resource: " + apiMethodErr.Error())
		}
		// The lambda resource only supports application/json Unmarshallable
		// requests.
		apiMethod.SupportedRequestContentTypes = []string{"application/json"}
		return append(lambdaFunctions, lambdaFn)
	}

	apigatewayHooks := func(apiGateway *sparta.API) *sparta.WorkflowHooks {
		hooks := &sparta.WorkflowHooks{}

		serviceDecorator := APIGatewayDomainDecorator(apiGateway,
			gocf.String("arn:aws:acm:us-west-2:123412341234:certificate/6486C3FF-A3B7-46B6-83A0-9AE329FEC4E3"),
			"", // Optional base path value
			"noice.spartademo.net")
		hooks.ServiceDecorators = []sparta.ServiceDecoratorHookHandler{
			serviceDecorator,
		}
		return hooks
	}

	// Register the function with the API Gateway
	apiStage := sparta.NewStage("v1")

	apiGateway := sparta.NewAPIGateway("SpartaHTMLDomain", apiStage)
	apiGateway.EndpointConfiguration = &gocf.APIGatewayRestAPIEndpointConfiguration{
		Types: gocf.StringList(
			gocf.String("REGIONAL"),
		),
	}
	hooks := apigatewayHooks(apiGateway)
	// Deploy it
	spartaTesting.ProvisionEx(t,
		lambdaFuncs(apiGateway),
		apiGateway,
		nil,
		hooks,
		false,
		nil)
}

func ExampleAPIGatewayDomainDecorator() {
	helloWorld := func(ctx context.Context,
		gatewayEvent spartaAWSEvents.APIGatewayRequest) (interface{}, error) {
		return "Hello World", nil
	}
	lambdaFuncs := func(api *sparta.API) []*sparta.LambdaAWSInfo {
		var lambdaFunctions []*sparta.LambdaAWSInfo
		lambdaFn := sparta.HandleAWSLambda(sparta.LambdaName(helloWorld),
			helloWorld,
			sparta.IAMRoleDefinition{})
		apiGatewayResource, _ := api.NewResource("/hello", lambdaFn)

		// We only return http.StatusOK
		apiMethod, apiMethodErr := apiGatewayResource.NewMethod("GET",
			http.StatusOK,
			http.StatusInternalServerError)
		if nil != apiMethodErr {
			panic("Failed to create /hello resource: " + apiMethodErr.Error())
		}
		// The lambda resource only supports application/json Unmarshallable
		// requests.
		apiMethod.SupportedRequestContentTypes = []string{"application/json"}
		return append(lambdaFunctions, lambdaFn)
	}

	apigatewayHooks := func(apiGateway *sparta.API) *sparta.WorkflowHooks {
		hooks := &sparta.WorkflowHooks{}

		serviceDecorator := APIGatewayDomainDecorator(apiGateway,
			gocf.String("arn:aws:acm:us-west-2:123412341234:certificate/6486C3FF-A3B7-46B6-83A0-9AE329FEC4E3"),
			"", // Optional base path value
			"noice.spartademo.net")
		hooks.ServiceDecorators = []sparta.ServiceDecoratorHookHandler{
			serviceDecorator,
		}
		return hooks
	}

	// Register the function with the API Gateway
	apiStage := sparta.NewStage("v1")

	apiGateway := sparta.NewAPIGateway("SpartaHTMLDomain", apiStage)
	apiGateway.EndpointConfiguration = &gocf.APIGatewayRestAPIEndpointConfiguration{
		Types: gocf.StringList(
			gocf.String("REGIONAL"),
		),
	}
	hooks := apigatewayHooks(apiGateway)
	// Deploy it
	stackName := spartaCF.UserScopedStackName("CustomAPIGateway")
	sparta.MainEx(stackName,
		"CustomAPIGateway defines a stack with a custom APIGateway Domain Name",
		lambdaFuncs(apiGateway),
		apiGateway,
		nil,
		hooks,
		false)
}
