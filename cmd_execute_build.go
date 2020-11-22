// +build !lambdabinary

package sparta

import (
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// StampedBuildID is the buildID stamped into the binary. For the case of a
// local build this is set by the provision command and the same value
// is stamped into the cross compiled binary at AWS Lambda execution time
var StampedBuildID string

// Execute creates an HTTP listener to dispatch execution. Typically
// called via Main() via command line arguments.
func Execute(serviceName string,
	lambdaAWSInfos []*LambdaAWSInfo,
	logger *zerolog.Logger) error {
	// Execute no longer supported in non AWS binaries...
	return errors.Errorf("Execute not supported outside of AWS Lambda environment")
}

// awsLambdaFunctionName returns the name of the function, which
// is set in the CloudFormation template that is published
// into the container as `AWS_LAMBDA_FUNCTION_NAME`.  The function name
// is dependent on the CloudFormation stack name so that
// CodePipeline based builds can properly create unique FunctionNAmes
// within an account
func awsLambdaFunctionName(internalFunctionName string) gocf.Stringable {
	sanitizedName := awsLambdaInternalName(internalFunctionName)
	// When we build, we return a gocf.Join that
	// will use the stack name and the internal name. When we run, we're going
	// to use the name discovered from the environment.
	return gocf.Join("",
		gocf.Ref("AWS::StackName"),
		gocf.String(functionNameDelimiter),
		gocf.String(sanitizedName))
}
