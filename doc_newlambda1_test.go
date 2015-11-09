package sparta

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"net/http"
)

func lambdaHelloWorld(event *json.RawMessage, context *LambdaContext, w *http.ResponseWriter, logger *logrus.Logger) {
	fmt.Fprintf(*w, "Hello World!")
}

func ExampleNewLambda_preexistingIAMRoleName() {
	helloWorldLambda := NewLambda("PreexistingAWSLambdaRoleName", lambdaHelloWorld, nil)
	if nil != helloWorldLambda {
		fmt.Printf("Failed to create new Lambda function")
	}
}
