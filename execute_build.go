// +build !lambdabinary

package sparta

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// Execute creates an HTTP listener to dispatch execution. Typically
// called via Main() via command line arguments.
func Execute(serviceName string,
	lambdaAWSInfos []*LambdaAWSInfo,
	logger *logrus.Logger) error {
	// Execute no longer supported in non AWS binaries...
	return fmt.Errorf("Execute not supported outside of AWS Lambda environment")
}

// awsLambdaFunctionName returns the name of the function, which
// is set in the CloudFormation template that is published
// into the container as `AWS_LAMBDA_FUNCTION_NAME`. Rather
// than publish custom vars which are editable in the Console,
// tunneling this value through allows Sparta to leverage the
// built in env vars.
func awsLambdaFunctionName(serviceName string,
	internalFunctionName string) string {
	return awsLambdaFunctionNameImplementation(serviceName, internalFunctionName)
}
