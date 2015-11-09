package sparta

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"net/http"
)

const S3_BUCKET = "arn:aws:sns:us-west-2:123412341234:myBucket"

func s3LambdaProcessor(event *json.RawMessage, context *LambdaContext, w *http.ResponseWriter, logger *logrus.Logger) {
	logger.WithFields(logrus.Fields{
		"RequestID": context.AWSRequestId,
	}).Info("S3Event")

	logger.Info("Event data: ", string(*event))
}

func ExampleS3Permission() {
	var lambdaFunctions []*LambdaAWSInfo
	// Define the IAM role
	roleDefinition := IAMRoleDefinition{}
	roleDefinition.Privileges = append(roleDefinition.Privileges, IAMRolePrivilege{
		Actions: []string{"s3:GetObject",
			"s3:PutObject"},
		Resource: S3_BUCKET,
	})
	// Create the Lambda
	s3Lambda := NewLambda(IAMRoleDefinition{}, s3LambdaProcessor, nil)

	// Add a Permission s.t. the Lambda function automatically registers for S3 events
	s3Lambda.Permissions = append(s3Lambda.Permissions, S3Permission{
		BasePermission: BasePermission{
			SourceArn: S3_BUCKET,
		},
		Events: []string{"s3:ObjectCreated:*", "s3:ObjectRemoved:*"},
	})

	lambdaFunctions = append(lambdaFunctions, s3Lambda)
	Main("S3LambdaApp", "Registers for S3 events", lambdaFunctions)
}
