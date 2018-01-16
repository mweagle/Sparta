package sparta

import "context"

// NOTE: your application MUST use `package main` and define a `main()` function.  The
// example text is to make the documentation compatible with godoc.
// Should be main() in your application

func mainHelloWorld(ctx context.Context) (string, error) {
	return "Hello World!", nil
}

func ExampleMain_basic() {
	var lambdaFunctions []*LambdaAWSInfo
	helloWorldLambda := HandleAWSLambda("PreexistingAWSLambdaRoleName",
		mainHelloWorld,
		IAMRoleDefinition{})

	lambdaFunctions = append(lambdaFunctions, helloWorldLambda)
	Main("HelloWorldLambdaService", "Description for Hello World Lambda", lambdaFunctions, nil, nil)
}
