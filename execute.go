package sparta

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"reflect"
)

// awsLambdaFunctionName returns the name of the function, which
// is set in the CloudFormation template that is published
// into the container as `AWS_LAMBDA_FUNCTION_NAME`. Rather
// than publish custom vars which are editable in the Console,
// tunneling this value through allows Sparta to leverage the
// built in env vars.
func awsLambdaFunctionNameImplementation(serviceName string,
	internalFunctionName string) string {
	// Ok, so we need to scope the functionname with the StackName, otherwise
	// there will be collisions in the account. So how to publish
	// the stack name into the awsbinary?
	// How about
	// Linker flags would be nice...sparta.StampedServiceName ?
	awsLambdaName := fmt.Sprintf("%s-%s",
		serviceName,
		internalFunctionName)
	if len(awsLambdaName) > 64 {
		hash := sha1.New()
		hash.Write([]byte(awsLambdaName))
		awsLambdaName = fmt.Sprintf("Sparta-Lambda-%s", hex.EncodeToString(hash.Sum(nil)))
	}
	// If the name is longer than 64 chars, just hash it
	return sanitizedName(awsLambdaName)
}

func validateArguments(handler reflect.Type) error {
	handlerTakesContext := false
	if handler.NumIn() > 2 {
		return fmt.Errorf("handlers may not take more than two arguments, but handler takes %d", handler.NumIn())
	} else if handler.NumIn() > 0 {
		contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
		argumentType := handler.In(0)
		handlerTakesContext = argumentType.Implements(contextType)
		if handler.NumIn() > 1 && !handlerTakesContext {
			return fmt.Errorf("handler takes two arguments, but the first is not Context. got %s", argumentType.Kind())
		}
	}
	return nil
}
func validateReturns(handler reflect.Type) error {
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if handler.NumOut() > 2 {
		return fmt.Errorf("handler may not return more than two values")
	} else if handler.NumOut() > 1 {
		if !handler.Out(1).Implements(errorType) {
			return fmt.Errorf("handler returns two values, but the second does not implement error")
		}
	} else {
		if !handler.Out(0).Implements(errorType) {
			return fmt.Errorf("handler returns a single value, but it does not implement error")
		}
	}
	return nil
}

func ensureValidSignature(lambdaName string, handlerSymbol interface{}) error {
	handlerType := reflect.TypeOf(handlerSymbol)
	if handlerType == nil {
		return fmt.Errorf("Failed to confirm function type: %#v", handlerSymbol)
	}
	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("Lambda function (%s) is a %s type, not a %s type",
			lambdaName,
			handlerType.Kind(),
			reflect.Func)
	}
	argumentErr := validateArguments(handlerType)
	if argumentErr != nil {
		return fmt.Errorf("Lambda function (%s) has invalid formal arguments: %s",
			lambdaName,
			argumentErr)
	}
	returnsErr := validateReturns(handlerType)
	if returnsErr != nil {
		return fmt.Errorf("Lambda function (%s) has invalid returns: %s",
			lambdaName,
			returnsErr)
	}
	return nil
}
