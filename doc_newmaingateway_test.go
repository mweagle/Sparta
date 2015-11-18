package sparta

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
)

// NOTE: your application MUST use `package main` and define a `main()` function.  The
// example text is to make the documentation compatible with godoc.

func mainHelloWorldGateway(event *json.RawMessage, context *LambdaContext, w *http.ResponseWriter, logger *logrus.Logger) {
	fmt.Fprintf(*w, "Hello World!")
}

// Should be main() in your application
//
// The compiled application then supports self-deployment (and update), description,
// and execution via command line arguments:
//
// $ ./MyApplication
//
// Usage: MyApplication [global options] <verb> [verb options]
//
// Global options:
//         -l, --level    Log level [panic, fatal, error, warn, info, debug] (default: info)
//         -h, --help     Show this help
//
// Verbs:
//     provision:
//         -b, --s3Bucket S3 Bucket to use for Lambda source (*)
//     describe:
//         -o, --out      Output file for HTML description (*)
//     execute:
//         -p, --port     Alternative port for HTTP binding
//         -s, --signal   Process ID to signal with SIGUSR2 once ready
func ExampleMain_Gateway() {
	var lambdaFunctions []*LambdaAWSInfo
	helloWorldLambda := NewLambda("PreexistingAWSLambdaRoleName", mainHelloWorldGateway, nil)
	lambdaFunctions = append(lambdaFunctions, helloWorldLambda)

	// TODO: Add the API Gateway sample
	Main("HelloWorldLambdaService", "Description for Hello World Lambda", lambdaFunctions, nil)
}
