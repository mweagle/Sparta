package sparta

import (
	"context"
	"reflect"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var (
	reSplitCustomType = regexp.MustCompile(`\:+`)

	reSplitFunctionName = regexp.MustCompile(`\W+`)
)

const (
	// LogFieldRequestID is the fieldname in the context-scoped logger
	// that includes the AWS assigned request ID
	LogFieldRequestID = "reqID"
	// LogFieldARN is the InvokedFunctionArn value
	LogFieldARN = "arn"
	// LogFieldBuildID is the Build ID stamped into the binary exposing
	// the lambda functions
	LogFieldBuildID = "buildID"
	// LogFieldInstanceID is a unique identifier for a given container instance
	LogFieldInstanceID = "instanceID"
)

const functionNameDelimiter = "_"

// awsLambdaFunctionName returns the name of the function, which
// is set in the CloudFormation template that is published
// into the container as `AWS_LAMBDA_FUNCTION_NAME`. Rather
// than publish custom vars which are editable in the Console,
// tunneling this value through allows Sparta to leverage the
// built in env vars.
func awsLambdaInternalName(internalFunctionName string) string {

	var internalNameParts []string

	// If this is something that implements something else, trim the
	// leading *
	internalFunctionName = strings.TrimPrefix(internalFunctionName, "*")
	customTypeParts := reSplitCustomType.Split(internalFunctionName, -1)
	if len(customTypeParts) > 1 {
		internalNameParts = []string{customTypeParts[len(customTypeParts)-1]}
	} else {
		internalNameParts = reSplitFunctionName.Split(internalFunctionName, -1)
	}
	return strings.Join(internalNameParts, functionNameDelimiter)
}

func validateArguments(handler reflect.Type) error {
	switch handler.Kind() {
	case reflect.Func:
		if handler.NumIn() > 2 {
			return errors.Errorf("handlers may not take more than two arguments, but handler takes %d", handler.NumIn())
		} else if handler.NumIn() > 0 {
			contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
			argumentType := handler.In(0)
			handlerTakesContext := argumentType.Implements(contextType)
			if handler.NumIn() > 1 && !handlerTakesContext {
				return errors.Errorf("handler takes two arguments, but the first is not Context. got %s", argumentType.Kind())
			}
		}
	default:
		// NOP
	}
	return nil
}

func validateReturns(handler reflect.Type) error {
	switch handler.Kind() {
	case reflect.Func:
		errorType := reflect.TypeOf((*error)(nil)).Elem()
		if handler.NumOut() > 2 {
			return errors.Errorf("handler may not return more than two values")
		} else if handler.NumOut() > 1 {
			if !handler.Out(1).Implements(errorType) {
				return errors.Errorf("handler returns two values, but the second does not implement error")
			}
		} else {
			if !handler.Out(0).Implements(errorType) {
				return errors.Errorf("handler returns a single value, but it does not implement error")
			}
		}
	default:
		// NOP
	}
	return nil
}

func ensureValidSignature(lambdaName string, handlerSymbol interface{}) error {
	handlerType := reflect.TypeOf(handlerSymbol)
	if handlerType == nil {
		return errors.Errorf("Failed to confirm function type: %#v", handlerSymbol)
	}
	if handlerType.Kind() != reflect.Func {
		return errors.Errorf("Lambda function (%s) is a (%s) type, not a (%s) type",
			lambdaName,
			handlerType.Kind(),
			reflect.Func)
	}
	argumentErr := validateArguments(handlerType)
	if argumentErr != nil {
		return errors.Errorf("Lambda function (%s) has invalid formal arguments: %s",
			lambdaName,
			argumentErr)
	}
	returnsErr := validateReturns(handlerType)
	if returnsErr != nil {
		return errors.Errorf("Lambda function (%s) has invalid returns: %s",
			lambdaName,
			returnsErr)
	}
	return nil
}
