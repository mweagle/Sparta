package cloudtest

import (
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"
)

// stackLambdaLiteralSelector returns a lambda function
type stackLambdaLiteralSelector struct {
	literalName string
}

func (slls *stackLambdaLiteralSelector) Select(t CloudTest) (*lambda.GetFunctionOutput, error) {
	output := cache.getFunction(t, slls.literalName)
	if output == nil {
		return nil, errors.Errorf("Failed to find named function: %s", slls.literalName)
	}
	return output, nil
}

// NewLambdaLiteralSelector returns a new LambdaSelector that just uses
// the hardcoded name
func NewLambdaLiteralSelector(functionName string) LambdaSelector {
	return &stackLambdaLiteralSelector{
		literalName: functionName,
	}
}

// stackLambdasSelector returns a lambda function
type stackLambdasSelector struct {
	stackName    string
	jmesSelector string
}

func (sls *stackLambdasSelector) Select(t CloudTest) (*lambda.GetFunctionOutput, error) {

	functionOutput := cache.getStackFunction(t, sls.stackName, sls.jmesSelector)
	if functionOutput == nil {
		return nil, errors.Errorf("Failed to find AWS Lambda in stack %s for selector: %s",
			sls.stackName,
			sls.jmesSelector)
	}
	return functionOutput, nil
}

// NewStackLambdaSelector returns a new LambdaSelector using the jmesSelector
// expression against all GetFunctionOutputs for that lambda function
func NewStackLambdaSelector(stackName string, jmesSelector string) LambdaSelector {
	return &stackLambdasSelector{
		stackName:    stackName,
		jmesSelector: jmesSelector,
	}
}
