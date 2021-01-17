package sparta

import (
	"context"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

const s3Bucket = "arn:aws:sns:us-west-2:123412341234:myBucket"

func s3LambdaProcessor(ctx context.Context,
	props map[string]interface{}) (map[string]interface{}, error) {
	lambdaCtx, _ := lambdacontext.FromContext(ctx)
	Logger().Info().
		Str("RequestID", lambdaCtx.AwsRequestID).
		Interface("Properties", props).
		Msg("Lambda event")
	return props, nil
}

func ExampleS3Permission() {
	var lambdaFunctions []*LambdaAWSInfo
	// Define the IAM role
	roleDefinition := IAMRoleDefinition{}
	roleDefinition.Privileges = append(roleDefinition.Privileges, IAMRolePrivilege{
		Actions: []string{"s3:GetObject",
			"s3:PutObject"},
		Resource: s3Bucket,
	})
	// Create the Lambda
	s3Lambda, _ := NewAWSLambda(LambdaName(s3LambdaProcessor),
		s3LambdaProcessor,
		IAMRoleDefinition{})

	// Add a Permission s.t. the Lambda function automatically registers for S3 events
	s3Lambda.Permissions = append(s3Lambda.Permissions, S3Permission{
		BasePermission: BasePermission{
			SourceArn: s3Bucket,
		},
		Events: []string{"s3:ObjectCreated:*", "s3:ObjectRemoved:*"},
	})

	lambdaFunctions = append(lambdaFunctions, s3Lambda)
	mainErr := Main("S3LambdaApp", "Registers for S3 events", lambdaFunctions, nil, nil)
	if mainErr != nil {
		panic("Failed to invoke sparta.Main: %s" + mainErr.Error())
	}
}
