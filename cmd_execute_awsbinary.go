// +build lambdabinary

package sparta

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sync"

	awsLambdaGo "github.com/aws/aws-lambda-go/lambda"
	awsLambdaContext "github.com/aws/aws-lambda-go/lambdacontext"
	spartaAWS "github.com/mweagle/Sparta/aws"
	cloudformationResources "github.com/mweagle/Sparta/aws/cloudformation/resources"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/rs/zerolog"
)

// StampedServiceName is the name stamp
// https://blog.cloudflare.com/setting-go-variables-at-compile-time/
// StampedServiceName is the serviceName stamped into this binary
var StampedServiceName string

// StampedBuildID is the buildID stamped into the binary
var StampedBuildID string

var discoveryInfo *DiscoveryInfo
var once sync.Once

func initDiscoveryInfo() {
	info, _ := Discover()
	discoveryInfo = info
}

func awsLambdaFunctionName(internalFunctionName string) gocf.Stringable {
	// TODO - move this to use SSM so that it's not human editable?
	// But discover information is per-function, not per stack.
	// Could we put the stack discovery info in there?
	once.Do(initDiscoveryInfo)
	sanitizedName := awsLambdaInternalName(internalFunctionName)

	return gocf.String(fmt.Sprintf("%s%s%s",
		discoveryInfo.StackName,
		functionNameDelimiter,
		sanitizedName))
}

func takesContext(handler reflect.Type) bool {
	handlerTakesContext := false
	if handler.NumIn() > 0 {
		contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
		argumentType := handler.In(0)
		handlerTakesContext = argumentType.Implements(contextType)
	}
	return handlerTakesContext
}

// tappedHandler is the handler that represents this binary's mode
func tappedHandler(handlerSymbol interface{},
	interceptors *LambdaEventInterceptors,
	logger *zerolog.Logger) interface{} {

	// If there aren't any, make it a bit easier
	// to call the applyInterceptors function
	if interceptors == nil {
		interceptors = &LambdaEventInterceptors{}
	}

	// Tap the call chain to inject the context params...
	handler := reflect.ValueOf(handlerSymbol)
	handlerType := reflect.TypeOf(handlerSymbol)
	takesContext := takesContext(handlerType)

	// Apply interceptors is a utility function to apply the
	// specified interceptors as part of the lifecycle handler.
	// We can push the specific behaviors into the interceptors
	// and keep this function simple. 🎉
	applyInterceptors := func(ctx context.Context,
		msg json.RawMessage,
		interceptors InterceptorList) context.Context {
		for _, eachInterceptor := range interceptors {
			ctx = eachInterceptor.Interceptor(ctx, msg)
		}
		return ctx
	}

	// How to determine if this handler has tracing enabled? That would be a property
	// of the function template associated with this function.

	// TODO - add Context.Timeout handler to ensure orderly exit
	return func(ctx context.Context, msg json.RawMessage) (interface{}, error) {

		awsSession := spartaAWS.NewSession(logger)
		ctx = applyInterceptors(ctx, msg, interceptors.Begin)
		ctx = context.WithValue(ctx, ContextKeyLogger, logger)
		ctx = context.WithValue(ctx, ContextKeyAWSSession, awsSession)
		ctx = applyInterceptors(ctx, msg, interceptors.BeforeSetup)

		// Create the entry logger that has some context information
		var zerologRequestLogger zerolog.Logger
		lambdaContext, lambdaContextOk := awsLambdaContext.FromContext(ctx)
		if lambdaContextOk {
			zerologRequestLogger = logger.With().
				Str(LogFieldRequestID, lambdaContext.AwsRequestID).
				Str(LogFieldARN, lambdaContext.InvokedFunctionArn).
				Str(LogFieldBuildID, StampedBuildID).
				Str(LogFieldInstanceID, InstanceID()).
				Logger()
		}
		ctx = context.WithValue(ctx, ContextKeyRequestLogger, &zerologRequestLogger)
		ctx = applyInterceptors(ctx, msg, interceptors.AfterSetup)

		// construct arguments
		var args []reflect.Value
		if takesContext {
			args = append(args, reflect.ValueOf(ctx))
		}
		if (handlerType.NumIn() == 1 && !takesContext) ||
			handlerType.NumIn() == 2 {
			eventType := handlerType.In(handlerType.NumIn() - 1)
			event := reflect.New(eventType)
			unmarshalErr := json.Unmarshal(msg, event.Interface())
			if unmarshalErr != nil {
				return nil, unmarshalErr
			}
			args = append(args, event.Elem())
		}
		ctx = applyInterceptors(ctx, msg, interceptors.BeforeDispatch)
		response := handler.Call(args)
		ctx = applyInterceptors(ctx, msg, interceptors.AfterDispatch)

		// If the user function
		// convert return values into (interface{}, error)
		var err error
		if len(response) > 0 {
			if errVal, ok := response[len(response)-1].Interface().(error); ok {
				err = errVal
			}
		}
		ctx = context.WithValue(ctx, ContextKeyLambdaError, err)
		var val interface{}
		if len(response) > 1 {
			val = response[0].Interface()
		}
		ctx = context.WithValue(ctx, ContextKeyLambdaResponse, val)
		applyInterceptors(ctx, msg, interceptors.Complete)
		return val, err
	}
}

// Execute creates an HTTP listener to dispatch execution. Typically
// called via Main() via command line arguments.
func Execute(serviceName string,
	lambdaAWSInfos []*LambdaAWSInfo,
	logger *zerolog.Logger) error {

	logger.Debug().Msg("Initializing discovery service")

	// Initialize the discovery service
	initializeDiscovery(logger)

	// Find the function name based on the dispatch
	// https://docs.aws.amazon.com/lambda/latest/dg/current-supported-versions.html
	requestedLambdaFunctionName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	logger.Debug().
		Str("lambdaName", requestedLambdaFunctionName).
		Msg("Invoking requested lambda")

	// Log any info when we start up...
	logger.Debug().
		Msg("Querying for platform info")
	platformLogSysInfo(requestedLambdaFunctionName, logger)

	// So what if we have workflow hooks in here?
	var interceptors *LambdaEventInterceptors

	/*
		There are three types of targets:
			- User functions
			- User custom resources
			- Sparta custom resources
	*/
	// Based on the environment variable, setup the proper listener...
	var lambdaFunctionName gocf.Stringable
	testAWSName := ""
	var handlerSymbol interface{}
	knownNames := []string{}

	//////////////////////////////////////////////////////////////////////////////
	// User registered commands?
	//////////////////////////////////////////////////////////////////////////////
	logger.Debug().Msg("Checking user-defined lambda functions")
	for _, eachLambdaInfo := range lambdaAWSInfos {
		lambdaFunctionName = awsLambdaFunctionName(eachLambdaInfo.lambdaFunctionName())
		testAWSName = lambdaFunctionName.String().Literal

		knownNames = append(knownNames, testAWSName)
		if requestedLambdaFunctionName == testAWSName {
			handlerSymbol = eachLambdaInfo.handlerSymbol
			interceptors = eachLambdaInfo.Interceptors

		}

		// User defined custom resource handler?
		for _, eachCustomResource := range eachLambdaInfo.customResources {
			lambdaFunctionName = awsLambdaFunctionName(eachCustomResource.userFunctionName)
			testAWSName = lambdaFunctionName.String().Literal
			knownNames = append(knownNames, testAWSName)
			if requestedLambdaFunctionName == testAWSName {
				handlerSymbol = eachCustomResource.handlerSymbol
			}
		}
		if handlerSymbol != nil {
			break
		}
	}

	//////////////////////////////////////////////////////////////////////////////
	// Request to instantiate a CustomResourceHandler that implements
	// the CustomResourceCommand interface?
	//////////////////////////////////////////////////////////////////////////////
	if handlerSymbol == nil {
		logger.Debug().Msg("Checking CustomResourceHandler lambda functions")

		requestCustomResourceType := os.Getenv(EnvVarCustomResourceTypeName)
		if requestCustomResourceType != "" {
			knownNames = append(knownNames, fmt.Sprintf("CloudFormation Custom Resource: %s", requestCustomResourceType))
			logger.Debug().
				Interface("customResourceTypeName", requestCustomResourceType).
				Msg("Checking to see if there is a custom resource")

			resource := gocf.NewResourceByType(requestCustomResourceType)
			if resource != nil {
				// Handler?
				command, commandOk := resource.(cloudformationResources.CustomResourceCommand)
				if !commandOk {
					logger.Error().
						Str("ResourceType", requestCustomResourceType).
						Msg("CloudFormation type doesn't implement cloudformationResources.CustomResourceCommand")
				} else {
					customHandler := cloudformationResources.CloudFormationLambdaCustomResourceHandler(command, logger)
					if customHandler != nil {
						handlerSymbol = customHandler
					}
				}
			} else {
				logger.Error().
					Str("ResourceType", requestCustomResourceType).
					Msg("Failed to create CloudFormation custom resource of type")
			}
		}
	}

	if handlerSymbol == nil {
		errorMessage := fmt.Errorf("No handler found for AWS Lambda function: %s. Registered function name: %#v",
			requestedLambdaFunctionName,
			knownNames)
		logger.Error().Err(errorMessage).Msg("Handler error")
		return errorMessage
	}

	// Startup our version...
	tappedHandler := tappedHandler(handlerSymbol, interceptors, logger)
	awsLambdaGo.Start(tappedHandler)
	return nil
}
