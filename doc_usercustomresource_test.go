package sparta

import (
	"context"

	"github.com/aws/aws-lambda-go/lambdacontext"
	awsLambdaCtx "github.com/aws/aws-lambda-go/lambdacontext"
	spartaCFResources "github.com/mweagle/Sparta/aws/cloudformation/resources"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/sirupsen/logrus"
)

// Standard AWS λ function
func helloWorld(ctx context.Context,
	props map[string]interface{}) (string, error) {
	lambdaCtx, _ := lambdacontext.FromContext(ctx)
	Logger().WithFields(logrus.Fields{
		"RequestID":  lambdaCtx.AwsRequestID,
		"Properties": props,
	}).Info("Lambda event")
	return "Event processed", nil
}

// User defined λ-backed CloudFormation CustomResource
func userDefinedCustomResource(ctx context.Context,
	event spartaCFResources.CloudFormationLambdaEvent) (map[string]interface{}, error) {

	logger, _ := ctx.Value(ContextKeyLogger).(*logrus.Logger)
	lambdaCtx, _ := awsLambdaCtx.FromContext(ctx)

	var opResults = map[string]interface{}{
		"CustomResourceResult": "Victory!",
	}

	opErr := spartaCFResources.SendCloudFormationResponse(lambdaCtx,
		&event,
		opResults,
		nil,
		logger)
	return opResults, opErr
}

func ExampleLambdaAWSInfo_RequireCustomResource() {

	lambdaFn := HandleAWSLambda(LambdaName(helloWorld),
		helloWorld,
		IAMRoleDefinition{})

	cfResName, _ := lambdaFn.RequireCustomResource(IAMRoleDefinition{},
		userDefinedCustomResource,
		nil,
		nil)

	lambdaFn.Decorator = func(serviceName string,
		lambdaResourceName string,
		lambdaResource gocf.LambdaFunction,
		resourceMetadata map[string]interface{},
		S3Bucket string,
		S3Key string,
		buildID string,
		cfTemplate *gocf.Template,
		context map[string]interface{},
		logger *logrus.Logger) error {

		// Pass CustomResource outputs to the λ function
		resourceMetadata["CustomResource"] = gocf.GetAtt(cfResName, "CustomResourceResult")
		return nil
	}

	var lambdaFunctions []*LambdaAWSInfo
	lambdaFunctions = append(lambdaFunctions, lambdaFn)

	Main("SpartaUserCustomResource",
		"Uses a user-defined CloudFormation CustomResource",
		lambdaFunctions,
		nil,
		nil)
}
