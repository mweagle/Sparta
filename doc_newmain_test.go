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
	helloWorldLambda, _ := NewAWSLambda("PreexistingAWSLambdaRoleName",
		mainHelloWorld,
		IAMRoleDefinition{})

	lambdaFunctions = append(lambdaFunctions, helloWorldLambda)
	mainErr := Main("HelloWorldLambdaService", "Description for Hello World Lambda", lambdaFunctions, nil, nil)
	if mainErr != nil {
		panic("Failed to invoke sparta.Main: %s" + mainErr.Error())
	}
}
