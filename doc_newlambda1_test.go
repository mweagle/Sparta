package sparta

import (
	"fmt"
	"net/http"
)

func lambdaHelloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

func ExampleNewAWSLambda_preexistingIAMRoleName() {
	helloWorldLambda, _ := NewAWSLambda(LambdaName(lambdaHelloWorld),
		lambdaHelloWorld,
		IAMRoleDefinition{})
	if nil != helloWorldLambda {
		fmt.Printf("Failed to create new Lambda function")
	}
}
