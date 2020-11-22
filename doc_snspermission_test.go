package sparta

import (
	"context"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

const snsTopic = "arn:aws:sns:us-west-2:123412341234:mySNSTopic"

func snsProcessor(ctx context.Context,
	props map[string]interface{}) (map[string]interface{}, error) {
	lambdaCtx, _ := lambdacontext.FromContext(ctx)
	Logger().Info().
		Str("RequestID", lambdaCtx.AwsRequestID).
		Interface("Properties", props).
		Msg("Lambda event")
	return props, nil
}

func ExampleSNSPermission() {
	var lambdaFunctions []*LambdaAWSInfo

	snsLambda, _ := NewAWSLambda(LambdaName(snsProcessor),
		snsProcessor,
		IAMRoleDefinition{})
	snsLambda.Permissions = append(snsLambda.Permissions, SNSPermission{
		BasePermission: BasePermission{
			SourceArn: snsTopic,
		},
	})
	lambdaFunctions = append(lambdaFunctions, snsLambda)
	Main("SNSLambdaApp", "Registers for SNS events", lambdaFunctions, nil, nil)
}
